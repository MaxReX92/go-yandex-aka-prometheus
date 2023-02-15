package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/storage"
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
						if metricType == "" || metricName == "" || metricValue == "" {
							expected = getExpectedNotFound()
						} else {
							expected = getExpected(http.StatusMethodNotAllowed, "")
						}
					}

					// Unexpected metric type
					if expected == nil && metricType != "gauge" && metricType != "counter" {
						if metricType == "" || metricName == "" || metricValue == "" {
							expected = getExpectedNotFound()
						} else {
							expected = getExpected(http.StatusNotImplemented, "Unknown metric type\n")
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
								expected = getExpected(http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v\n", metricValue, err.Error()))
							}
						} else if metricType == "counter" {
							_, err := strconv.ParseInt(metricValue, 10, 64)
							if err != nil {
								expected = getExpected(http.StatusBadRequest, fmt.Sprintf("Value parsing fail %v: %v\n", metricValue, err.Error()))
							}
						}
					}

					// Success
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
			htmlPageBuilder := html.NewSimplePageBuilder()
			request := httptest.NewRequest(tt.httpMethod, urlBuilder.String(), nil)
			w := httptest.NewRecorder()
			router := initRouter(metricsStorage, htmlPageBuilder)
			router.ServeHTTP(w, request)
			actual := w.Result()

			assert.Equal(t, tt.expected.status, actual.StatusCode)

			defer actual.Body.Close()
			resBody, err := io.ReadAll(actual.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expected.response, string(resBody))
		})
	}
}

func Test_GetMetricValue(t *testing.T) {
	tests := []struct {
		name          string
		metricType    string
		metricName    string
		expectSuccess bool
	}{
		{
			name:          "type_not_found",
			metricType:    "not_existed_type",
			metricName:    "metricName",
			expectSuccess: false,
		},
		{
			name:          "metric_name_not_found",
			metricType:    "counter",
			metricName:    "not_existed_metric_name",
			expectSuccess: false,
		},
		{
			name:          "success_get_value",
			metricType:    "counter",
			metricName:    "metricName",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:8080/value/%v/%v", tt.metricType, tt.metricName)

			htmlPageBuilder := html.NewSimplePageBuilder()
			metricsStorage := storage.NewInMemoryStorage()
			metricsStorage.AddCounterMetricValue("metricName", 100)

			request := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router := initRouter(metricsStorage, htmlPageBuilder)
			router.ServeHTTP(w, request)
			actual := w.Result()

			if tt.expectSuccess {
				assert.Equal(t, http.StatusOK, actual.StatusCode)
				defer actual.Body.Close()
				body, err := io.ReadAll(actual.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "100", string(body))
			} else {
				assert.Equal(t, http.StatusNotFound, actual.StatusCode)
				defer actual.Body.Close()
				body, err := io.ReadAll(actual.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "Metric not found\n", string(body))
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
	return getExpected(http.StatusNotFound, "404 page not found\n")
}

func getMethods() []string {
	return []string{
		http.MethodPost,
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
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
