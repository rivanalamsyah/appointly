package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// OrgRepository is a thread-safe in-memory repository for organization domain.
type OrgRepository struct {
	mu        sync.RWMutex
	orgs      map[uuid.UUID]*organization.Organization
	slugIndex map[string]uuid.UUID
}

func NewOrgRepository() *OrgRepository {
	return &OrgRepository{
		orgs:      make(map[uuid.UUID]*organization.Organization),
		slugIndex: make(map[string]uuid.UUID),
	}
}

func (r *OrgRepository) Create(ctx context.Context, org *organization.Organization) (*organization.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slugKey := strings.ToLower(strings.TrimSpace(org.Slug))
	if _, exists := r.slugIndex[slugKey]; exists {
		return nil, apperror.AlreadyExists("organization with slug " + org.Slug)
	}

	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now
	if org.Status == "" {
		org.Status = organization.StatusActive
	}

	cp := *org
	r.orgs[org.ID] = &cp
	r.slugIndex[slugKey] = org.ID

	return &cp, nil
}

func (r *OrgRepository) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	org, exists := r.orgs[id]
	if !exists {
		return nil, apperror.NotFound("organization")
	}
	cp := *org
	return &cp, nil
}

func (r *OrgRepository) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slugKey := strings.ToLower(strings.TrimSpace(slug))
	id, exists := r.slugIndex[slugKey]
	if !exists {
		return nil, apperror.NotFound("organization")
	}
	org := r.orgs[id]
	cp := *org
	return &cp, nil
}

func (r *OrgRepository) Update(ctx context.Context, org *organization.Organization) (*organization.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.orgs[org.ID]
	if !exists {
		return nil, apperror.NotFound("organization")
	}

	org.UpdatedAt = time.Now()
	org.CreatedAt = existing.CreatedAt
	cp := *org
	r.orgs[org.ID] = &cp
	return &cp, nil
}

func (r *OrgRepository) UpdateSettings(ctx context.Context, orgID uuid.UUID, settings organization.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, exists := r.orgs[orgID]
	if !exists {
		return apperror.NotFound("organization")
	}
	org.Settings = settings
	org.UpdatedAt = time.Now()
	return nil
}

func (r *OrgRepository) UpdateStatus(ctx context.Context, orgID uuid.UUID, status organization.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, exists := r.orgs[orgID]
	if !exists {
		return apperror.NotFound("organization")
	}
	org.Status = status
	org.UpdatedAt = time.Now()
	return nil
}

func (r *OrgRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slugKey := strings.ToLower(strings.TrimSpace(slug))
	_, exists := r.slugIndex[slugKey]
	return exists, nil
}

func (r *OrgRepository) List(ctx context.Context, filter organization.ListOrganizationsFilter, page, perPage int) ([]*organization.Organization, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*organization.Organization
	for _, org := range r.orgs {
		if filter.Status != nil && org.Status != *filter.Status {
			continue
		}
		if filter.BusinessType != nil && org.BusinessType != *filter.BusinessType {
			continue
		}
		if filter.Search != "" {
			search := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(org.Name), search) && !strings.Contains(strings.ToLower(org.Slug), search) {
				continue
			}
		}
		cp := *org
		result = append(result, &cp)
	}

	total := len(result)
	start := (page - 1) * perPage
	if start >= total {
		return []*organization.Organization{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return result[start:end], total, nil
}

func (r *OrgRepository) Delete(ctx context.Context, orgID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, exists := r.orgs[orgID]
	if !exists {
		return apperror.NotFound("organization")
	}
	delete(r.slugIndex, strings.ToLower(org.Slug))
	delete(r.orgs, orgID)
	return nil
}
