// Package location defines the Location domain — physical locations or branches
// where services are delivered. An organization can have multiple locations.
package location

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the operational status of a location.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// Location represents a physical location where services are offered.
// This could be a single-location salon or one of multiple clinic branches.
type Location struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug,omitempty"` // optional per-location URL slug
	Description    string    `json:"description,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`

	// Address
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country"` // ISO 3166-1 alpha-2

	// Coordinates for map display
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	// Timezone for this location (inherits from org if not set)
	Timezone string `json:"timezone,omitempty"`

	Status       Status `json:"status"`
	IsDefault    bool   `json:"is_default"` // default location for single-location orgs
	DisplayOrder int    `json:"display_order"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Commands ----------------------------------------------------------------

type CreateLocationCmd struct {
	OrganizationID uuid.UUID
	Name           string
	Description    string
	Phone          string
	Email          string
	AddressLine1   string
	AddressLine2   string
	City           string
	State          string
	PostalCode     string
	Country        string
	Latitude       *float64
	Longitude      *float64
	Timezone       string
	IsDefault      bool
}

type UpdateLocationCmd struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Name           *string
	Description    *string
	Phone          *string
	Email          *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	State          *string
	PostalCode     *string
	Latitude       *float64
	Longitude      *float64
	Timezone       *string
	Status         *Status
	IsDefault      *bool
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	Create(ctx context.Context, cmd CreateLocationCmd) (*Location, error)
	GetByID(ctx context.Context, orgID, locationID uuid.UUID) (*Location, error)
	GetDefault(ctx context.Context, orgID uuid.UUID) (*Location, error)
	List(ctx context.Context, orgID uuid.UUID) ([]*Location, error)
	ListActive(ctx context.Context, orgID uuid.UUID) ([]*Location, error)
	Update(ctx context.Context, cmd UpdateLocationCmd) (*Location, error)
	SetDefault(ctx context.Context, orgID, locationID uuid.UUID) error
	Delete(ctx context.Context, orgID, locationID uuid.UUID) error
	CountActive(ctx context.Context, orgID uuid.UUID) (int, error)
}
