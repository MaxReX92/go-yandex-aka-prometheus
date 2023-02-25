package metrics

type Metric interface {
	GetName() string
	GetType() string
	GetStringValue() string
	SetValue(value float64) float64
	Flush()
}
