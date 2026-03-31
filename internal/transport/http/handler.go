package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/model"
	"github.com/Ishee11/DSP/internal/observability"
	"github.com/Ishee11/DSP/internal/storage"
)

type DSPMetrics interface {
	IncBidRequest()
	ObserveBidResult(result string, duration time.Duration)
}

// Handler adapts HTTP requests to the bidding engine API.
// Handler адаптирует HTTP-запросы к API bidding engine.
type Handler struct {
	engine       *engine.Engine
	campaignRepo storage.CampaignRepository
	metrics      DSPMetrics
}

// New constructs an HTTP handler with the engine and campaign source.
// New создаёт HTTP handler с движком и источником кампаний.
func New(e *engine.Engine, campaignRepo storage.CampaignRepository, metrics DSPMetrics) *Handler {
	if metrics == nil {
		metrics = noopDSPMetrics{}
	}
	if campaignRepo == nil {
		campaignRepo = storage.NewInMemoryCampaignRepository(nil)
	}

	return &Handler{
		engine:       e,
		campaignRepo: campaignRepo,
		metrics:      metrics,
	}
}

// Bid decodes a bid request, runs campaign selection, and writes either no-bid or JSON response.
// Bid декодирует bid request, запускает выбор кампании и записывает либо no-bid, либо JSON-ответ.
func (h *Handler) Bid(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.metrics.IncBidRequest()

	var req model.BidRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.metrics.ObserveBidResult(observability.ResultInvalidRequest, time.Since(start))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaigns, err := h.campaignRepo.ListCampaigns(r.Context())
	if err != nil {
		h.metrics.ObserveBidResult(observability.ResultError, time.Since(start))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, ok := h.engine.Decide(req, campaigns)

	if !ok {
		h.metrics.ObserveBidResult(observability.ResultNoBid, time.Since(start))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.metrics.ObserveBidResult(observability.ResultError, time.Since(start))
		return
	}

	h.metrics.ObserveBidResult(observability.ResultBid, time.Since(start))
}

type noopDSPMetrics struct{}

func (noopDSPMetrics) IncBidRequest() {}

func (noopDSPMetrics) ObserveBidResult(string, time.Duration) {}
