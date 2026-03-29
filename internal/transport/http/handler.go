package http

import (
	"encoding/json"
	"net/http"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/model"
)

// Handler adapts HTTP requests to the bidding engine API.
// Handler адаптирует HTTP-запросы к API bidding engine.
type Handler struct {
	engine    *engine.Engine
	campaigns []model.Campaign
}

// New constructs an HTTP handler with the engine and campaign source.
// New создаёт HTTP handler с движком и источником кампаний.
func New(e *engine.Engine, campaigns []model.Campaign) *Handler {
	return &Handler{
		engine:    e,
		campaigns: campaigns,
	}
}

// Bid decodes a bid request, runs campaign selection, and writes either no-bid or JSON response.
// Bid декодирует bid request, запускает выбор кампании и записывает либо no-bid, либо JSON-ответ.
func (h *Handler) Bid(w http.ResponseWriter, r *http.Request) {
	var req model.BidRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, ok := h.engine.Decide(req, h.campaigns)

	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
