package html

import "go-yandex-aka-prometheus/internal/metrics"

type HtmlPageBuilder interface {
	BuildMetricsPage(metricsByType map[string]map[string]metrics.Metric) string
}
