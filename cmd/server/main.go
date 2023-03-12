package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
)

const (
	metricContextKey = "metricContextKey"
	metricResultKey  = "metricResultKey"
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

type config struct {
	ServerURL     string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func main() {
	conf, err := createConfig()
	if err != nil {
		logger.ErrorFormat("Fail to create config file: %v", err)
		panic(err)
	}
	logger.InfoFormat("Starting server with the following configuration:%v", conf)

	inMemoryStorage := storage.NewInMemoryStorage()
	fileStorage := storage.NewFileStorage(conf)
	storageStrategy := storage.NewStorageStrategy(conf, inMemoryStorage, fileStorage)
	htmlPageBuilder := html.NewSimplePageBuilder()
	router := initRouter(storageStrategy, htmlPageBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if conf.Restore {
		logger.Info("Restore metrics from backup")
		err = storageStrategy.RestoreFromBackup()
		if err != nil {
			logger.ErrorFormat("Fail to restore state from backup: %v", err)
		}
	}

	if !conf.SyncMode() {
		logger.Info("Start periodic backup serice")
		backgroundStore := worker.NewPeriodicWorker(func(ctx context.Context) error { return storageStrategy.CreateBackup() })
		go backgroundStore.StartWork(ctx, conf.StoreInterval)
	}

	logger.Info("Start listen " + conf.ServerURL)
	err = http.ListenAndServe(conf.ServerURL, router)
	if err != nil {
		logger.ErrorObj(err)
	}
}

func createConfig() (*config, error) {
	conf := &config{}
	flag.BoolVar(&conf.Restore, "r", true, "Restore metric values from the server backup file")
	flag.DurationVar(&conf.StoreInterval, "i", time.Second*300, "Store backup interval")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Server listen URL")
	flag.StringVar(&conf.StoreFile, "f", "/tmp/devops-metrics-db.json", "Backup storage file path")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
}

func initRouter(metricsStorage storage.MetricsStorage, htmlPageBuilder html.HTMLPageBuilder) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Compress(gzip.BestSpeed, compressContentTypes...))
	router.Route("/update", func(r chi.Router) {
		r.With(fillJSONContext, updateMetric(metricsStorage)).
			Post("/", successJSONResponse())
		r.With(fillCommonURLContext, fillGaugeURLContext, updateMetric(metricsStorage)).
			Post("/gauge/{metricName}/{metricValue}", successURLResponse())
		r.With(fillCommonURLContext, fillCounterURLContext, updateMetric(metricsStorage)).
			Post("/counter/{metricName}/{metricValue}", successURLResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unknown metric type", http.StatusNotImplemented)
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

func fillGaugeURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)
		strValue := chi.URLParam(r, "metricValue")
		value, err := parser.ToFloat64(strValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", strValue, err), http.StatusBadRequest)
			return
		}

		metricContext.MType = "gauge"
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
			http.Error(w, fmt.Sprintf("Value parsing fail %v: %v", strValue, err), http.StatusBadRequest)
			return
		}

		metricContext.MType = "counter"
		metricContext.Delta = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fillJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricContext := ensureMetricContext(r)

		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.ErrorFormat("Fail to create gzip reader: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		err := json.NewDecoder(reader).Decode(metricContext)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid json: %v", err), http.StatusBadRequest)
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

func updateMetric(storage storage.MetricsStorage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			metricContext, ok := ctx.Value(metricInfoContextKey{key: metricContextKey}).(*model.Metrics)
			if !ok {
				http.Error(w, "Metric info not found in context", http.StatusInternalServerError)
				return
			}

			metric, err := model.FromModelMetric(metricContext)
			if err != nil {
				if errors.Is(err, model.ErrUnknownMetricType) {
					http.Error(w, err.Error(), http.StatusNotImplemented)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}

			resultMetric, err := storage.AddMetricValue(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			newValue, err := model.ToModelMetric(resultMetric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			logger.InfoFormat("Updated metric: %v. newValue: %v", metricContext.ID, newValue)
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

			metric, err := storage.GetMetric(metricContext.MType, metricContext.ID)
			if err != nil {
				logger.ErrorFormat("Fail to get metric value: %v", err)
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}

			resultValue, err := model.ToModelMetric(metric)
			if err != nil {
				logger.ErrorFormat("Fail to get metric value: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
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

		metric, err := model.FromModelMetric(metricValueResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		successResponse(w, "text/plain", metric.GetStringValue())
	}
}

func handleMetricsPage(builder html.HTMLPageBuilder, storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := storage.GetMetricValues()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		successResponse(w, "text/html", builder.BuildMetricsPage(values))
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
			logger.ErrorFormat("Fail to serialise result: %v", err)
			http.Error(w, "Fail to serialise result", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.ErrorFormat("Fail to write response: %v", err)
		}
	}
}

func successResponse(w http.ResponseWriter, contentType string, message string) {
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		logger.ErrorFormat("Fail to write response: %v", err)
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

func (c *config) StoreFilePath() string {
	return c.StoreFile
}

func (c *config) SyncMode() bool {
	return c.StoreInterval == 0
}

func (c *config) String() string {
	return fmt.Sprintf("\nServerURL:\t%v\nStoreInterval:\t%v\nStoreFile:\t%v\nRestore:\t%v",
		c.ServerURL, c.StoreInterval, c.StoreFile, c.Restore)
}
