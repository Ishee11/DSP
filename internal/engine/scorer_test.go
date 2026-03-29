package engine

import (
	"testing"

	"github.com/Ishee11/DSP/internal/model"
)

type fixedScorer struct {
	scores map[string]float64
}

func (s fixedScorer) Score(_ model.BidRequest, campaign model.Campaign) float64 {
	return s.scores[campaign.ID]
}

func TestEngine_Decide_UsesScorer(t *testing.T) {
	e := NewWithDeps(
		StaticFilter{},
		fixedScorer{scores: map[string]float64{"c1": 0.1, "c2": 0.9}},
	)

	req := model.BidRequest{
		RequestID:  "r1",
		ImpID:      "imp1",
		SiteID:     "1",
		DeviceType: "mobile",
		FloorPrice: 0.1,
	}

	campaigns := []model.Campaign{
		{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 3.0},
		{ID: "c2", SiteID: "1", DeviceType: "mobile", Price: 1.0},
	}

	resp, ok := e.Decide(req, campaigns)
	if !ok {
		t.Fatalf("expected bid, got no-bid")
	}

	if resp.AdID != "c2" {
		t.Fatalf("expected winner c2 by score, got %s", resp.AdID)
	}
}

func TestPriceScorer_Score(t *testing.T) {
	s := PriceScorer{}
	camp := model.Campaign{ID: "c1", Price: 1.75}
	if got := s.Score(model.BidRequest{}, camp); got != 1.75 {
		t.Fatalf("expected score=price=1.75, got %v", got)
	}
}
