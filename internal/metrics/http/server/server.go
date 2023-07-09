package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	metricsHttp "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/http"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

const (
	counterMetricName = "counter"
	gaugeMetricName   = "gauge"
)

var compressContentTypes = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

type metricInfoContextKey struct {
	key string
}

type metricsRequestContext struct {
	body           []byte
	requestMetrics []*model.Metrics
	resultMetrics  []*model.Metrics
}

type ServerConfig interface {
	ListenURL() string
	ClientsTrustedSubnet() *net.IPNet
}

type httpServer struct {
	srv *http.Server
}

func New(conf ServerConfig,
	converter *metricsHttp.MetricsConverter,
	decryptor crypto.Decryptor,
	requestHandler server.RequestHandler,
) *httpServer {
	return &httpServer{
		srv: &http.Server{
			Addr:    conf.ListenURL(),
			Handler: createRouter(converter, decryptor, conf.ClientsTrustedSubnet(), requestHandler),
		},
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	logger.Info("Start web service")
	return s.srv.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	logger.Info("Stopping web service")
	return s.srv.Shutdown(ctx)
}

func createRouter(
	converter *metricsHttp.MetricsConverter,
	decryptor crypto.Decryptor,
	clientSubnet *net.IPNet,
	requestHandler server.RequestHandler,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	if clientSubnet != nil {
		router.Use(middleware.RealIP)
		router.Use(checkClientSubnet(clientSubnet))
	}
	router.Use(middleware.Compress(gzip.BestSpeed, compressContentTypes...))
	router.Route("/update", func(r chi.Router) {
		r.With(decrypt(decryptor), fillSingleJSONContext, updateMetrics(requestHandler, converter)).
			Post("/", successSingleJSONResponse())
		r.With(fillCommonURLContext, fillGaugeURLContext, updateMetrics(requestHandler, converter)).
			Post("/gauge/{metricName}/{metricValue}", successURLResponse())
		r.With(fillCommonURLContext, fillCounterURLContext, updateMetrics(requestHandler, converter)).
			Post("/counter/{metricName}/{metricValue}", successURLResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintf("unknown metric type: %s", chi.URLParam(r, "metricType"))
			logger.Error("failed to update metric: " + message)
			http.Error(w, message, http.StatusNotImplemented)
		})
	})

	router.Route("/updates", func(r chi.Router) {
		r.With(decrypt(decryptor), fillMultiJSONContext, updateMetrics(requestHandler, converter)).
			Post("/", successMultiJSONResponse())
	})

	router.Route("/value", func(r chi.Router) {
		r.With(decrypt(decryptor), fillSingleJSONContext, fillMetricValues(requestHandler, converter)).
			Post("/", successSingleJSONResponse())

		r.With(fillCommonURLContext, fillMetricValues(requestHandler, converter)).
			Get("/{metricType}/{metricName}", successURLValueResponse(converter))
	})

	router.Route("/ping", func(r chi.Router) {
		r.Get("/", handleDBPing(requestHandler))
	})

	router.Route("/debug", func(r chi.Router) {
		r.Handle("/*", http.DefaultServeMux)
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", handleMetricsPage(requestHandler))
		r.Get("/metrics", handleMetricsPage(requestHandler))
	})

	return router
}

