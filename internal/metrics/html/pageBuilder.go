package html

// PageBuilder is a provider of served metrics report.
type PageBuilder interface {

	// BuildMetricsPage build and return served metrics report.
	BuildMetricsPage(metricsByType map[string]map[string]string) string
}
