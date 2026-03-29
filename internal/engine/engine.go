package engine

import "github.com/Ishee11/DSP/internal/model"

type Engine struct {
	filter TargetingFilter
}

func New() *Engine {
	return &Engine{filter: StaticFilter{}}
}

func NewWithFilter(filter TargetingFilter) *Engine {
	if filter == nil {
		filter = StaticFilter{}
	}

	return &Engine{filter: filter}
}

func (e *Engine) Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	candidates := e.filter.Filter(req, campaigns)
	return decideFromCandidates(req, candidates)
}

func Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	candidates := StaticFilter{}.Filter(req, campaigns)
	return decideFromCandidates(req, candidates)
}

func decideFromCandidates(
	req model.BidRequest,
	candidates []model.Campaign,
) (*model.BidResponse, bool) {
	var best *model.Campaign

	for _, c := range candidates {
		if best == nil || c.Price > best.Price {
			tmp := c
			best = &tmp
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
