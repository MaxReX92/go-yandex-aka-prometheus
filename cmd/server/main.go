package main

import (
	"fmt"
	"go-yandex-aka-prometheus/internal/logger"
	"go-yandex-aka-prometheus/internal/parser"
	"go-yandex-aka-prometheus/internal/storage"
	"net/http"
	"strings"
)

const (
	listenURL = "127.0.0.1:8080"
)

func main() {

	metricsStorage := storage.NewInMemoryStorage()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	http.HandleFunc("/update/gauge/", func(w http.ResponseWriter, r *http.Request) { handleMetric(w, r, metricsStorage) })
	http.HandleFunc("/update/counter/", func(w http.ResponseWriter, r *http.Request) { handleMetric(w, r, metricsStorage) })
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(metricsStorage.GetMetrics())) })
	http.HandleFunc("/", http.NotFound)

	logger.Info("Start listen " + listenURL)
	err := http.ListenAndServe(listenURL, nil)
	if err != nil {
		logger.Error(err.Error())
	}
}

func handleMetric(w http.ResponseWriter, r *http.Request, storage storage.MetricsStorage) {
	parts := strings.Split(r.RequestURI, "/")
	if len(parts) != 5 {
		notFound(w)
		return
	}

	metricName := parts[3]
	stringValue := parts[4]

	switch parts[2] {
	case "gauge":
		{
			value, err := parser.ToFloat64(stringValue)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("Value parsing fail %v: %v", stringValue, err.Error())))
				return
			}

			storage.AddGaugeMetricValue(metricName, value)
		}
	case "counter":
		{
			value, err := parser.ToInt64(stringValue)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("Value parsing fail %v: %v", stringValue, err.Error())))
				return
			}

			storage.AddCounterMetricValue(metricName, value)
		}

	default:
		{
			notFound(w)
			return
		}
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		logger.ErrorFormat("Fail to write response: %v", err.Error())
	}
	logger.InfoFormat("Updated metric: %v. value: %v", metricName, stringValue)
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(404)
	_, err := w.Write([]byte("404 page not found"))
	if err != nil {
		logger.ErrorFormat("Fail to write response: %v", err.Error())
		return
	}
}
