package main

import (
	"log"
	"net/http"

	httpTransport "github.com/Ishee11/DSP/internal/transport/http"

	"github.com/Ishee11/DSP/internal/engine"
	"github.com/Ishee11/DSP/internal/model"
	"github.com/Ishee11/DSP/internal/observability"
)

// main boots a minimal DSP HTTP server with in-memory campaigns.
// main поднимает минимальный DSP HTTP-сервер с кампаниями в памяти.
func main() {
	// Demo campaigns used as an in-memory source instead of a real storage layer.
	// Демонстрационные кампании используются как in-memory источник вместо реального слоя хранения.
	campaigns := []model.Campaign{
		{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.2},
		{ID: "c2", SiteID: "1", DeviceType: "desktop", Price: 0.8},
		{ID: "c3", SiteID: "2", DeviceType: "mobile", Price: 1.5},
	}

	e := engine.New()
	metrics := observability.NewMetrics(nil, nil)
	h := httpTransport.New(e, campaigns, metrics.DSP)
	mux := httpTransport.NewMux(h, metrics)

	log.Println("DSP started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
