package storage

import (
	"context"
	"sync"

	"github.com/Ishee11/DSP/internal/model"
)

// CampaignRepository provides campaigns to the bidding pipeline.
// It keeps storage concerns outside of transport and engine layers.
type CampaignRepository interface {
	ListCampaigns(ctx context.Context) ([]model.Campaign, error)
}

// InMemoryCampaignRepository is the current storage implementation.
// It copies data on read/write so callers cannot mutate shared state.
type InMemoryCampaignRepository struct {
	mu        sync.RWMutex
	campaigns []model.Campaign
}

func NewInMemoryCampaignRepository(campaigns []model.Campaign) *InMemoryCampaignRepository {
	return &InMemoryCampaignRepository{
		campaigns: cloneCampaigns(campaigns),
	}
}

func (r *InMemoryCampaignRepository) ListCampaigns(_ context.Context) ([]model.Campaign, error) {
	if r == nil {
		return nil, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneCampaigns(r.campaigns), nil
}

func (r *InMemoryCampaignRepository) ReplaceCampaigns(campaigns []model.Campaign) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.campaigns = cloneCampaigns(campaigns)
}

func DemoCampaigns() []model.Campaign {
	return []model.Campaign{
		{ID: "c1", SiteID: "1", DeviceType: "mobile", Price: 1.2},
		{ID: "c2", SiteID: "1", DeviceType: "desktop", Price: 0.8},
		{ID: "c3", SiteID: "2", DeviceType: "mobile", Price: 1.5},
	}
}

func cloneCampaigns(campaigns []model.Campaign) []model.Campaign {
	if len(campaigns) == 0 {
		return nil
	}

	cloned := make([]model.Campaign, len(campaigns))
	copy(cloned, campaigns)
	return cloned
}
