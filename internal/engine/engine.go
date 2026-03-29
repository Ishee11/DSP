package engine

import "github.com/Ishee11/DSP/internal/model"

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	return Decide(req, campaigns)
}

func Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {

	var best *model.Campaign

	for _, c := range campaigns {

		// 1. фильтрация
		if c.SiteID != req.SiteID {
			continue
		}

		if c.DeviceType != req.DeviceType {
			continue
		}

		if c.Price < req.FloorPrice {
			continue
		}

		// 2. выбор лучшего
		if best == nil || c.Price > best.Price {
			tmp := c
			best = &tmp
		}
	}

	// 3. no-bid
	if best == nil {
		return nil, false
	}

	// 4. ответ
	return &model.BidResponse{
		RequestID: req.RequestID,
		ImpID:     req.ImpID,
		Price:     best.Price,
		AdID:      best.ID,
	}, true
}
