package html

type HTMLPageBuilder interface {
	BuildMetricsPage(metricsByType map[string]map[string]string) string
}
