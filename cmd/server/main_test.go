package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/memory"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type callResult struct {
	status      int
	response    string
	responseObj *model.Metrics
}

type modelRequest struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type jsonAPIRequest struct {
	httpMethod string
	path       string
	request    *modelRequest
	metrics    []metrics.Metric
}

type testDescription struct {
	testName    string
	httpMethod  string
	metricType  string
	metricName  string
	metricValue string
	expected    callResult
}

type testConf struct {
	key         []byte
	singEnabled bool
}

type testDBStorage struct {
}

func Test_UpdateUrlRequest(t *testing.T) {
	tests := []testDescription{}
	for _, method := range getMethods() {
		for _, metricType := range getMetricType() {
			for _, metricName := range getMetricName() {
				for _, metricValue := range getMetricValue() {

					var expected *callResult

					// json api
					if metricType == "" && metricName == "" && metricValue == "" {
						if method == http.MethodPost {
							expected = expectedBadRequest("Invalid json: EOF\n")
						} else {
							expected = expectedNotAllowed()
						}
					}

					// Unexpected method types
					if expected == nil && method != http.MethodPost {
						if metricType == "" || metricName == "" || metricValue == "" {
							expected = expectedNotFound()
						} else {
							expected = expectedNotAllowed()
						}
					}

					// Unexpected metric types
					if expected == nil && metricType != "gauge" && metricType != "counter" {
						if metricType == "" || metricName == "" || metricValue == "" {
							expected = expectedNotFound()
						} else {
							expected = expectedNotImplemented()
						}
					}

					// Empty metric name
					if expected == nil && metricName == "" {
						expected = expectedNotFound()
					}

					// Incorrect metric value
					if expected == nil {
						if metricValue == "" {
							expected = expectedNotFound()
						} else if metricType == "gauge" {
							_, err := strconv.ParseFloat(metricValue, 64)
							if err != nil {
								expected = expectedBadRequest(fmt.Sprintf("Value parsing fail %v: %v\n", metricValue, err))
							}
						} else if metricType == "counter" {
							_, err := strconv.ParseInt(metricValue, 10, 64)
							if err != nil {
								expected = expectedBadRequest(fmt.Sprintf("Value parsing fail %v: %v\n", metricValue, err))
							}
						}
					}

					// Success
					if expected == nil {
						expected = expectedOk()
					}

					tests = append(tests, testDescription{
						testName:    "url_" + method + "_" + metricType + "_" + metricName + "_" + metricValue,
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

			metricsStorage := memory.NewInMemoryStorage()
			htmlPageBuilder := html.NewSimplePageBuilder()
			request := httptest.NewRequest(tt.httpMethod, urlBuilder.String(), nil)
			w := httptest.NewRecorder()

			conf := &testConf{key: nil, singEnabled: false}
			signer := hash.NewSigner(conf)
			converter := model.NewMetricsConverter(conf, signer)
			router := initRouter(metricsStorage, converter, htmlPageBuilder, &testDBStorage{})
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

func Test_UpdateJsonRequest_MethodNotAllowed(t *testing.T) {
	expected := expectedNotAllowed()
	for _, method := range getMethods() {
		if method == http.MethodPost || method == http.MethodGet {
			continue
		}

		t.Run("json_"+method+"_methodNotAllowed", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{httpMethod: method, path: "update"})
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_UpdateJsonRequest_MetricName(t *testing.T) {
	for _, metricType := range []string{"counter", "gauge"} {
		for _, metricName := range getMetricName() {
			requestObj := modelRequest{
				ID:    metricName,
				MType: metricType,
			}

			var expected *callResult
			if metricName == "" {
				expected = expectedBadRequest("metric name is missed\n")
			} else {
				if metricType == "counter" {
					delta := int64(100)
					requestObj.Delta = &delta
					expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", &delta, nil)
				} else if metricType == "gauge" {
					value := float64(100)
					requestObj.Value = &value
					expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", nil, &value)
				}
			}

			t.Run("json_"+metricName+"_"+metricType+"_metricName", func(t *testing.T) {
				actual := runJSONTest(t, jsonAPIRequest{httpMethod: http.MethodPost, path: "update", request: &requestObj})
				assert.Equal(t, expected, actual)
			})
		}
	}
}

func Test_UpdateJsonRequest_MetricType(t *testing.T) {
	for _, metricType := range getMetricType() {
		requestObj := modelRequest{
			ID:    "testMetricName",
			MType: metricType,
		}

		var expected *callResult
		if metricType == "" {
			expected = expectedBadRequest("metric types is missed\n")
		} else if metricType == "counter" {
			delta := int64(100)
			requestObj.Delta = &delta
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", &delta, nil)
		} else if metricType == "gauge" {
			value := float64(100)
			requestObj.Value = &value
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", nil, &value)
		} else {
			expected = expectedNotImplemented()
		}

		t.Run("json_"+metricType+"_metricType", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{httpMethod: http.MethodPost, path: "update", request: &requestObj})
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_UpdateJsonRequest_CounterMetricValue(t *testing.T) {
	delta := int64(100)
	for _, metricValue := range []*int64{nil, &delta} {
		requestObj := modelRequest{
			ID:    "testMetricName",
			MType: "counter",
			Delta: metricValue,
		}

		var valueString string
		var expected *callResult
		if metricValue == nil {
			valueString = "nil"
			expected = expectedBadRequest("metric value is missed\n")
		} else {
			valueString = parser.IntToString(*metricValue)
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", metricValue, nil)
		}

		t.Run("json_"+valueString+"_counterMetricValue", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{httpMethod: http.MethodPost, path: "update", request: &requestObj})
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_UpdateJsonRequest_GaugeMetricValue(t *testing.T) {
	value := float64(100)
	for _, metricValue := range []*float64{nil, &value} {
		requestObj := modelRequest{
			ID:    "testMetricName",
			MType: "gauge",
			Value: metricValue,
		}

		var valueString string
		var expected *callResult
		if metricValue == nil {
			valueString = "nil"
			expected = expectedBadRequest("metric value is missed\n")
		} else {
			valueString = parser.FloatToString(*metricValue)
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", nil, metricValue)
		}

		t.Run("json_"+valueString+"_gaugeMetricValue", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{httpMethod: http.MethodPost, path: "update", request: &requestObj})
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_GetMetricUrlRequest(t *testing.T) {
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
			metricsStorage := memory.NewInMemoryStorage()
			_, err := metricsStorage.AddMetricValue(context.Background(), createCounterMetric("metricName", 100))
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			conf := &testConf{key: nil, singEnabled: false}
			signer := hash.NewSigner(conf)
			converter := model.NewMetricsConverter(conf, signer)
			router := initRouter(metricsStorage, converter, htmlPageBuilder, &testDBStorage{})
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

func Test_GetMetricJsonRequest_MethodNotAllowed(t *testing.T) {
	expected := expectedNotAllowed()
	for _, method := range getMethods() {
		if method == http.MethodPost {
			continue
		}

		t.Run("json_"+method+"_methodNotAllowed", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{httpMethod: method, path: "value"})
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_GetMetricJsonRequest_MetricName(t *testing.T) {
	for _, metricType := range []string{"counter", "gauge"} {
		for _, metricName := range getMetricName() {
			requestObj := modelRequest{
				ID:    metricName,
				MType: metricType,
			}

			var expected *callResult
			metricList := []metrics.Metric{}

			if metricName == "" {
				expected = expectedBadRequest("metric name is missed\n")
			} else {
				if metricType == "counter" {
					delta := int64(100)
					metricList = append(metricList, createCounterMetric(requestObj.ID, float64(delta)))
					expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", &delta, nil)
				} else if metricType == "gauge" {
					value := float64(100)
					metricList = append(metricList, createGaugeMetric(requestObj.ID, value))
					expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", nil, &value)
				}
			}

			t.Run("json_"+metricName+"_"+metricType+"_metricName", func(t *testing.T) {
				actual := runJSONTest(t, jsonAPIRequest{
					httpMethod: http.MethodPost,
					path:       "value",
					request:    &requestObj,
					metrics:    metricList,
				})
				assert.Equal(t, expected, actual)
			})
		}
	}
}

func Test_GetMetricJsonRequest_MetricType(t *testing.T) {
	for _, metricType := range getMetricType() {
		requestObj := modelRequest{
			ID:    "testMetricName",
			MType: metricType,
		}

		var expected *callResult
		metricList := []metrics.Metric{}

		if metricType == "" {
			expected = expectedBadRequest("metric types is missed\n")
		} else if metricType == "counter" {
			delta := int64(100)
			metricList = append(metricList, createCounterMetric(requestObj.ID, float64(delta)))
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", &delta, nil)
		} else if metricType == "gauge" {
			value := float64(100)
			metricList = append(metricList, createGaugeMetric(requestObj.ID, value))
			expected = getExpectedObj(200, requestObj.MType, requestObj.ID, "", nil, &value)
		} else {
			expected = expectedNotFoundMessage("Metric not found\n")
		}

		t.Run("json_"+metricType+"_metricType", func(t *testing.T) {
			actual := runJSONTest(t, jsonAPIRequest{
				httpMethod: http.MethodPost,
				path:       "value",
				request:    &requestObj,
				metrics:    metricList,
			})
			assert.Equal(t, expected, actual)
		})
	}
}

func runJSONTest(t *testing.T, apiRequest jsonAPIRequest) *callResult {
	var buffer bytes.Buffer
	metricsStorage := memory.NewInMemoryStorage()
	if apiRequest.metrics != nil {
		for _, metric := range apiRequest.metrics {
			_, err := metricsStorage.AddMetricValue(context.Background(), metric)
			assert.NoError(t, err)
		}
	}
	htmlPageBuilder := html.NewSimplePageBuilder()

	if apiRequest.request != nil {
		encoder := json.NewEncoder(&buffer)
		err := encoder.Encode(apiRequest.request)
		require.NoError(t, err)
	}

	request := httptest.NewRequest(apiRequest.httpMethod, "http://localhost:8080/"+apiRequest.path, &buffer)
	w := httptest.NewRecorder()

	conf := &testConf{}
	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	router := initRouter(metricsStorage, converter, htmlPageBuilder, &testDBStorage{})
	router.ServeHTTP(w, request)
	actual := w.Result()
	result := &callResult{status: actual.StatusCode}

	defer actual.Body.Close()
	resBody, _ := io.ReadAll(actual.Body)
	resultObj := &model.Metrics{}
	err := json.Unmarshal(resBody, resultObj)
	if err != nil {
		result.response = string(resBody)
	} else {
		result.responseObj = resultObj
	}

	return result
}

func appendIfNotEmpty(builder *strings.Builder, str string) {
	if str != "" {
		builder.WriteString("/")
		builder.WriteString(str)
	}
}

func expectedNotFound() *callResult {
	return expectedNotFoundMessage("404 page not found\n")
}

func expectedNotFoundMessage(message string) *callResult {
	return getExpected(http.StatusNotFound, message)
}

func expectedNotAllowed() *callResult {
	return getExpected(http.StatusMethodNotAllowed, "")
}

func expectedBadRequest(message string) *callResult {
	return getExpected(http.StatusBadRequest, message)
}

func expectedNotImplemented() *callResult {
	return getExpected(http.StatusNotImplemented, "unknown metric types\n")
}

func expectedOk() *callResult {
	return getExpected(http.StatusOK, "ok")
}

func getExpected(status int, response string) *callResult {
	return &callResult{
		status:   status,
		response: response,
	}
}

func getExpectedObj(status int, metricType string, metricName string, errorString string, delta *int64, value *float64) *callResult {
	return &callResult{
		status:   status,
		response: errorString,
		responseObj: &model.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: delta,
			Value: value,
		},
	}
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

func createCounterMetric(name string, value float64) metrics.Metric {
	return createMetric(types.NewCounterMetric, name, value)
}

func createGaugeMetric(name string, value float64) metrics.Metric {
	return createMetric(types.NewGaugeMetric, name, value)
}

func createMetric(metricFactory func(string) metrics.Metric, name string, value float64) metrics.Metric {
	metric := metricFactory(name)
	metric.SetValue(value)
	return metric
}

func (t *testConf) SignMetrics() bool {
	return t.singEnabled
}

func (t *testConf) GetKey() []byte {
	return t.key
}

func (t testDBStorage) Ping(context.Context) error {
	return nil
}

func (t testDBStorage) Close() error {
	return nil
}

func (t *testDBStorage) UpdateRecords(ctx context.Context, records []*database.DBRecord) error {
	//TODO implement me
	panic("implement me")
}

func (t *testDBStorage) ReadRecord(ctx context.Context, metricType string, metricName string) (*database.DBRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (t *testDBStorage) ReadAll(ctx context.Context) ([]*database.DBRecord, error) {
	//TODO implement me
	panic("implement me")
}
