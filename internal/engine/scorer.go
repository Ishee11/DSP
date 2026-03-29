package engine

import "github.com/Ishee11/DSP/internal/model"

// Scorer assigns comparable numeric scores to eligible campaigns.
type Scorer interface {
	Score(req model.BidRequest, campaign model.Campaign) float64
}

// PriceScorer keeps current behavior: higher campaign price wins.
type PriceScorer struct{}

func (s PriceScorer) Score(_ model.BidRequest, campaign model.Campaign) float64 {
	return campaign.Price
}
