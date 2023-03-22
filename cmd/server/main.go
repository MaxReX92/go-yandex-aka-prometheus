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

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database/postgres"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database/stub"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/db"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/file"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/memory"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
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
	requestMetrics []*model.Metrics
	resultMetrics  []*model.Metrics
}

type config struct {
	Key           string        `env:"KEY"`
	ServerURL     string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	DB            string        `env:"DATABASE_DSN"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("create config file", err))
	}
	logger.InfoFormat("Starting server with the following configuration:%v", conf)

	var base database.DataBase
	var backupStorage storage.MetricsStorage
	if conf.DB == "" {
		base = &stub.StubDataBase{}
		backupStorage = file.NewFileStorage(conf)
	} else {
		base, err = postgres.NewPostgresDataBase(ctx, conf)
		if err != nil {
			panic(logger.WrapError("create database", err))
		}

		backupStorage = db.NewDBStorage(base)
	}
	defer base.Close()

	inMemoryStorage := memory.NewInMemoryStorage()
	storageStrategy := storage.NewStorageStrategy(conf, inMemoryStorage, backupStorage)
	defer storageStrategy.Close()

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	htmlPageBuilder := html.NewSimplePageBuilder()
	router := initRouter(storageStrategy, converter, htmlPageBuilder, base)

	if conf.Restore {
		logger.Info("Restore metrics from backup")
		err = storageStrategy.RestoreFromBackup(ctx)
		if err != nil {
			logger.ErrorFormat("failed to restore state from backup: %v", err)
		}
	}

	if !conf.SyncMode() {
		logger.Info("Start periodic backup serice")
		backgroundStore := worker.NewPeriodicWorker(func(ctx context.Context) error { return storageStrategy.CreateBackup(ctx) })
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

	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.BoolVar(&conf.Restore, "r", true, "Restore metric values from the server backup file")
	flag.DurationVar(&conf.StoreInterval, "i", time.Second*300, "Store backup interval")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Server listen URL")
	flag.StringVar(&conf.StoreFile, "f", "/tmp/devops-metrics-dataBase.json", "Backup storage file path")
	flag.StringVar(&conf.DB, "d", "", "Database connection stirng")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
}

func initRouter(metricsStorage storage.MetricsStorage, converter *model.MetricsConverter, htmlPageBuilder html.PageBuilder, dbStorage database.DataBase) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Compress(gzip.BestSpeed, compressContentTypes...))
	router.Route("/update", func(r chi.Router) {
		r.With(fillSingleJSONContext, updateMetrics(metricsStorage, converter)).
			Post("/", successSingleJSONResponse())
		r.With(fillCommonURLContext, fillGaugeURLContext, updateMetrics(metricsStorage, converter)).
			Post("/gauge/{metricName}/{metricValue}", successURLResponse())
		r.With(fillCommonURLContext, fillCounterURLContext, updateMetrics(metricsStorage, converter)).
			Post("/counter/{metricName}/{metricValue}", successURLResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintf("unknown metric type: %s", chi.URLParam(r, "metricType"))
			logger.Error("failed to update metric: " + message)
			http.Error(w, message, http.StatusNotImplemented)
		})
	})

	router.Route("/updates", func(r chi.Router) {
		r.With(fillMultiJSONContext, updateMetrics(metricsStorage, converter)).
			Post("/", successMultiJSONResponse())
	})

	router.Route("/value", func(r chi.Router) {
		r.With(fillSingleJSONContext, fillMetricValues(metricsStorage, converter)).
			Post("/", successSingleJSONResponse())

		r.With(fillCommonURLContext, fillMetricValues(metricsStorage, converter)).
			Get("/{metricType}/{metricName}", successURLValueResponse(converter))
	})

	router.Route("/ping", func(r chi.Router) {
		r.Get("/", handleDBPing(dbStorage))
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", handleMetricsPage(htmlPageBuilder, metricsStorage))
		r.Get("/metrics", handleMetricsPage(htmlPageBuilder, metricsStorage))
	})

	return router
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
		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, logger.WrapError("create gzip reader", err).Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		metricContext := &model.Metrics{}
		metricsContext.requestMetrics = append(metricsContext.requestMetrics, metricContext)

		err := json.NewDecoder(reader).Decode(metricContext)
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
		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, logger.WrapError("create gzip reader", err).Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		metricsContext.requestMetrics = []*model.Metrics{}
		err := json.NewDecoder(reader).Decode(&metricsContext.requestMetrics)
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

func updateMetrics(storage storage.MetricsStorage, converter *model.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsList := make([]metrics.Metric, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := converter.FromModelMetric(metricContext)
				if err != nil {
					logger.ErrorFormat("Fail to parse metric: %v", err)

					var errUnknownMetricType *model.UnknownMetricTypeError
					if errors.As(err, &errUnknownMetricType) {
						http.Error(w, fmt.Sprintf("unknown metric type: %s", errUnknownMetricType.UnknownType), http.StatusNotImplemented)
					} else {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
					return
				}

				metricsList[i] = metric
			}

			resultMetrics, err := storage.AddMetricValues(ctx, metricsList)
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

func fillMetricValues(storage storage.MetricsStorage, converter *model.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsContext.resultMetrics = make([]*model.Metrics, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := storage.GetMetric(ctx, metricContext.MType, metricContext.ID)
				if err != nil {
					logger.ErrorFormat("Fail to get metric value: %v", err)
					http.Error(w, "Metric not found", http.StatusNotFound)
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

func successURLValueResponse(converter *model.MetricsConverter) func(w http.ResponseWriter, r *http.Request) {
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

func handleMetricsPage(builder html.PageBuilder, storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := storage.GetMetricValues(r.Context())
		if err != nil {
			http.Error(w, logger.WrapError("get metric values", err).Error(), http.StatusInternalServerError)
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
		// _, metricsContext := ensureMetricsContext(r)
		// metricsContext.resultMetrics <-- update results

		// just a stub, maybe temporary
		stubResult := &model.Metrics{}
		result, err := json.Marshal(stubResult)
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

func handleDBPing(dbStorage database.DataBase) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := dbStorage.Ping(r.Context())
		if err == nil {
			successResponse(w, "text/plain", "ok")
		} else {
			http.Error(w, logger.WrapError("ping database", err).Error(), http.StatusInternalServerError)
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

func (c *config) StoreFilePath() string {
	return c.StoreFile
}

func (c *config) SyncMode() bool {
	return c.DB != "" || c.StoreInterval == 0
}

func (c *config) String() string {
	return fmt.Sprintf("\nServerURL:\t%v\nStoreInterval:\t%v\nStoreFile:\t%v\nRestore:\t%v\nDb:\t%v",
		c.ServerURL, c.StoreInterval, c.StoreFile, c.Restore, c.DB)
}

func (c *config) GetKey() []byte {
	return []byte(c.Key)
}

func (c *config) SignMetrics() bool {
	return c.Key != ""
}

func (c *config) GetConnectionString() string {
	return c.DB
}
