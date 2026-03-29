package engine

import "github.com/Ishee11/DSP/internal/model"

// TargetingFilter applies cheap eligibility checks and returns candidates
// that are allowed to continue in the decision pipeline.
// TargetingFilter выполняет дешёвые eligibility-проверки и возвращает кандидатов,
// которые могут продолжить движение по пайплайну принятия решения.
type TargetingFilter interface {
	Filter(req model.BidRequest, campaigns []model.Campaign) []model.Campaign
}

// StaticFilter is the current in-process filter implementation.
// It can later be replaced by a remote service without changing DecisionCore.
// StaticFilter — текущая in-process реализация фильтра.
// Позже её можно заменить на удалённый сервис без изменения decision core.
type StaticFilter struct{}

// Filter keeps only campaigns that match site, device type, and floor price.
// Filter оставляет только кампании, которые совпадают по сайту, типу устройства и floor price.
func (f StaticFilter) Filter(req model.BidRequest, campaigns []model.Campaign) []model.Campaign {
	filtered := make([]model.Campaign, 0, len(campaigns))

	for _, c := range campaigns {
		if c.SiteID != req.SiteID {
			continue
		}
		if c.DeviceType != req.DeviceType {
			continue
		}
		if c.Price < req.FloorPrice {
			continue
		}

		filtered = append(filtered, c)
	}

	return filtered
}
