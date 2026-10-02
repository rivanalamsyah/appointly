package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type ResourceRepository struct {
	mu               sync.RWMutex
	resources        map[uuid.UUID]*resource.Resource
	serviceResources map[uuid.UUID][]uuid.UUID // serviceID -> []resourceID
}

func NewResourceRepository() *ResourceRepository {
	return &ResourceRepository{
		resources:        make(map[uuid.UUID]*resource.Resource),
		serviceResources: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *ResourceRepository) CreateResource(ctx context.Context, cmd resource.CreateResourceCmd) (*resource.Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, apperror.ValidationFailed("Resource name is required")
	}

	cap := cmd.Capacity
	if cap < 1 {
		cap = 1
	}

	resType := cmd.Type
	if resType == "" {
		resType = resource.TypeOther
	}

	status := cmd.Status
	if status == "" {
		status = resource.StatusActive
	}

	now := time.Now()
	resID := uuid.New()

	res := &resource.Resource{
		ID:             resID,
		OrganizationID: cmd.OrganizationID,
		LocationID:     cmd.LocationID,
		Name:           name,
		Type:           resType,
		Capacity:       cap,
		Status:         status,
		Description:    strings.TrimSpace(cmd.Description),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.resources[resID] = res
	cp := *res
	return &cp, nil
}

func (r *ResourceRepository) GetResourceByID(ctx context.Context, orgID, resourceID uuid.UUID) (*resource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, exists := r.resources[resourceID]
	if !exists || res.OrganizationID != orgID {
		return nil, apperror.NotFound("resource")
	}

	cp := *res
	return &cp, nil
}

func (r *ResourceRepository) UpdateResource(ctx context.Context, orgID uuid.UUID, cmd resource.UpdateResourceCmd) (*resource.Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resources[cmd.ID]
	if !exists || res.OrganizationID != orgID {
		return nil, apperror.NotFound("resource")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, apperror.ValidationFailed("Resource name is required")
	}

	cap := cmd.Capacity
	if cap < 1 {
		cap = 1
	}

	resType := cmd.Type
	if resType == "" {
		resType = resource.TypeOther
	}

	status := cmd.Status
	if status == "" {
		status = resource.StatusActive
	}

	res.LocationID = cmd.LocationID
	res.Name = name
	res.Type = resType
	res.Capacity = cap
	res.Status = status
	res.Description = strings.TrimSpace(cmd.Description)
	res.UpdatedAt = time.Now()

	cp := *res
	return &cp, nil
}

func (r *ResourceRepository) DeleteResource(ctx context.Context, orgID, resourceID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resources[resourceID]
	if !exists || res.OrganizationID != orgID {
		return apperror.NotFound("resource")
	}

	delete(r.resources, resourceID)

	// Clean up service resource assignments
	for svcID, resIDs := range r.serviceResources {
		updated := make([]uuid.UUID, 0, len(resIDs))
		for _, id := range resIDs {
			if id != resourceID {
				updated = append(updated, id)
			}
		}
		r.serviceResources[svcID] = updated
	}

	return nil
}

func (r *ResourceRepository) ListResources(ctx context.Context, filter resource.ListResourcesFilter) ([]*resource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*resource.Resource
	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, res := range r.resources {
		if res.OrganizationID != filter.OrganizationID {
			continue
		}
		if filter.LocationID != nil && (res.LocationID == nil || *res.LocationID != *filter.LocationID) {
			continue
		}
		if filter.Type != nil && res.Type != *filter.Type {
			continue
		}
		if filter.Status != nil && res.Status != *filter.Status {
			continue
		}
		if searchLower != "" {
			if !strings.Contains(strings.ToLower(res.Name), searchLower) &&
				!strings.Contains(strings.ToLower(res.Description), searchLower) {
				continue
			}
		}

		cp := *res
		result = append(result, &cp)
	}

	return result, nil
}

func (r *ResourceRepository) AssignServiceResources(ctx context.Context, cmd resource.AssignServiceResourcesCmd) ([]*resource.Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify all target resourceIDs belong to orgID
	var assigned []*resource.Resource
	validIDs := make([]uuid.UUID, 0, len(cmd.ResourceIDs))

	for _, id := range cmd.ResourceIDs {
		res, exists := r.resources[id]
		if !exists || res.OrganizationID != cmd.OrganizationID {
			return nil, apperror.ValidationFailed("One or more resources do not exist in this organization")
		}
		validIDs = append(validIDs, id)
		cp := *res
		assigned = append(assigned, &cp)
	}

	r.serviceResources[cmd.ServiceID] = validIDs
	return assigned, nil
}

func (r *ResourceRepository) GetServiceResources(ctx context.Context, orgID, serviceID uuid.UUID) ([]*resource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resIDs, exists := r.serviceResources[serviceID]
	if !exists {
		return []*resource.Resource{}, nil
	}

	var result []*resource.Resource
	for _, id := range resIDs {
		res, exists := r.resources[id]
		if exists && res.OrganizationID == orgID {
			cp := *res
			result = append(result, &cp)
		}
	}

	return result, nil
}

// AvailabilityChecker Interface Implementation (Contract for Availability Engine)
func (r *ResourceRepository) IsResourceAvailable(ctx context.Context, orgID, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, exists := r.resources[resourceID]
	if !exists || res.OrganizationID != orgID {
		return false, apperror.NotFound("resource")
	}

	// Active status check
	if res.Status != resource.StatusActive || res.Capacity <= 0 {
		return false, nil
	}

	// Booking collision detection will be evaluated by Phase 7 Availability Engine against appointments.
	return true, nil
}

func (r *ResourceRepository) GetAvailableResourcesForService(ctx context.Context, orgID, serviceID uuid.UUID, locationID *uuid.UUID, startTime, endTime time.Time) ([]*resource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resIDs, exists := r.serviceResources[serviceID]
	if !exists || len(resIDs) == 0 {
		// Service doesn't require any specific resource
		return []*resource.Resource{}, nil
	}

	var available []*resource.Resource
	for _, id := range resIDs {
		res, ok := r.resources[id]
		if !ok || res.OrganizationID != orgID {
			continue
		}
		if locationID != nil && res.LocationID != nil && *res.LocationID != *locationID {
			continue
		}
		if res.Status == resource.StatusActive && res.Capacity > 0 {
			cp := *res
			available = append(available, &cp)
		}
	}

	return available, nil
}
