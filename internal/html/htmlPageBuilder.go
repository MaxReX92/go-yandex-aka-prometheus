package html

type HtmlPageBuilder interface {
	BuildMetricsPage(metricsByType map[string]map[string]string) string
}
