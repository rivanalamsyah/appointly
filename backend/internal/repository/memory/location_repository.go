package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type LocationRepository struct {
	mu        sync.RWMutex
	locations map[uuid.UUID]*location.Location
}

func NewLocationRepository() *LocationRepository {
	return &LocationRepository{
		locations: make(map[uuid.UUID]*location.Location),
	}
}

func (r *LocationRepository) Create(ctx context.Context, cmd location.CreateLocationCmd) (*location.Location, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If this is set as default or is the first location for the org, set as default
	orgLocationCount := 0
	for _, l := range r.locations {
		if l.OrganizationID == cmd.OrganizationID {
			orgLocationCount++
			if cmd.IsDefault {
				l.IsDefault = false
			}
		}
	}

	isDefault := cmd.IsDefault || orgLocationCount == 0
	locID := uuid.New()
	now := time.Now()

	loc := &location.Location{
		ID:             locID,
		OrganizationID: cmd.OrganizationID,
		Name:           strings.TrimSpace(cmd.Name),
		Description:    strings.TrimSpace(cmd.Description),
		Phone:          strings.TrimSpace(cmd.Phone),
		Email:          strings.TrimSpace(cmd.Email),
		AddressLine1:   strings.TrimSpace(cmd.AddressLine1),
		AddressLine2:   strings.TrimSpace(cmd.AddressLine2),
		City:           strings.TrimSpace(cmd.City),
		State:          strings.TrimSpace(cmd.State),
		PostalCode:     strings.TrimSpace(cmd.PostalCode),
		Country:        strings.TrimSpace(cmd.Country),
		Latitude:       cmd.Latitude,
		Longitude:      cmd.Longitude,
		Timezone:       strings.TrimSpace(cmd.Timezone),
		Status:         location.StatusActive,
		IsDefault:      isDefault,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.locations[locID] = loc
	cp := *loc
	return &cp, nil
}

func (r *LocationRepository) GetByID(ctx context.Context, orgID, locationID uuid.UUID) (*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loc, exists := r.locations[locationID]
	if !exists || loc.OrganizationID != orgID {
		return nil, apperror.NotFound("location")
	}
	cp := *loc
	return &cp, nil
}

func (r *LocationRepository) GetDefault(ctx context.Context, orgID uuid.UUID) (*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, l := range r.locations {
		if l.OrganizationID == orgID && l.IsDefault {
			cp := *l
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("default location")
}

func (r *LocationRepository) List(ctx context.Context, orgID uuid.UUID) ([]*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*location.Location
	for _, l := range r.locations {
		if l.OrganizationID == orgID {
			cp := *l
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *LocationRepository) ListActive(ctx context.Context, orgID uuid.UUID) ([]*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*location.Location
	for _, l := range r.locations {
		if l.OrganizationID == orgID && l.Status == location.StatusActive {
			cp := *l
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *LocationRepository) Update(ctx context.Context, cmd location.UpdateLocationCmd) (*location.Location, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	loc, exists := r.locations[cmd.ID]
	if !exists || loc.OrganizationID != cmd.OrganizationID {
		return nil, apperror.NotFound("location")
	}

	if cmd.Name != nil {
		loc.Name = strings.TrimSpace(*cmd.Name)
	}
	if cmd.Description != nil {
		loc.Description = strings.TrimSpace(*cmd.Description)
	}
	if cmd.Phone != nil {
		loc.Phone = strings.TrimSpace(*cmd.Phone)
	}
	if cmd.Email != nil {
		loc.Email = strings.TrimSpace(*cmd.Email)
	}
	if cmd.AddressLine1 != nil {
		loc.AddressLine1 = strings.TrimSpace(*cmd.AddressLine1)
	}
	if cmd.AddressLine2 != nil {
		loc.AddressLine2 = strings.TrimSpace(*cmd.AddressLine2)
	}
	if cmd.City != nil {
		loc.City = strings.TrimSpace(*cmd.City)
	}
	if cmd.State != nil {
		loc.State = strings.TrimSpace(*cmd.State)
	}
	if cmd.PostalCode != nil {
		loc.PostalCode = strings.TrimSpace(*cmd.PostalCode)
	}
	if cmd.Timezone != nil {
		loc.Timezone = strings.TrimSpace(*cmd.Timezone)
	}
	if cmd.Status != nil {
		loc.Status = *cmd.Status
	}
	if cmd.Latitude != nil {
		loc.Latitude = cmd.Latitude
	}
	if cmd.Longitude != nil {
		loc.Longitude = cmd.Longitude
	}

	if cmd.IsDefault != nil && *cmd.IsDefault {
		for _, l := range r.locations {
			if l.OrganizationID == cmd.OrganizationID {
				l.IsDefault = false
			}
		}
		loc.IsDefault = true
	}

	loc.UpdatedAt = time.Now()
	cp := *loc
	return &cp, nil
}

func (r *LocationRepository) SetDefault(ctx context.Context, orgID, locationID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	target, exists := r.locations[locationID]
	if !exists || target.OrganizationID != orgID {
		return apperror.NotFound("location")
	}

	for _, l := range r.locations {
		if l.OrganizationID == orgID {
			l.IsDefault = (l.ID == locationID)
		}
	}
	return nil
}

func (r *LocationRepository) Delete(ctx context.Context, orgID, locationID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	loc, exists := r.locations[locationID]
	if !exists || loc.OrganizationID != orgID {
		return apperror.NotFound("location")
	}

	delete(r.locations, locationID)

	// If deleted location was default, set another location as default if any remain
	if loc.IsDefault {
		for _, l := range r.locations {
			if l.OrganizationID == orgID {
				l.IsDefault = true
				break
			}
		}
	}

	return nil
}

func (r *LocationRepository) CountActive(ctx context.Context, orgID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, l := range r.locations {
		if l.OrganizationID == orgID && l.Status == location.StatusActive {
			count++
		}
	}
	return count, nil
}
