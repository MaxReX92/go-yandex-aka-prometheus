package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/caarlos0/env/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/storage"
)

const (
	metricContextKey = "metricContextKey"
	metricResultKey  = "metricResultKey"
)

type metricInfoContextKey struct {
	key string
}

type config struct {
	ServerURL            string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreIntervalSeconds int64  `env:"STORE_INTERVAL" envDefault:"300"`
	StoreFile            string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore              bool   `env:"RESTORE" envDefault:"true"`
}

func main() {
	conf, err := createConfig()
	if err != nil {
		panic(err)
	}

	metricsStorage := storage.NewInMemoryStorage()
	htmlPageBuilder := html.NewSimplePageBuilder()
	router := initRouter(metricsStorage, htmlPageBuilder)

	logger.Info("Start listen " + conf.ServerURL)
	err = http.ListenAndServe(conf.ServerURL, router)
	if err != nil {
		logger.Error(err.Error())
	}
}

func createConfig() (*config, error) {
	conf := &config{}
	err := env.Parse(conf)
	return conf, err
}

func initRouter(metricsStorage storage.MetricsStorage, htmlPageBuilder html.HTMLPageBuilder) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/update", func(r chi.Router) {
		r.With(fillJSONContext, updateTypedMetric(metricsStorage)).
			Post("/", successJSONResponse())
		r.With(fillCommonURLContext, fillGaugeContext, updateGaugeMetric(metricsStorage)).
			Post("/gauge/{metricName}/{metricValue}", successURLResponse())
		r.With(fillCommonURLContext, fillCounterURLContext, updateCounterMetric(metricsStorage)).
			Post("/counter/{metricName}/{metricValue}", successURLResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
		})
	})

	router.Route("/value", func(r chi.Router) {
		r.With(fillJSONContext, fillMetricValue(metricsStorage)).
			Post("/", successJSONResponse())

		r.With(fillCommonURLContext, fillMetricValue(metricsStorage)).
			Get("/{metricType}/{metricName}", successURLValueResponse())
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", handleMetricsPage(htmlPageBuilder, metricsStorage))
		r.Get("/metrics", handleMetricsPage(htmlPageBuilder, metricsStorage))
	})

	return router
}

func fillCommonURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)

		metricContext.ID = chi.URLParam(r, "metricName")
		metricContext.MType = chi.URLParam(r, "metricType")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillGaugeContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)
		strValue := chi.URLParam(r, "metricValue")
		value, err := parser.ToFloat64(strValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", strValue, err.Error()), http.StatusBadRequest)
			return
		}

		metricContext.Value = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillCounterURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)
		strValue := chi.URLParam(r, "metricValue")
		value, err := parser.ToInt64(strValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", strValue, err.Error()), http.StatusBadRequest)
			return
		}

		metricContext.Delta = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		err := decoder.Decode(metricContext)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid json: %v", err.Error()), http.StatusBadRequest)
			return
		}

		if metricContext.ID == "" {
			http.Error(w, "metric name is missed", http.StatusBadRequest)
			return
		}

		if metricContext.MType == "" {
			http.Error(w, "metric type is missed", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func updateTypedMetric(storage storage.MetricsStorage) func(next http.Handler) http.Handler {
	return updateMetric(func(w http.ResponseWriter, metricContext *model.Metrics) (*model.Metrics, bool) {
		result := &model.Metrics{
			ID:    metricContext.ID,
			MType: metricContext.MType,
		}

		switch metricContext.MType {
		case "gauge":
			if metricContext.Value == nil {
				http.Error(w, "metric value is missed", http.StatusBadRequest)
				return nil, false
			}
			newValue := storage.AddGaugeMetricValue(metricContext.ID, *metricContext.Value)
			result.Value = &newValue
		case "counter":
			if metricContext.Delta == nil {
				http.Error(w, "metric value is missed", http.StatusBadRequest)
				return nil, false
			}
			newValue := storage.AddCounterMetricValue(metricContext.ID, *metricContext.Delta)
			result.Delta = &newValue
		default:
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
			return nil, false
		}

		return result, true
	})
}

func updateGaugeMetric(storage storage.MetricsStorage) func(next http.Handler) http.Handler {
	return updateMetric(func(w http.ResponseWriter, metricContext *model.Metrics) (*model.Metrics, bool) {
		res := storage.AddGaugeMetricValue(metricContext.ID, *metricContext.Value)
		return &model.Metrics{
			ID:    metricContext.ID,
			MType: metricContext.MType,
			Value: &res,
		}, true
	})
}

func updateCounterMetric(storage storage.MetricsStorage) func(next http.Handler) http.Handler {
	return updateMetric(func(w http.ResponseWriter, metricContext *model.Metrics) (*model.Metrics, bool) {
		res := storage.AddCounterMetricValue(metricContext.ID, *metricContext.Delta)
		return &model.Metrics{
			ID:    metricContext.ID,
			MType: metricContext.MType,
			Delta: &res,
		}, true
	})
}

func updateMetric(updateAction func(w http.ResponseWriter, metrics *model.Metrics) (*model.Metrics, bool)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			metricContext, ok := ctx.Value(metricInfoContextKey{key: metricContextKey}).(*model.Metrics)
			if !ok {
				http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
				return
			}

			newValue, ok := updateAction(w, metricContext)
			if !ok {
				return
			}

			logger.InfoFormat("Updated metric: %v. newValue: %v", metricContext.ID, *newValue)
			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, metricInfoContextKey{key: metricResultKey}, newValue)))
		})
	}
}

