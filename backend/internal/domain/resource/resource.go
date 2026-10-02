package resource

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	TypeRoom      ResourceType = "ROOM"
	TypeEquipment ResourceType = "EQUIPMENT"
	TypeChair     ResourceType = "CHAIR"
	TypeStudio    ResourceType = "STUDIO"
	TypeFacility  ResourceType = "FACILITY"
	TypeOther     ResourceType = "OTHER"
)

type ResourceStatus string

const (
	StatusActive      ResourceStatus = "ACTIVE"
	StatusInactive    ResourceStatus = "INACTIVE"
	StatusMaintenance ResourceStatus = "MAINTENANCE"
)

// Resource represents a physical or logical asset with limited capacity.
type Resource struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	LocationID     *uuid.UUID     `json:"location_id,omitempty"`
	Name           string         `json:"name"`
	Type           ResourceType   `json:"type"`
	Capacity       int            `json:"capacity"`
	Status         ResourceStatus `json:"status"`
	Description    string         `json:"description,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ServiceResource represents a requirement mapping between a Service and a Resource.
type ServiceResource struct {
	ID               uuid.UUID `json:"id"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	ServiceID        uuid.UUID `json:"service_id"`
	ResourceID       uuid.UUID `json:"resource_id"`
	QuantityRequired int       `json:"quantity_required"`
	CreatedAt        time.Time `json:"created_at"`
}

// DTOs & Commands
type CreateResourceCmd struct {
	OrganizationID uuid.UUID      `json:"organization_id"`
	LocationID     *uuid.UUID     `json:"location_id,omitempty"`
	Name           string         `json:"name"`
	Type           ResourceType   `json:"type"`
	Capacity       int            `json:"capacity"`
	Status         ResourceStatus `json:"status"`
	Description    string         `json:"description,omitempty"`
}

type UpdateResourceCmd struct {
	ID          uuid.UUID      `json:"id"`
	LocationID  *uuid.UUID     `json:"location_id,omitempty"`
	Name        string         `json:"name"`
	Type        ResourceType   `json:"type"`
	Capacity    int            `json:"capacity"`
	Status      ResourceStatus `json:"status"`
	Description string         `json:"description,omitempty"`
}

type ListResourcesFilter struct {
	OrganizationID uuid.UUID       `json:"organization_id"`
	LocationID     *uuid.UUID      `json:"location_id,omitempty"`
	Type           *ResourceType   `json:"type,omitempty"`
	Status         *ResourceStatus `json:"status,omitempty"`
	Search         string          `json:"search,omitempty"`
}

type AssignServiceResourcesCmd struct {
	OrganizationID uuid.UUID   `json:"organization_id"`
	ServiceID      uuid.UUID   `json:"service_id"`
	ResourceIDs    []uuid.UUID `json:"resource_ids"`
}

// Repository defines storage operations for resources and service-resource mappings.
type Repository interface {
	CreateResource(ctx context.Context, cmd CreateResourceCmd) (*Resource, error)
	GetResourceByID(ctx context.Context, orgID, resourceID uuid.UUID) (*Resource, error)
	UpdateResource(ctx context.Context, orgID uuid.UUID, cmd UpdateResourceCmd) (*Resource, error)
	DeleteResource(ctx context.Context, orgID, resourceID uuid.UUID) error
	ListResources(ctx context.Context, filter ListResourcesFilter) ([]*Resource, error)
	AssignServiceResources(ctx context.Context, cmd AssignServiceResourcesCmd) ([]*Resource, error)
	GetServiceResources(ctx context.Context, orgID, serviceID uuid.UUID) ([]*Resource, error)
}

// AvailabilityChecker defines the domain contract for the Availability Engine (Phase 7)
// to check whether required resources are available for a given time window without duplication.
type AvailabilityChecker interface {
	IsResourceAvailable(ctx context.Context, orgID, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error)
	GetAvailableResourcesForService(ctx context.Context, orgID, serviceID uuid.UUID, locationID *uuid.UUID, startTime, endTime time.Time) ([]*Resource, error)
}
