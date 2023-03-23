package html

type PageBuilder interface {
	BuildMetricsPage(metricsByType map[string]map[string]string) string
}