func decrypt(decryptor crypto.Decryptor) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reader io.Reader
			if r.Header.Get(`Content-Encoding`) == `gzip` {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, logger.WrapError("create gzip reader", err).Error(), http.StatusInternalServerError)
					return
				}
				reader = gz
			} else {
				reader = r.Body
			}

			defer func() {
				closer, ok := reader.(io.Closer)
				if ok {
					closer.Close()
				}
			}()

			body, err := io.ReadAll(reader)
			if err != nil {
				http.Error(w, logger.WrapError("read body data", err).Error(), http.StatusInternalServerError)
				return
			}

			if decryptor != nil {
				body, err = decryptor.Decrypt(body)
				if err != nil {
					http.Error(w, logger.WrapError("decrypt body data", err).Error(), http.StatusBadRequest)
					return
				}
			}

			ctx, metricsContext := ensureMetricsContext(r)
			metricsContext.body = body
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func fillCommonURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		metricsContext.requestMetrics = append(metricsContext.requestMetrics, &model.Metrics{
			ID:    chi.URLParam(r, "metricName"),
			MType: chi.URLParam(r, "metricType"),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillGaugeURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if len(metricsContext.requestMetrics) != 1 {
			logger.Error("fillGaugeURLContext: wrong context")
			http.Error(w, "fillGaugeURLContext: wrong context", http.StatusInternalServerError)
			return
		}

		strValue := chi.URLParam(r, "metricValue")
		value, err := parser.ToFloat64(strValue)
		if err != nil {
			http.Error(w, logger.WrapError(fmt.Sprintf("parse value: %v", strValue), err).Error(), http.StatusBadRequest)
			return
		}

		metricsContext.requestMetrics[0].MType = gaugeMetricName
		metricsContext.requestMetrics[0].Value = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillCounterURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if len(metricsContext.requestMetrics) != 1 {
			logger.Error("fillCounterURLContext: wrong context")
			http.Error(w, "fillCounterURLContext: wrong context", http.StatusInternalServerError)
			return
		}

		strValue := chi.URLParam(r, "metricValue")
		value, err := parser.ToInt64(strValue)
		if err != nil {
			http.Error(w, logger.WrapError(fmt.Sprintf("parse value: %v", strValue), err).Error(), http.StatusBadRequest)
			return
		}

		metricsContext.requestMetrics[0].MType = counterMetricName
		metricsContext.requestMetrics[0].Delta = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillSingleJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if metricsContext.body == nil {
			logger.Error("fillSingleJSONContext: wrong context")
			http.Error(w, "fillSingleJSONContext: wrong context", http.StatusInternalServerError)
			return
		}

		metricContext := &model.Metrics{}
		metricsContext.requestMetrics = append(metricsContext.requestMetrics, metricContext)

		err := json.Unmarshal(metricsContext.body, metricContext)
		if err != nil {
			http.Error(w, logger.WrapError("unmarhsal json context", err).Error(), http.StatusBadRequest)
			return
		}

		if metricContext.ID == "" {
			logger.Error("Fail to collect json context: metric name is missed")
			http.Error(w, "metric name is missed", http.StatusBadRequest)
			return
		}

		if metricContext.MType == "" {
			logger.Error("Fail to collect json context: metric type is missed")
			http.Error(w, "metric types is missed", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillMultiJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if metricsContext.body == nil {
			logger.Error("fillMultiJSONContext: wrong context")
			http.Error(w, "fillMultiJSONContext: wrong context", http.StatusInternalServerError)
			return
		}

		metricsContext.requestMetrics = []*model.Metrics{}
		err := json.Unmarshal(metricsContext.body, &metricsContext.requestMetrics)
		if err != nil {
			http.Error(w, logger.WrapError("unmarshal request metrics", err).Error(), http.StatusBadRequest)
			return
		}

		for _, requestMetric := range metricsContext.requestMetrics {
			if requestMetric.ID == "" {
				logger.Error("Fail to collect json context: metric name is missed")
				http.Error(w, "metric name is missed", http.StatusBadRequest)
				return
			}

			if requestMetric.MType == "" {
				logger.Error("Fail to collect json context: metric type is missed")
				http.Error(w, "metric types is missed", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func updateMetrics(requestHandler server.RequestHandler, converter *metricsHttp.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsList := make([]metrics.Metric, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := converter.FromModelMetric(metricContext)
				if err != nil {
					logger.ErrorFormat("Fail to parse metric: %v", err)

					if errors.Is(err, metrics.ErrUnknownMetricType) {
						http.Error(w, fmt.Sprintf("unknown metric type: %s", metricContext.MType), http.StatusNotImplemented)
					} else {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
					return
				}

				metricsList[i] = metric
			}

			resultMetrics, err := requestHandler.UpdateMetricValues(ctx, metricsList)
			if err != nil {
				http.Error(w, logger.WrapError("update metric", err).Error(), http.StatusInternalServerError)
				return
			}

			metricsContext.resultMetrics = make([]*model.Metrics, len(resultMetrics))
			for i, resultMetric := range resultMetrics {
				newValue, err := converter.ToModelMetric(resultMetric)
				if err != nil {
					http.Error(w, logger.WrapError("convert metric", err).Error(), http.StatusInternalServerError)
					return
				}

				logger.InfoFormat("Updated metric: %v. newValue: %v", resultMetric.GetName(), newValue)
				metricsContext.resultMetrics[i] = newValue
			}

			next.ServeHTTP(w, r)
		})
	}
}

func fillMetricValues(requestHandler server.RequestHandler, converter *metricsHttp.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsContext.resultMetrics = make([]*model.Metrics, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := requestHandler.GetMetricValue(ctx, metricContext.MType, metricContext.ID)
				if err != nil {
					var status int
					if errors.Is(err, server.ErrMetricNotFound) {
						status = http.StatusNotFound
					} else {
						status = http.StatusInternalServerError
					}

					http.Error(w, logger.WrapError("get metric value", err).Error(), status)
					return
				}

				resultValue, err := converter.ToModelMetric(metric)
				if err != nil {
					http.Error(w, logger.WrapError("get metric value", err).Error(), http.StatusInternalServerError)
					return
				}

				metricsContext.resultMetrics[i] = resultValue
			}

			next.ServeHTTP(w, r)
		})
	}
}

func successURLValueResponse(converter *metricsHttp.MetricsConverter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, metricsContext := ensureMetricsContext(r)

		if len(metricsContext.resultMetrics) != 1 {
			logger.Error("successURLValueResponse: wrong context")
			http.Error(w, "successURLValueResponse: wrong context", http.StatusInternalServerError)
			return
		}

		metric, err := converter.FromModelMetric(metricsContext.resultMetrics[0])
		if err != nil {
			http.Error(w, logger.WrapError("convert result metric", err).Error(), http.StatusInternalServerError)
			return
		}

		successResponse(w, "text/plain", metric.GetStringValue())
	}
}

func handleMetricsPage(requestHandler server.RequestHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := requestHandler.GetReportPage(r.Context())
		if err != nil {
			http.Error(w, logger.WrapError("get metric values", err).Error(), http.StatusInternalServerError)
			return
		}
		successResponse(w, "text/html", page)
	}
}

func successURLResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/plain", "ok")
	}
}

func successSingleJSONResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, metricsContext := ensureMetricsContext(r)

		if len(metricsContext.resultMetrics) != 1 {
			logger.Error("successSingleJSONResponse: wrong context")
			http.Error(w, "successSingleJSONResponse: wrong context", http.StatusInternalServerError)
			return
		}

		result, err := json.Marshal(metricsContext.resultMetrics[0])
		if err != nil {
			http.Error(w, logger.WrapError("serialise result", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.ErrorFormat("failed to write response: %v", err)
		}
	}
}

func successMultiJSONResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, metricsContext := ensureMetricsContext(r)
		result, err := json.Marshal(metricsContext.resultMetrics)
		if err != nil {
			http.Error(w, logger.WrapError("serialise result", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.ErrorFormat("failed to write response: %v", err)
		}
	}
}

func successResponse(w http.ResponseWriter, contentType string, message string) {
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		logger.ErrorFormat("failed to write response: %v", err)
	}
}

func handleDBPing(requestHandler server.RequestHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := requestHandler.Ping(r.Context())
		if err == nil {
			successResponse(w, "text/plain", "ok")
		} else {
			http.Error(w, logger.WrapError("ping", err).Error(), http.StatusInternalServerError)
		}
	}
}

func ensureMetricsContext(r *http.Request) (context.Context, *metricsRequestContext) {
	const metricsContextKey = "metricsContextKey"
	ctx := r.Context()
	metricsContext, ok := ctx.Value(metricInfoContextKey{key: metricsContextKey}).(*metricsRequestContext)
	if !ok {
		metricsContext = &metricsRequestContext{}
		ctx = context.WithValue(r.Context(), metricInfoContextKey{key: metricsContextKey}, metricsContext)
	}

	return ctx, metricsContext
}

func checkClientSubnet(clientSubnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := net.ParseIP(r.RemoteAddr)
			if clientIP == nil {
				logger.Error("failed to receive client ip")
				http.Error(w, "failed to receive client ip", http.StatusForbidden)

				return
			}

			if !clientSubnet.Contains(clientIP) {
				logger.Error("client net not trusted")
				http.Error(w, "client net not trusted", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
