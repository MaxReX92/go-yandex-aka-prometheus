package html

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

func TestSimplePageBuilder_BuildMetricsPage(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics map[string]string
		gaugeMetrics   map[string]string
		expected       string
	}{
		{
			name:           "no_metric",
			counterMetrics: map[string]string{},
			gaugeMetrics:   map[string]string{},
			expected:       "<html></html>",
		}, {
			name: "all_metric",
			counterMetrics: map[string]string{
				"metricName2": "300",
				"metricName3": "-400",
				"metricName1": "200",
			},
			gaugeMetrics: map[string]string{
				"metricName5": "300.003",
				"metricName4": "100.001",
				"metricName6": "-400.004",
			},
			expected: "<html>" +
				"metricName1: 200<br>" +
				"metricName2: 300<br>" +
				"metricName3: -400<br>" +
				"metricName4: 100.001<br>" +
				"metricName5: 300.003<br>" +
				"metricName6: -400.004<br>" +
				"</html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSimplePageBuilder()
			metricsByType := map[string]map[string]string{
				"counter": tt.counterMetrics,
				"gauge":   tt.gaugeMetrics,
			}

			actual := builder.BuildMetricsPage(metricsByType)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func BenchmarkSimplePageBuilder_BuildMetricsPage(b *testing.B) {
	b.StopTimer()

	metricsList := map[string]map[string]string{
		"counter": {},
		"gauge":   {},
	}
	for i := 0; i < 100; i++ {
		metricsList["counter"]["counter"+parser.IntToString(rand.Int63())] = parser.IntToString(rand.Int63())
		metricsList["gauge"]["gauge"+parser.IntToString(rand.Int63())] = parser.FloatToString(rand.Float64())
	}
	builder := NewSimplePageBuilder()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		builder.BuildMetricsPage(metricsList)
	}
}
