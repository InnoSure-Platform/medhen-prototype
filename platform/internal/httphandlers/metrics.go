package httphandlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MountMetrics(r chi.Router) {
	r.Handle("/metrics", promhttp.Handler())
}
