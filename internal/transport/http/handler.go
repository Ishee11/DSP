package http

import (
	"encoding/json"
	"net/http"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/model"
)

type Handler struct {
	engine    *engine.Engine
	campaigns []model.Campaign
}

func New(e *engine.Engine, campaigns []model.Campaign) *Handler {
	return &Handler{
		engine:    e,
		campaigns: campaigns,
	}
}

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
