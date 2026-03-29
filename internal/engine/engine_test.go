package engine

import (
	"testing"

	"github.com/Ishee11/DSP/internal/model"
)

func TestDecide(t *testing.T) {
	tests := []struct {
		name       string
		req        model.BidRequest
		campaigns  []model.Campaign
		expectBid  bool
		expectAdID string
	}{
		{
			name: "single matching campaign",
			req: model.BidRequest{
				RequestID:  "r1",
				ImpID:      "imp1",
				SiteID:     "1",
				DeviceType: "mobile",
				FloorPrice: 1.0,
			},
			campaigns: []model.Campaign{
				{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.2},
			},
			expectBid:  true,
			expectAdID: "c1",
		},
		{
			name: "no matching site",
			req: model.BidRequest{
				SiteID:     "2",
				DeviceType: "mobile",
				FloorPrice: 1.0,
			},
			campaigns: []model.Campaign{
				{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.2},
			},
			expectBid: false,
		},
		{
			name: "below floor price",
			req: model.BidRequest{
				SiteID:     "1",
				DeviceType: "mobile",
				FloorPrice: 2.0,
			},
			campaigns: []model.Campaign{
				{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.5},
			},
			expectBid: false,
		},
		{
			name: "choose highest price",
			req: model.BidRequest{
				SiteID:     "1",
				DeviceType: "mobile",
				FloorPrice: 0.5,
			},
			campaigns: []model.Campaign{
				{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.0},
				{ID: "c2", SiteID: "1", DeviceType: "mobile", Price: 1.5},
			},
			expectBid:  true,
			expectAdID: "c2",
		},
		{
			name: "device mismatch",
			req: model.BidRequest{
				SiteID:     "1",
				DeviceType: "desktop",
				FloorPrice: 0.5,
			},
			campaigns: []model.Campaign{
				{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.5},
			},
			expectBid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			resp, ok := engine.Decide(tt.req, tt.campaigns)

			if ok != tt.expectBid {
				t.Fatalf("expected bid=%v, got=%v", tt.expectBid, ok)
			}

			if !ok {
				return
			}

			if resp.AdID != tt.expectAdID {
				t.Fatalf("expected adID=%s, got=%s", tt.expectAdID, resp.AdID)
			}
		})
	}
}
