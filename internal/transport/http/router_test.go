package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/model"
	"github.com/Ishee11/DSP/internal/observability"
	"github.com/Ishee11/DSP/internal/storage"
)

func TestMetricsEndpointIsAvailable(t *testing.T) {
	router, _ := newTestRouter()

	bidReq := httptest.NewRequest(http.MethodPost, "/bid", strings.NewReader(`{
		"request_id":"r1",
		"imp_id":"imp1",
		"site_id":"1",
		"placement_id":"p1",
		"floor_price":1.0,
		"user_id":"u1",
		"device_type":"mobile",
		"ts":1710000000
	}`))
	bidRec := httptest.NewRecorder()
	router.ServeHTTP(bidRec, bidReq)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Fatalf("expected metrics payload to contain http_requests_total, got %q", body)
	}
	if !strings.Contains(body, "dsp_bid_requests_total") {
		t.Fatalf("expected metrics payload to contain dsp_bid_requests_total, got %q", body)
	}
}

func TestBidEndpointUpdatesHTTPAndDSPMetrics(t *testing.T) {
	router, registry := newTestRouter()

	body := `{
		"request_id":"r1",
		"imp_id":"imp1",
		"site_id":"1",
		"placement_id":"p1",
		"floor_price":1.0,
		"user_id":"u1",
		"device_type":"mobile",
		"ts":1710000000
	}`
	req := httptest.NewRequest(http.MethodPost, "/bid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if got := counterValue(t, registry, "dsp_bid_requests_total", nil); got != 1 {
		t.Fatalf("expected dsp_bid_requests_total=1, got %v", got)
	}
	if got := counterValue(t, registry, "dsp_bid_responses_total", map[string]string{"result": observability.ResultBid}); got != 1 {
		t.Fatalf("expected dsp_bid_responses_total{result=bid}=1, got %v", got)
	}
	if got := counterValue(t, registry, "http_requests_total", map[string]string{
		"route":  "/bid",
		"method": http.MethodPost,
		"status": "200",
	}); got != 1 {
		t.Fatalf("expected http_requests_total for /bid POST 200 to be 1, got %v", got)
	}
	if got := gaugeValue(t, registry, "http_requests_in_flight", nil); got != 0 {
		t.Fatalf("expected http_requests_in_flight=0 after request completion, got %v", got)
	}
	if got := histogramCount(t, registry, "http_request_duration_seconds", map[string]string{
		"route":  "/bid",
		"method": http.MethodPost,
		"status": "200",
	}); got != 1 {
		t.Fatalf("expected http_request_duration_seconds count=1, got %d", got)
	}
	if got := histogramCount(t, registry, "dsp_bid_processing_duration_seconds", map[string]string{
		"result": observability.ResultBid,
	}); got != 1 {
		t.Fatalf("expected dsp_bid_processing_duration_seconds{result=bid} count=1, got %d", got)
	}
}

func TestInvalidBidRequestUpdatesErrorMetrics(t *testing.T) {
	router, registry := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/bid", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	if got := counterValue(t, registry, "dsp_bid_requests_total", nil); got != 1 {
		t.Fatalf("expected dsp_bid_requests_total=1, got %v", got)
	}
	if got := counterValue(t, registry, "dsp_bid_responses_total", map[string]string{"result": observability.ResultInvalidRequest}); got != 1 {
		t.Fatalf("expected invalid_request counter=1, got %v", got)
	}
	if got := counterValue(t, registry, "http_requests_total", map[string]string{
		"route":  "/bid",
		"method": http.MethodPost,
		"status": "400",
	}); got != 1 {
		t.Fatalf("expected http_requests_total for /bid POST 400 to be 1, got %v", got)
	}
}

func TestBidEndpointReturnsInternalServerErrorWhenCampaignRepositoryFails(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry, registry)
	handler := New(engine.New(), failingCampaignRepository{}, metrics.DSP)
	router := NewMux(handler, metrics)

	req := httptest.NewRequest(http.MethodPost, "/bid", strings.NewReader(`{
		"request_id":"r1",
		"imp_id":"imp1",
		"site_id":"1",
		"placement_id":"p1",
		"floor_price":1.0,
		"user_id":"u1",
		"device_type":"mobile",
		"ts":1710000000
	}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	if got := counterValue(t, registry, "dsp_bid_responses_total", map[string]string{"result": observability.ResultError}); got != 1 {
		t.Fatalf("expected error counter=1, got %v", got)
	}
}

func newTestRouter() (*http.ServeMux, *prometheus.Registry) {
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry, registry)

	campaigns := []model.Campaign{
		{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.2},
		{ID: "c2", SiteID: "1", DeviceType: "desktop", Price: 0.8},
	}

	handler := New(engine.New(), storage.NewInMemoryCampaignRepository(campaigns), metrics.DSP)

	return NewMux(handler, metrics), registry
}

type failingCampaignRepository struct{}

func (failingCampaignRepository) ListCampaigns(context.Context) ([]model.Campaign, error) {
	return nil, errors.New("repository unavailable")
}

func counterValue(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()

	metric := findMetric(t, registry, name, labels)
	return metric.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()

	metric := findMetric(t, registry, name, labels)
	return metric.GetGauge().GetValue()
}

func histogramCount(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) uint64 {
	t.Helper()

	metric := findMetric(t, registry, name, labels)
	return metric.GetHistogram().GetSampleCount()
}

func findMetric(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) *dto.Metric {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() != name {
			continue
		}

		for _, metric := range family.GetMetric() {
			if labelsMatch(metric, labels) {
				return metric
			}
		}
	}

	t.Fatalf("metric %s with labels %v not found", name, labels)
	return nil
}

func labelsMatch(metric *dto.Metric, expected map[string]string) bool {
	if len(expected) == 0 {
		return len(metric.GetLabel()) == 0
	}

	if len(metric.GetLabel()) != len(expected) {
		return false
	}

	for _, label := range metric.GetLabel() {
		value, ok := expected[label.GetName()]
		if !ok || value != label.GetValue() {
			return false
		}
	}

	return true
}
