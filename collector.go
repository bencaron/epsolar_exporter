package collector

import "github.com/prometheus/client_golang/prometheus"

type MyCollector struct {
	counterDesc *prometheus.Desc
}

func (c *MyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.counterDesc
}

func (c *MyCollector) Collect(ch chan<- prometheus.Metric) {
	value := 1.0 // Your code to fetch the counter value goes here.
	ch <- prometheus.MustNewConstMetric(
		c.counterDesc,
		prometheus.CounterValue,
		value,
	)
}

func NewMyCollector() *MyCollector {
	return &MyCollector{
		counterDesc: prometheus.NewDesc("my_counter_total", "Help string", nil, nil),
	}
}

// To hook in the collector: prometheus.MustRegister(NewMyCollector())
