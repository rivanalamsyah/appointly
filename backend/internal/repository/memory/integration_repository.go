package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type IntegrationRepository struct {
	mu           sync.RWMutex
	integrations map[uuid.UUID]*integration.Integration
	syncStates   map[uuid.UUID]*integration.SyncState
}

func NewIntegrationRepository() *IntegrationRepository {
	return &IntegrationRepository{
		integrations: make(map[uuid.UUID]*integration.Integration),
		syncStates:   make(map[uuid.UUID]*integration.SyncState),
	}
}

func (r *IntegrationRepository) Create(ctx context.Context, item *integration.Integration) (*integration.Integration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	cp := *item
	r.integrations[item.ID] = &cp
	return &cp, nil
}

func (r *IntegrationRepository) Update(ctx context.Context, item *integration.Integration) (*integration.Integration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.integrations[item.ID]
	if !exists {
		return nil, apperror.NotFound("integration setting")
	}

	item.UpdatedAt = time.Now()
	item.CreatedAt = existing.CreatedAt
	cp := *item
	r.integrations[item.ID] = &cp
	return &cp, nil
}

func (r *IntegrationRepository) GetByProvider(ctx context.Context, orgID uuid.UUID, provider integration.Provider) (*integration.Integration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.integrations {
		if item.OrganizationID == orgID && item.Provider == provider {
			cp := *item
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("integration provider for organization")
}

func (r *IntegrationRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*integration.Integration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*integration.Integration
	for _, item := range r.integrations {
		if item.OrganizationID == orgID {
			cp := *item
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *IntegrationRepository) Delete(ctx context.Context, orgID uuid.UUID, provider integration.Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, item := range r.integrations {
		if item.OrganizationID == orgID && item.Provider == provider {
			delete(r.integrations, id)
			return nil
		}
	}
	return apperror.NotFound("integration record to delete")
}

func (r *IntegrationRepository) SaveSyncState(ctx context.Context, state *integration.SyncState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if state.ID == uuid.Nil {
		state.ID = uuid.New()
	}
	state.CreatedAt = time.Now()
	cp := *state
	r.syncStates[state.ID] = &cp
	return nil
}

func (r *IntegrationRepository) GetSyncState(ctx context.Context, orgID, apptID uuid.UUID) (*integration.SyncState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.syncStates {
		if s.OrganizationID == orgID && s.AppointmentID == apptID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("sync state record")
}
