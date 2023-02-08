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
	http.HandleFunc("/update/", handleMetric(metricsStorage))
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, http.StatusOK, metricsStorage.GetMetrics())
	})
	http.HandleFunc("/", http.NotFound)

	logger.Info("Start listen " + listenURL)
	err := http.ListenAndServe(listenURL, nil)
	if err != nil {
		logger.Error(err.Error())
	}
}

func handleMetric(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		parts := strings.Split(r.RequestURI, "/")
		if len(parts) != 5 {
			writeResponse(w, http.StatusNotFound, "404 page not found")
			return
		}

		metricName := parts[3]
		stringValue := parts[4]

		switch parts[2] {
		case "gauge":
			{
				value, err := parser.ToFloat64(stringValue)
				if err != nil {
					writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v", stringValue, err.Error()))
					return
				}

				storage.AddGaugeMetricValue(metricName, value)
			}
		case "counter":
			{
				value, err := parser.ToInt64(stringValue)
				if err != nil {
					writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v", stringValue, err.Error()))
					return
				}

				storage.AddCounterMetricValue(metricName, value)
			}

		default:
			{
				writeResponse(w, http.StatusNotImplemented, fmt.Sprintf("Unknown metric type: %v", parts[2]))
				return
			}
		}

		writeResponse(w, http.StatusOK, "ok")
		logger.InfoFormat("Updated metric: %v. value: %v", metricName, stringValue)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(message))
	if err != nil {
		logger.ErrorFormat("Fail to write response: %v", err.Error())
	}
}
