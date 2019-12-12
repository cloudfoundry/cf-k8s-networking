package metrics

import (
	"net/http"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DefaultMetrics = initMetrics()
)

const metricsNamespace = "cfroutesync"

type Metrics struct {
	Handler        http.Handler
	ObservedValues ObservedValues
}

type ObservedValues struct {
	LastUpdatedAt  prometheus.Gauge
	NumberOfRoutes prometheus.Gauge
}

func initMetrics() Metrics {
	m := Metrics{
		Handler: promhttp.Handler(),
		ObservedValues: ObservedValues{
			LastUpdatedAt: prometheus.NewGauge(
				prometheus.GaugeOpts{Namespace: metricsNamespace, Name: "last_updated_at", Help: "Unix timestamp indicating last successful sync"}),
			NumberOfRoutes: prometheus.NewGauge(
				prometheus.GaugeOpts{Namespace: metricsNamespace, Name: "fetched_routes", Help: "Number of routes fetched from Cloud Controller"}),
		},
	}

	prometheus.MustRegister(m.ObservedValues.LastUpdatedAt)
	prometheus.MustRegister(m.ObservedValues.NumberOfRoutes)

	return m
}

func Update(snapshot *models.RouteSnapshot) {
	DefaultMetrics.ObservedValues.LastUpdatedAt.SetToCurrentTime()
	DefaultMetrics.ObservedValues.NumberOfRoutes.Set(float64(len(snapshot.Routes)))
}
