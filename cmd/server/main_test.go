package main

import (
	"fmt"
	"go-yandex-aka-prometheus/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type expectedResult struct {
	status   int
	response string
}

func Test_UpdateRequest(t *testing.T) {
	type testDesctiption struct {
		testName    string
		httpMethod  string
		metricType  string
		metricName  string
		metricValue string
		expected    expectedResult
	}

	tests := []testDesctiption{}
	for _, method := range getMethods() {
		for _, metricType := range getMetricType() {
			for _, metricName := range getMetricName() {
				for _, metricValue := range getMetricValue() {

					var expected *expectedResult

					// Unexpected method type
					if method != http.MethodPost {
						expected = getExpected(http.StatusMethodNotAllowed, "Method not allowed")
					}

					// Unexpected metric type
					if expected == nil && metricType != "gauge" && metricType != "counter" {
						if metricType == "" || metricName == "" || metricValue == "" {
							expected = getExpectedNotFound()
						} else {
							expected = getExpected(http.StatusNotImplemented, "Unknown metric type: "+metricType)
						}
					}

					// Empty metric name
					if expected == nil && metricName == "" {
						expected = getExpectedNotFound()
					}

					// Incorrect metric value
					if expected == nil {
						if metricValue == "" {
							expected = getExpectedNotFound()
						} else if metricType == "gauge" {
							_, err := strconv.ParseFloat(metricValue, 64)
							if err != nil {
								expected = getExpected(http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v", metricValue, err.Error()))
							}
						} else if metricType == "counter" {
							_, err := strconv.ParseInt(metricValue, 10, 64)
							if err != nil {
								expected = getExpected(http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v", metricValue, err.Error()))
							}
						}
					}

					// SUccess
					if expected == nil {
						expected = getExpected(http.StatusOK, "ok")
					}

					tests = append(tests, testDesctiption{
						testName:    method + "_" + metricType + "_" + metricName + "_" + metricValue,
						httpMethod:  method,
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						expected:    *expected,
					})

				}
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			urlBuilder := &strings.Builder{}
			urlBuilder.WriteString("http://localhost:8080/update")
			appendIfNotEmpty(urlBuilder, tt.metricType)
			appendIfNotEmpty(urlBuilder, tt.metricName)
			appendIfNotEmpty(urlBuilder, tt.metricValue)

			metricsStorage := storage.NewInMemoryStorage()
			request := httptest.NewRequest(tt.httpMethod, urlBuilder.String(), nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handleMetric(metricsStorage))
			h.ServeHTTP(w, request)
			actual := w.Result()

			if tt.expected.status != actual.StatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expected.status, w.Code)
			}

			defer actual.Body.Close()
			resBody, err := io.ReadAll(actual.Body)
			if err != nil {
				t.Fatal(err)
			}
			if tt.expected.response != string(resBody) {
				t.Errorf("Expected body %s, got %s", tt.expected.response, w.Body.String())
			}
		})
	}
}

func appendIfNotEmpty(builder *strings.Builder, str string) {
	if str != "" {
		builder.WriteString("/")
		builder.WriteString(str)
	}
}

func getExpected(status int, response string) *expectedResult {
	return &expectedResult{
		status:   status,
		response: response,
	}
}

func getExpectedNotFound() *expectedResult {
	return getExpected(http.StatusNotFound, "404 page not found")
}

func getMethods() []string {
	return []string{
		http.MethodPost,
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
}

func getMetricType() []string {
	return []string{
		"gauge",
		"counter",
		"test",
		"",
	}
}

func getMetricName() []string {
	return []string{
		"test",
		"",
	}
}

func getMetricValue() []string {
	return []string{
		"100",
		"100.001",
		"test",
		"",
	}
}
