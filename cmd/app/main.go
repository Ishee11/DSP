package main

import (
	"log"
	"net/http"

	httpTransport "github.com/Ishee11/DSP/internal/transport/http"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/observability"
	"github.com/Ishee11/DSP/internal/storage"
)

// main boots a minimal DSP HTTP server with in-memory campaigns.
// main поднимает минимальный DSP HTTP-сервер с кампаниями в памяти.
func main() {
	e := engine.New()
	campaignRepo := storage.NewInMemoryCampaignRepository(storage.DemoCampaigns())
	metrics := observability.NewMetrics(nil, nil)
	h := httpTransport.New(e, campaignRepo, metrics.DSP)
	mux := httpTransport.NewMux(h, metrics)

	log.Println("DSP started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
