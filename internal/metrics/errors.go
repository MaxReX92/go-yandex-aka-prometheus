package metrics

import "errors"

var (
	ErrEmptyURL                 = errors.New("empty url string")
	ErrFieldNameNotFound        = errors.New("field name was not found")
	ErrInvalidRecordMetricType  = errors.New("invalid record metric type")
	ErrInvalidRecordMetricName  = errors.New("invalid record metric name")
	ErrInvalidRecordMetricValue = errors.New("invalid record metric value")
	ErrInvalidSignature         = errors.New("invalid signature")
	ErrMetricNotFound           = errors.New("metric not found")
	ErrMetricValueMissed        = errors.New("metric value is missed")
	ErrUnexpectedStatusCode     = errors.New("unexpected status code")
	ErrUnknownMetricType        = errors.New("unknown metric type")
)
