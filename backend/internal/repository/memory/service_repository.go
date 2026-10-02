package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type ServiceRepository struct {
	mu         sync.RWMutex
	services   map[uuid.UUID]*service.Service
	categories map[uuid.UUID]*service.ServiceCategory
}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{
		services:   make(map[uuid.UUID]*service.Service),
		categories: make(map[uuid.UUID]*service.ServiceCategory),
	}
}

func (r *ServiceRepository) CreateService(ctx context.Context, cmd service.CreateServiceCmd) (*service.Service, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	svcID := uuid.New()
	now := time.Now()

	maxCap := cmd.MaxCapacity
	if maxCap < 1 {
		maxCap = 1
	}

	curr := cmd.Currency
	if curr == "" {
		curr = "USD"
	}

	svc := &service.Service{
		ID:               svcID,
		OrganizationID:   cmd.OrganizationID,
		CategoryID:       cmd.CategoryID,
		Name:             strings.TrimSpace(cmd.Name),
		Description:      strings.TrimSpace(cmd.Description),
		Status:           service.StatusActive,
		DurationMinutes:  cmd.DurationMinutes,
		BufferBefore:     cmd.BufferBefore,
		BufferAfter:      cmd.BufferAfter,
		PriceCents:       cmd.PriceCents,
		Currency:         curr,
		MaxCapacity:      maxCap,
		RequiresResource: cmd.RequiresResource,
		Color:            cmd.Color,
		IsPublic:         cmd.IsPublic,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	r.services[svcID] = svc
	cp := *svc
	return &cp, nil
}

func (r *ServiceRepository) GetServiceByID(ctx context.Context, orgID, serviceID uuid.UUID) (*service.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[serviceID]
	if !exists || svc.OrganizationID != orgID {
		return nil, apperror.NotFound("service")
	}
	cp := *svc
	return &cp, nil
}

func (r *ServiceRepository) ListServices(ctx context.Context, filter service.ListServicesFilter, page, perPage int) ([]*service.Service, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*service.Service
	for _, s := range r.services {
		if s.OrganizationID != filter.OrganizationID {
			continue
		}
		if filter.CategoryID != nil && (s.CategoryID == nil || *s.CategoryID != *filter.CategoryID) {
			continue
		}
		if filter.Status != nil && s.Status != *filter.Status {
			continue
		}
		if filter.IsPublic != nil && s.IsPublic != *filter.IsPublic {
			continue
		}
		if filter.Search != "" {
			query := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(s.Name), query) && !strings.Contains(strings.ToLower(s.Description), query) {
				continue
			}
		}
		cp := *s
		matched = append(matched, &cp)
	}

	total := len(matched)
	start := (page - 1) * perPage
	if start >= total {
		return []*service.Service{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *ServiceRepository) UpdateService(ctx context.Context, cmd service.UpdateServiceCmd) (*service.Service, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.services[cmd.ID]
	if !exists || svc.OrganizationID != cmd.OrganizationID {
		return nil, apperror.NotFound("service")
	}

	if cmd.Name != nil {
		svc.Name = strings.TrimSpace(*cmd.Name)
	}
	if cmd.Description != nil {
		svc.Description = strings.TrimSpace(*cmd.Description)
	}
	if cmd.CategoryID != nil {
		svc.CategoryID = cmd.CategoryID
	}
	if cmd.DurationMinutes != nil {
		svc.DurationMinutes = *cmd.DurationMinutes
	}
	if cmd.BufferBefore != nil {
		svc.BufferBefore = *cmd.BufferBefore
	}
	if cmd.BufferAfter != nil {
		svc.BufferAfter = *cmd.BufferAfter
	}
	if cmd.PriceCents != nil {
		svc.PriceCents = *cmd.PriceCents
	}
	if cmd.MaxCapacity != nil {
		svc.MaxCapacity = *cmd.MaxCapacity
	}
	if cmd.RequiresResource != nil {
		svc.RequiresResource = *cmd.RequiresResource
	}
	if cmd.Color != nil {
		svc.Color = *cmd.Color
	}
	if cmd.ImageURL != nil {
		svc.ImageURL = *cmd.ImageURL
	}
	if cmd.IsPublic != nil {
		svc.IsPublic = *cmd.IsPublic
	}
	if cmd.DisplayOrder != nil {
		svc.DisplayOrder = *cmd.DisplayOrder
	}
	if cmd.Status != nil {
		svc.Status = *cmd.Status
	}

	svc.UpdatedAt = time.Now()
	cp := *svc
	return &cp, nil
}

func (r *ServiceRepository) DeleteService(ctx context.Context, orgID, serviceID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.services[serviceID]
	if !exists || svc.OrganizationID != orgID {
		return apperror.NotFound("service")
	}
	// Soft delete / archive status so historical appointments stay valid
	svc.Status = service.StatusArchived
	svc.UpdatedAt = time.Now()
	return nil
}

func (r *ServiceRepository) CreateCategory(ctx context.Context, cmd service.CreateCategoryCmd) (*service.ServiceCategory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	catID := uuid.New()
	now := time.Now()

	cat := &service.ServiceCategory{
		ID:             catID,
		OrganizationID: cmd.OrganizationID,
		Name:           strings.TrimSpace(cmd.Name),
		Description:    strings.TrimSpace(cmd.Description),
		Color:          cmd.Color,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.categories[catID] = cat
	cp := *cat
	return &cp, nil
}

func (r *ServiceRepository) GetCategoryByID(ctx context.Context, orgID, categoryID uuid.UUID) (*service.ServiceCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cat, exists := r.categories[categoryID]
	if !exists || cat.OrganizationID != orgID {
		return nil, apperror.NotFound("service category")
	}
	cp := *cat
	return &cp, nil
}

func (r *ServiceRepository) ListCategories(ctx context.Context, orgID uuid.UUID) ([]*service.ServiceCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*service.ServiceCategory
	for _, c := range r.categories {
		if c.OrganizationID == orgID && c.IsActive {
			cp := *c
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *ServiceRepository) UpdateCategory(ctx context.Context, cat *service.ServiceCategory) (*service.ServiceCategory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.categories[cat.ID]
	if !exists || existing.OrganizationID != cat.OrganizationID {
		return nil, apperror.NotFound("service category")
	}

	cat.UpdatedAt = time.Now()
	cat.CreatedAt = existing.CreatedAt
	cp := *cat
	r.categories[cat.ID] = &cp
	return &cp, nil
}

func (r *ServiceRepository) DeleteCategory(ctx context.Context, orgID, categoryID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cat, exists := r.categories[categoryID]
	if !exists || cat.OrganizationID != orgID {
		return apperror.NotFound("service category")
	}
	cat.IsActive = false
	cat.UpdatedAt = time.Now()
	return nil
}

func (r *ServiceRepository) ListPublicServices(ctx context.Context, orgID uuid.UUID) ([]*service.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*service.Service
	for _, s := range r.services {
		if s.OrganizationID == orgID && s.IsPublic && s.Status == service.StatusActive {
			cp := *s
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *ServiceRepository) ListPublicCategories(ctx context.Context, orgID uuid.UUID) ([]*service.ServiceCategory, error) {
	return r.ListCategories(ctx, orgID)
}
