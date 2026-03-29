package engine

import (
	"testing"

	"github.com/Ishee11/DSP/internal/model"
)

func TestStaticFilter_Filter(t *testing.T) {
	filter := StaticFilter{}

	campaigns := []model.Campaign{
		{ID: "site-mismatch", SiteID: "2", DeviceType: "mobile", Price: 3.0},
		{ID: "device-mismatch", SiteID: "1", DeviceType: "desktop", Price: 3.0},
		{ID: "below-floor", SiteID: "1", DeviceType: "mobile", Price: 0.5},
		{ID: "ok-1", SiteID: "1", DeviceType: "mobile", Price: 1.1},
		{ID: "ok-2", SiteID: "1", DeviceType: "mobile", Price: 2.2},
	}

	req := model.BidRequest{SiteID: "1", DeviceType: "mobile", FloorPrice: 1.0}

	filtered := filter.Filter(req, campaigns)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 campaigns after filtering, got %d", len(filtered))
	}

	if filtered[0].ID != "ok-1" || filtered[1].ID != "ok-2" {
		t.Fatalf("unexpected filter result order/content: %+v", filtered)
	}
}
