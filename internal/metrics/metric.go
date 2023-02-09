package metrics

type Metric interface {
	GetName() string
	GetType() string
	StringValue() string
	SetValue(value float64)
}
