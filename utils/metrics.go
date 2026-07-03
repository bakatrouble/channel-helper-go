package utils

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg *prometheus.Registry

	postsUnsent prometheus.Gauge
}

func (m *Metrics) SetPostsUnsent(postsUnsent int) {
	m.postsUnsent.Set(float64(postsUnsent))
}

func (m *Metrics) PromHandler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{Registry: m.reg})
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		reg: reg,
		postsUnsent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "posts_unsent",
			Help: "Number of posts unsent",
		}),
	}
	reg.MustRegister(m.postsUnsent)
	return m
}
