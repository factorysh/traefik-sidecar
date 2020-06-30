package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	BackendBirth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_birth",
		Help: "Amount of new backends",
	})
	BackendDeath = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_death",
		Help: "Amount of closed backends",
	})
)
