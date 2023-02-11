package main

import "time"

type config struct {
	serverURL             string
	pushTimeout           time.Duration
	sendMetricsInterval   time.Duration
	updateMetricsInterval time.Duration
	collectMetricsList    []string
}

func (c *config) MetricsList() []string {
	return c.collectMetricsList
}

func (c *config) MetricsServerURL() string {
	return c.serverURL
}

func (c *config) PushMetricsTimeout() time.Duration {
	return c.pushTimeout
}
