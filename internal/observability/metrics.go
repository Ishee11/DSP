package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	ResultBid            = "bid"
	ResultNoBid          = "no_bid"
	ResultInvalidRequest = "invalid_request"
	ResultError          = "error"
)

// Metrics groups HTTP and DSP collectors and exposes a Prometheus handler.
type Metrics struct {
	HTTP     *HTTPMetrics
	DSP      *DSPMetrics
	gatherer prometheus.Gatherer
}

// HTTPMetrics stores low-cardinality transport-level metrics.
type HTTPMetrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

// DSPMetrics stores bid-processing business metrics.
type DSPMetrics struct {
	bidRequestsTotal      prometheus.Counter
	bidResponsesTotal     *prometheus.CounterVec
	bidProcessingDuration *prometheus.HistogramVec
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusCapturingResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(p)
}

// NewMetrics creates and registers service metrics.
func NewMetrics(registerer prometheus.Registerer, gatherer prometheus.Gatherer) *Metrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}

	httpMetrics := &HTTPMetrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests handled by the service.",
			},
			[]string{"route", "method", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"route", "method", "status"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of in-flight HTTP requests.",
			},
		),
	}

	dspMetrics := &DSPMetrics{
		bidRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "dsp_bid_requests_total",
				Help: "Total number of DSP bid requests received.",
			},
		),
		bidResponsesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dsp_bid_responses_total",
				Help: "Total number of DSP bid processing results by result type.",
			},
			[]string{"result"},
		),
		bidProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dsp_bid_processing_duration_seconds",
				Help:    "DSP bid processing duration in seconds by result type.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"result"},
		),
	}

	registerer.MustRegister(
		httpMetrics.requestsTotal,
		httpMetrics.requestDuration,
		httpMetrics.requestsInFlight,
		dspMetrics.bidRequestsTotal,
		dspMetrics.bidResponsesTotal,
		dspMetrics.bidProcessingDuration,
	)

	return &Metrics{
		HTTP:     httpMetrics,
		DSP:      dspMetrics,
		gatherer: gatherer,
	}
}

// Handler exposes registered metrics in Prometheus exposition format.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
}

// Middleware instruments an HTTP handler using a normalized route label.
func (m *Metrics) Middleware(route string, next http.Handler) http.Handler {
	if m == nil || m.HTTP == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.HTTP.requestsInFlight.Inc()
		defer m.HTTP.requestsInFlight.Dec()

		recorder := &statusCapturingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		statusLabel := strconv.Itoa(statusCode)
		m.HTTP.requestsTotal.WithLabelValues(route, r.Method, statusLabel).Inc()
		m.HTTP.requestDuration.WithLabelValues(route, r.Method, statusLabel).Observe(time.Since(start).Seconds())
	})
}

// IncBidRequest counts every incoming bid request before it is validated.
func (m *DSPMetrics) IncBidRequest() {
	if m == nil {
		return
	}

	m.bidRequestsTotal.Inc()
}

// ObserveBidResult records the final DSP result and processing latency.
func (m *DSPMetrics) ObserveBidResult(result string, duration time.Duration) {
	if m == nil {
		return
	}

	m.bidResponsesTotal.WithLabelValues(result).Inc()
	m.bidProcessingDuration.WithLabelValues(result).Observe(duration.Seconds())
}
