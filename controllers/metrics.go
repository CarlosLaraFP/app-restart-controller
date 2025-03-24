package controllers

import "github.com/prometheus/client_golang/prometheus"

var (
	RestartedDeployments = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "apprestart_restarts_total",
			Help: "Number of Deployments restarted by the controller",
		},
	)
)

func init() {
	prometheus.MustRegister(RestartedDeployments)
}
