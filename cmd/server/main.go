package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-yandex-aka-prometheus/internal/html"
	"go-yandex-aka-prometheus/internal/logger"
	"go-yandex-aka-prometheus/internal/parser"
	"go-yandex-aka-prometheus/internal/storage"
	"net/http"
)

const (
	listenURL = "127.0.0.1:8080"
)

type metricInfo struct {
	metricName  string
	metricValue string
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

func initRouter(metricsStorage storage.MetricsStorage, htmlPageBuilder html.HtmlPageBuilder) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/update", func(r chi.Router) {
		r.Route("/gauge/{metricName}/{metricValue}", func(r chi.Router) {
			r.Use(fillMetricContext)
			r.Post("/", handleGaugeMetric(metricsStorage))
		})
		r.Route("/counter/{metricName}/{metricValue}", func(r chi.Router) {
			r.Use(fillMetricContext)
			r.Post("/", handleCounterMetric(metricsStorage))
		})
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
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

func fillMetricContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricContext := &metricInfo{
			metricName:  chi.URLParam(r, "metricName"),
			metricValue: chi.URLParam(r, "metricValue"),
		}

		ctx := context.WithValue(r.Context(), "metricInfo", metricContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleGaugeMetric(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value("metricInfo").(*metricInfo)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		value, err := parser.ToFloat64(metricContext.metricValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", metricContext.metricValue, err.Error()), http.StatusBadRequest)
			return
		}

		storage.AddGaugeMetricValue(metricContext.metricName, value)

		successResponse(w, "text/plain", "ok")
		logger.InfoFormat("Updated metric: %v. value: %v", metricContext.metricName, metricContext.metricValue)
	}
}

func handleCounterMetric(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		metricContext, ok := ctx.Value("metricInfo").(*metricInfo)
		if !ok {
			http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
			return
		}

		value, err := parser.ToInt64(metricContext.metricValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", metricContext.metricValue, err.Error()), http.StatusBadRequest)
			return
		}

		storage.AddCounterMetricValue(metricContext.metricName, value)

		successResponse(w, "text/plain", "ok")
		logger.InfoFormat("Updated metric: %v. value: %v", metricContext.metricName, metricContext.metricValue)
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
