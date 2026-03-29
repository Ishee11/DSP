package engine

import "github.com/Ishee11/DSP/internal/model"

// Scorer assigns comparable numeric scores to eligible campaigns.
// Scorer присваивает допустимым кампаниям сравнимые числовые оценки.
type Scorer interface {
	Score(req model.BidRequest, campaign model.Campaign) float64
}

// PriceScorer keeps current behavior: higher campaign price wins.
// PriceScorer сохраняет текущее поведение: выигрывает кампания с более высокой ценой.
type PriceScorer struct{}

// Score returns campaign price as the ranking score.
// Score возвращает цену кампании как итоговый ranking score.
func (s PriceScorer) Score(_ model.BidRequest, campaign model.Campaign) float64 {
	return campaign.Price
}
