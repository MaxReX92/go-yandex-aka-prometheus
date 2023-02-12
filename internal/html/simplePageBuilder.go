package html

import (
	"fmt"
	"sort"
	"strings"
)

type simplePageBuilder struct {
}

func NewSimplePageBuilder() HtmlPageBuilder {
	return &simplePageBuilder{}
}

func (s simplePageBuilder) BuildMetricsPage(metricsByType map[string]map[string]string) string {
	sb := strings.Builder{}
	sb.WriteString("<html>")

	metricTypes := make([]string, len(metricsByType))
	i := 0
	for metricType := range metricsByType {
		metricTypes[i] = metricType
		i++
	}
	sort.Strings(metricTypes)

	for _, metricsList := range metricsByType {
		metricNames := make([]string, len(metricsList))
		j := 0
		for key := range metricsList {
			metricNames[j] = key
			j++
		}
		sort.Strings(metricNames)

		for _, key := range metricNames {
			sb.WriteString(fmt.Sprintf("%v: %v", key, metricsList[key]))
			sb.WriteString("<br>")
		}
	}

	sb.WriteString("</html>")
	return sb.String()
}