func fillMetricValue(storage storage.MetricsStorage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			metricContext, ok := ctx.Value(metricInfoContextKey{key: metricContextKey}).(*model.Metrics)
			if !ok {
				http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
				return
			}

			metricValue, ok := storage.GetMetricValue(metricContext.MType, metricContext.ID)
			if !ok {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}

			resultValue := &model.Metrics{
				ID:    metricContext.ID,
				MType: metricContext.MType,
			}

			switch metricContext.MType {
			case "counter":
				counterValue := int64(metricValue)
				resultValue.Delta = &counterValue
			case "gauge":
				resultValue.Value = &metricValue
			default:
				http.Error(w, "Unknown metric type", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, metricInfoContextKey{key: metricResultKey}, resultValue)))
		})
	}
}

func successURLValueResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricValueResult, ok := ctx.Value(metricInfoContextKey{key: metricResultKey}).(*model.Metrics)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		var result string
		switch metricValueResult.MType {
		case "counter":
			result = parser.IntToString(*metricValueResult.Delta)
		case "gauge":
			result = parser.FloatToString(*metricValueResult.Value)
		default:
			http.Error(w, "Unknown metric type", http.StatusInternalServerError)
			return
		}

		successResponse(w, "text/plain", result)
	}
}

func handleMetricsPage(builder html.HTMLPageBuilder, storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/html", builder.BuildMetricsPage(storage.GetMetricValues()))
	}
}

func successURLResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/plain", "ok")
	}
}

func successJSONResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricUpdateResult, ok := ctx.Value(metricInfoContextKey{key: metricResultKey}).(*model.Metrics)
		if !ok {
			http.Error(w, "Metric update result not found in context", http.StatusInternalServerError)
			return
		}

		result, err := json.Marshal(metricUpdateResult)
		if err != nil {
			logger.ErrorFormat("Fail to serialise result: %v", err.Error())
			http.Error(w, "Fail to serialise result", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.ErrorFormat("Fail to write response: %v", err.Error())
		}
	}
}

func successResponse(w http.ResponseWriter, contentType string, message string) {
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		logger.ErrorFormat("Fail to write response: %v", err.Error())
	}
}

func ensureMetricContext(r *http.Request) (context.Context, *model.Metrics) {
	ctx := r.Context()
	metricContext, ok := ctx.Value(metricInfoContextKey{key: metricContextKey}).(*model.Metrics)
	if !ok {
		metricContext = &model.Metrics{}
		ctx = context.WithValue(r.Context(), metricInfoContextKey{key: metricContextKey}, metricContext)
	}

	return ctx, metricContext
}
