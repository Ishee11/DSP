package engine

import "github.com/Ishee11/DSP/internal/model"

type Engine struct {
	filter TargetingFilter
	scorer Scorer
}

func New() *Engine {
	return &Engine{
		filter: StaticFilter{},
		scorer: PriceScorer{},
	}
}

func NewWithFilter(filter TargetingFilter) *Engine {
	return NewWithDeps(filter, nil)
}

func NewWithDeps(filter TargetingFilter, scorer Scorer) *Engine {
	if filter == nil {
		filter = StaticFilter{}
	}
	if scorer == nil {
		scorer = PriceScorer{}
	}

	return &Engine{
		filter: filter,
		scorer: scorer,
	}
}

func (e *Engine) Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	candidates := e.filter.Filter(req, campaigns)
	return decideFromCandidates(req, candidates, e.scorer)
}

func Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	candidates := StaticFilter{}.Filter(req, campaigns)
	return decideFromCandidates(req, candidates, PriceScorer{})
}

func decideFromCandidates(
	req model.BidRequest,
	candidates []model.Campaign,
	scorer Scorer,
) (*model.BidResponse, bool) {
	if scorer == nil {
		scorer = PriceScorer{}
	}

	var best *model.Campaign
	bestScore := 0.0

	for _, c := range candidates {
		score := scorer.Score(req, c)
		if best == nil || score > bestScore || (score == bestScore && c.Price > best.Price) {
			tmp := c
			best = &tmp
			bestScore = score
		}
	}

	if best == nil {
		return nil, false
	}

	return &model.BidResponse{
		RequestID: req.RequestID,
		ImpID:     req.ImpID,
		Price:     best.Price,
		AdID:      best.ID,
	}, true
}
