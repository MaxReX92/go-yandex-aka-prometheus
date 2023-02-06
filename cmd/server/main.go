package main

import (
	"log"
	"net/http"
)

const (
	listenUrl = "127.0.0.1:8080"
)

func main() {

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	http.HandleFunc("/update/", handleMetric)
	log.Fatal(http.ListenAndServe(listenUrl, nil))
}

func handleMetric(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
}
