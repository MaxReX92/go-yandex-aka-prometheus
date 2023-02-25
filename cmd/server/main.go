package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/storage"
)

const (
	listenURL        = "127.0.0.1:8080"
	metricContextKey = "metricContextKey"
	metricResultKey  = "metricResultKey"
)

type metricInfoContextKey struct {
	key string
}

func main() {
	metricsStorage := storage.NewInMemoryStorage()
	htmlPageBuilder := html.NewSimplePageBuilder()
	router := initRouter(metricsStorage, htmlPageBuilder)

	logger.Info("Start listen " + listenURL)
	err := http.ListenAndServe(listenURL, router)
	if err != nil {
		logger.Error(err.Error())
	}
}

func initRouter(metricsStorage storage.MetricsStorage, htmlPageBuilder html.HTMLPageBuilder) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/update", func(r chi.Router) {
		r.With(fillJsonContext, updateTypedMetric(metricsStorage)).
			Post("/", successJsonResponse())
		r.With(fillCommonUrlContext, fillGaugeContext, updateGaugeMetric(metricsStorage)).
			Post("/gauge/{metricName}/{metricValue}", successUrlResponse())
		r.With(fillCommonUrlContext, fillCounterUrlContext, updateCounterMetric(metricsStorage)).
			Post("/counter/{metricName}/{metricValue}", successUrlResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
		})
	})

	router.Route("/value", func(r chi.Router) {
		r.With(fillCommonUrlContext).
			Get("/{metricType}/{metricName}", handleMetricValue(metricsStorage))

		// handle json request here
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", handleMetricsPage(htmlPageBuilder, metricsStorage))
		r.Get("/metrics", handleMetricsPage(htmlPageBuilder, metricsStorage))
	})

	return router
}

func fillCommonUrlContext(next http.Handler) http.Handler {
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

func fillCounterUrlContext(next http.Handler) http.Handler {
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

func fillJsonContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		err := decoder.Decode(metricContext)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid json: %v", err.Error()), http.StatusBadRequest)
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
			newValue := storage.AddGaugeMetricValue(metricContext.ID, *metricContext.Value)
			result.Value = &newValue
		case "counter":
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

func handleMetricValue(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value(metricInfoContextKey{key: metricContextKey}).(*model.Metrics)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		value, ok := storage.GetMetricValue(metricContext.MType, metricContext.ID)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		successResponse(w, "text/plain", value)
	}
}

func handleMetricsPage(builder html.HTMLPageBuilder, storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/html", builder.BuildMetricsPage(storage.GetMetricValues()))
	}
}

func successUrlResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/plain", "ok")
	}
}

func successJsonResponse() func(w http.ResponseWriter, r *http.Request) {
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
