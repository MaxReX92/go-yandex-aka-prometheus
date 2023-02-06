package main

import (
	"go-yandex-aka-prometheus/internal/logger"
	"net/http"
)

const (
	listenUrl = "127.0.0.1:8080"
)

func main() {

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	http.HandleFunc("/update/", handleMetric)

	logger.Info("Start listen " + listenUrl)
	err := http.ListenAndServe(listenUrl, nil)
	if err != nil {
		logger.Error(err.Error())
	}
}

func handleMetric(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.RequestURI)
}
