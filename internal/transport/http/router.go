package http

import (
	"net/http"

	"github.com/Ishee11/DSP/internal/observability"
)

const (
	bidRouteLabel     = "/bid"
	metricsRouteLabel = "/metrics"
)

// NewMux builds the service HTTP router with normalized route instrumentation.
func NewMux(handler *Handler, metrics *observability.Metrics) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(bidRouteLabel, metrics.Middleware(bidRouteLabel, http.HandlerFunc(handler.Bid)))
	mux.Handle(metricsRouteLabel, metrics.Middleware(metricsRouteLabel, metrics.Handler()))

	return mux
}
