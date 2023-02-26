package metrics

type Metric interface {
	GetName() string
	GetType() string
	GetValue() float64
	GetStringValue() string
	SetValue(value float64) float64
	Flush()
}
