package html

import (
	"fmt"
	"sort"
	"strings"
)

type simplePageBuilder struct {
}

func NewSimplePageBuilder() HTMLPageBuilder {
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

	for _, metricType := range metricTypes {
		metricsList := metricsByType[metricType]
		metricNames := make([]string, len(metricsList))
		j := 0
		for metricName := range metricsList {
			metricNames[j] = metricName
			j++
		}
		sort.Strings(metricNames)

		for _, metricName := range metricNames {
			sb.WriteString(fmt.Sprintf("%v: %v", metricName, metricsList[metricName]))
			sb.WriteString("<br>")
		}
	}

	sb.WriteString("</html>")
	return sb.String()
}
