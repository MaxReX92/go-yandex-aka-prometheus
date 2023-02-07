package main

import (
	"go-yandex-aka-prometheus/internal/logger"
	"go-yandex-aka-prometheus/internal/storage"
	"net/http"
	"strings"
)

const (
	listenURL = "127.0.0.1:8080"
)

func main() {

	metricsStorage := storage.NewMetricsStorage()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	http.HandleFunc("/update/gauge/", func(w http.ResponseWriter, r *http.Request) { handleMetric(w, r, &metricsStorage) })
	http.HandleFunc("/update/counter/", func(w http.ResponseWriter, r *http.Request) { handleMetric(w, r, &metricsStorage) })
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(metricsStorage.GetMetrics())) })
	http.HandleFunc("/", http.NotFound)

	logger.Info("Start listen " + listenURL)
	err := http.ListenAndServe(listenURL, nil)
	if err != nil {
		logger.Error(err.Error())
	}
}

func handleMetric(w http.ResponseWriter, r *http.Request, storage *storage.MetricsStorage) {
	parts := strings.Split(r.RequestURI, "/")
	if len(parts) != 5 {
		notFound(w)
		return
	}

	metricName := parts[3]
	stringValue := parts[4]

	switch parts[2] {
	case "gauge":
		storage.AddGaugeMetricValue(metricName, stringValue)
	case "counter":
		storage.AddCounterMetricValue(metricName, stringValue)
	default:
		{
			notFound(w)
			return
		}
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
