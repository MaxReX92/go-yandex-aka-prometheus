package metrics

type CounterMetric struct {
	name  string
	value int64
}

func (m CounterMetric) GetType() string {
	return "counter"
}

func (m CounterMetric) GetName() string {
	return m.name
}

func (m CounterMetric) StringValue() string {
	return string(m.value)
}

func (m CounterMetric) SetValue(value float64) {
	m.value = m.value + int64(value)
}
