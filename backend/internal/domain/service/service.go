// Package service defines the Service domain — the offerings that an organization
// provides to customers. Services are vertical-agnostic: haircut, consultation,
// legal advice, tutoring session, etc., are all represented as Service entities.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents whether a service is available for booking.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusArchived Status = "archived"
)

// ServiceCategory groups related services (e.g., "Haircut", "Color Treatment").
type ServiceCategory struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Color          string    `json:"color,omitempty"` // hex color for calendar display
	DisplayOrder   int       `json:"display_order"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Service represents a bookable offering.
// Price is stored in the smallest currency unit (cents/IDR/etc.) to avoid
// floating point precision issues.
type Service struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      Status `json:"status"`

	// Timing (in minutes)
	DurationMinutes int `json:"duration_minutes"`
	BufferBefore    int `json:"buffer_before_minutes"` // prep time before appointment
	BufferAfter     int `json:"buffer_after_minutes"`  // cleanup time after appointment

	// Pricing
	PriceCents int64  `json:"price_cents"` // price in smallest currency unit
	Currency   string `json:"currency"`    // ISO 4217

	// Booking constraints
	MaxCapacity     int  `json:"max_capacity"`      // 1 for individual, >1 for group sessions
	RequiresResource bool `json:"requires_resource"` // if true, resource must be assigned

	// Display
	Color        string `json:"color,omitempty"`     // hex color for calendar
	ImageURL     string `json:"image_url,omitempty"` // service thumbnail
	DisplayOrder int    `json:"display_order"`
	IsPublic     bool   `json:"is_public"` // visible on public booking page

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PriceFormatted returns the price in major currency units (e.g., 15000 IDR → 15000.00).
// For display purposes only; use PriceCents for calculations.
func (s *Service) PriceFormatted() float64 {
	return float64(s.PriceCents) / 100.0
}

// TotalDurationMinutes returns the total blocked time including buffers.
func (s *Service) TotalDurationMinutes() int {
	return s.DurationMinutes + s.BufferBefore + s.BufferAfter
}

// --- Commands ----------------------------------------------------------------

type CreateServiceCmd struct {
	OrganizationID  uuid.UUID
	CategoryID      *uuid.UUID
	Name            string
	Description     string
	DurationMinutes int
	BufferBefore    int
	BufferAfter     int
	PriceCents      int64
	Currency        string
	MaxCapacity     int
	RequiresResource bool
	Color           string
	IsPublic        bool
}

type UpdateServiceCmd struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	CategoryID      *uuid.UUID
	Name            *string
	Description     *string
	DurationMinutes *int
	BufferBefore    *int
	BufferAfter     *int
	PriceCents      *int64
	MaxCapacity     *int
	RequiresResource *bool
	Color           *string
	ImageURL        *string
	IsPublic        *bool
	DisplayOrder    *int
	Status          *Status
}

type CreateCategoryCmd struct {
	OrganizationID uuid.UUID
	Name           string
	Description    string
	Color          string
}

// --- Queries -----------------------------------------------------------------

type ListServicesFilter struct {
	OrganizationID uuid.UUID
	CategoryID     *uuid.UUID
	StaffID        *uuid.UUID
	Status         *Status
	IsPublic       *bool
	Search         string
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// Services
	CreateService(ctx context.Context, cmd CreateServiceCmd) (*Service, error)
	GetServiceByID(ctx context.Context, orgID, serviceID uuid.UUID) (*Service, error)
	ListServices(ctx context.Context, filter ListServicesFilter, page, perPage int) ([]*Service, int, error)
	UpdateService(ctx context.Context, cmd UpdateServiceCmd) (*Service, error)
	DeleteService(ctx context.Context, orgID, serviceID uuid.UUID) error

	// Categories
	CreateCategory(ctx context.Context, cmd CreateCategoryCmd) (*ServiceCategory, error)
	GetCategoryByID(ctx context.Context, orgID, categoryID uuid.UUID) (*ServiceCategory, error)
	ListCategories(ctx context.Context, orgID uuid.UUID) ([]*ServiceCategory, error)
	UpdateCategory(ctx context.Context, cat *ServiceCategory) (*ServiceCategory, error)
	DeleteCategory(ctx context.Context, orgID, categoryID uuid.UUID) error

	// Public listing (for booking page — only active/public services)
	ListPublicServices(ctx context.Context, orgID uuid.UUID) ([]*Service, error)
	ListPublicCategories(ctx context.Context, orgID uuid.UUID) ([]*ServiceCategory, error)
}
