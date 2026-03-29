package engine

import (
	"math"

	"github.com/Ishee11/DSP/internal/model"
)

const scoreEpsilon = 1e-9

// Engine coordinates campaign filtering and scoring.
// Engine координирует фильтрацию кампаний и их скоринг.
type Engine struct {
	filter TargetingFilter
	scorer Scorer
}

// New creates an engine with the default in-process filter and scorer.
// New создаёт движок со стандартными in-process фильтром и скорером.
func New() *Engine {
	return &Engine{
		filter: StaticFilter{},
		scorer: PriceScorer{},
	}
}

// NewWithFilter creates an engine with a custom filter and the default scorer.
// NewWithFilter создаёт движок с пользовательским фильтром и стандартным скорером.
func NewWithFilter(filter TargetingFilter) *Engine {
	return NewWithDeps(filter, nil)
}

// NewWithScorer creates an engine with the default filter and a custom scorer.
// NewWithScorer создаёт движок со стандартным фильтром и пользовательским скорером.
func NewWithScorer(scorer Scorer) *Engine {
	return NewWithDeps(nil, scorer)
}

// NewWithDeps wires a fully configurable engine and falls back to defaults for nil dependencies.
// NewWithDeps собирает настраиваемый движок и подставляет зависимости по умолчанию для nil.
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

// Decide filters campaigns, scores remaining candidates, and returns the winning bid response.
// Decide фильтрует кампании, оценивает оставшихся кандидатов и возвращает ответ с победившей ставкой.
func (e *Engine) Decide(
	req model.BidRequest,
	campaigns []model.Campaign,
) (*model.BidResponse, bool) {
	candidates := e.filter.Filter(req, campaigns)
	return decideFromCandidates(req, candidates, e.scorer)
}

// decideFromCandidates assumes campaigns are already eligible and picks the best one using scorer.
// decideFromCandidates предполагает, что кампании уже прошли eligibility-проверки, и выбирает лучшую через scorer.
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
		if best == nil || isBetterCandidate(c, score, *best, bestScore) {
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

// isBetterCandidate compares two campaigns deterministically:
// first by score, then by price, then by ID for stable tie-breaking.
// isBetterCandidate сравнивает две кампании детерминированно:
// сначала по score, затем по цене, затем по ID для стабильного tie-break.
func isBetterCandidate(candidate model.Campaign, candidateScore float64, best model.Campaign, bestScore float64) bool {
	if candidateScore-bestScore > scoreEpsilon {
		return true
	}
	if math.Abs(candidateScore-bestScore) > scoreEpsilon {
		return false
	}

	if candidate.Price > best.Price {
		return true
	}
	if candidate.Price < best.Price {
		return false
	}

	return candidate.ID < best.ID
}
