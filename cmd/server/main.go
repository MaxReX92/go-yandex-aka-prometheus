package main

import (
	"context"
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
	listenURL     = "127.0.0.1:8080"
	metricInfoKey = "metricInfo"
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
		r.Route("/gauge/{metricName}/{metricValue}", func(r chi.Router) {
			r.Use(fillCommonUrlContext)
			r.With(fillGaugeContext).
				Post("/", updateGaugeMetric(metricsStorage))
		})
		r.Route("/counter/{metricName}/{metricValue}", func(r chi.Router) {
			r.Use(fillCommonUrlContext)
			r.With(fillCounterUrlContext).
				Post("/", updateCounterMetric(metricsStorage))
		})
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
		})
	})
	router.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}/{metricName}", func(r chi.Router) {
			r.Use(fillCommonUrlContext)
			r.Get("/", handleMetricValue(metricsStorage))
		})
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			successResponse(w, "text/html", htmlPageBuilder.BuildMetricsPage(metricsStorage.GetMetricValues()))
		})
		r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
			successResponse(w, "text/html", htmlPageBuilder.BuildMetricsPage(metricsStorage.GetMetricValues()))
		})
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

func updateGaugeMetric(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value(metricInfoContextKey{key: metricInfoKey}).(*model.Metrics)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		storage.AddGaugeMetricValue(metricContext.ID, *metricContext.Value)
		successResponse(w, "text/plain", "ok")
		logger.InfoFormat("Updated metric: %v. value: %v", metricContext.ID, *metricContext.Value)
	}
}

func updateCounterMetric(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value(metricInfoContextKey{key: metricInfoKey}).(*model.Metrics)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		storage.AddCounterMetricValue(metricContext.ID, *metricContext.Delta)

		successResponse(w, "text/plain", "ok")
		logger.InfoFormat("Updated metric: %v. value: %v", metricContext.ID, *metricContext.Delta)
	}
}

func handleMetricValue(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value(metricInfoContextKey{key: metricInfoKey}).(*model.Metrics)
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
	metricContext, ok := ctx.Value(metricInfoContextKey{key: metricInfoKey}).(*model.Metrics)
	if !ok {
		metricContext = &model.Metrics{}
		ctx = context.WithValue(r.Context(), metricInfoContextKey{key: metricInfoKey}, metricContext)
	}

	return ctx, metricContext
}
