// Package staff defines the Staff domain — the business entity representing
// a person who delivers services within an organization.
// Staff is distinct from User: a staff member may or may not have an account.
// Staff is also vertical-agnostic — it works for stylists, doctors, consultants, etc.
package staff

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Staff represents a service provider within an organization.
// Staff members can be linked to a user account (via UserID) or exist
// as standalone business entities for organizations managing staff without
// giving them system access.
type Staff struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	UserID         *uuid.UUID `json:"user_id,omitempty"` // optional link to a user account
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	Title          string     `json:"title,omitempty"`      // e.g., "Senior Stylist", "Dr."
	Bio            string     `json:"bio,omitempty"`
	IsActive       bool       `json:"is_active"`
	AcceptsOnline  bool       `json:"accepts_online"` // visible for online booking
	DisplayOrder   int        `json:"display_order"`  // sort order for public pages
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FullName returns the staff member's full name.
func (s *Staff) FullName() string {
	return s.FirstName + " " + s.LastName
}

// StaffLocation represents the assignment of a staff member to a location.
// A staff member can work at multiple locations.
type StaffLocation struct {
	StaffID    uuid.UUID `json:"staff_id"`
	LocationID uuid.UUID `json:"location_id"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
}

// StaffService represents the assignment of a staff member to a service.
// A staff member can provide multiple services.
type StaffService struct {
	StaffID        uuid.UUID  `json:"staff_id"`
	ServiceID      uuid.UUID  `json:"service_id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	// Override price/duration for this specific staff (nil = use service defaults)
	CustomPriceCents *int64  `json:"custom_price_cents,omitempty"`
	CustomDuration   *int    `json:"custom_duration_minutes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// --- Commands ----------------------------------------------------------------

type CreateStaffCmd struct {
	OrganizationID uuid.UUID
	UserID         *uuid.UUID
	FirstName      string
	LastName       string
	Email          string
	Phone          string
	Title          string
	Bio            string
	AcceptsOnline  bool
	LocationIDs    []uuid.UUID
	ServiceIDs     []uuid.UUID
}

type UpdateStaffCmd struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	FirstName      *string
	LastName       *string
	Email          *string
	Phone          *string
	AvatarURL      *string
	Title          *string
	Bio            *string
	AcceptsOnline  *bool
	IsActive       *bool
}

type AssignLocationsCmd struct {
	StaffID        uuid.UUID
	OrganizationID uuid.UUID
	LocationIDs    []uuid.UUID
}

type AssignServicesCmd struct {
	StaffID        uuid.UUID
	OrganizationID uuid.UUID
	ServiceIDs     []uuid.UUID
}

// --- Queries -----------------------------------------------------------------

type ListStaffFilter struct {
	OrganizationID uuid.UUID
	LocationID     *uuid.UUID
	ServiceID      *uuid.UUID
	IsActive       *bool
	AcceptsOnline  *bool
	Search         string
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	Create(ctx context.Context, cmd CreateStaffCmd) (*Staff, error)
	GetByID(ctx context.Context, orgID, staffID uuid.UUID) (*Staff, error)
	GetByUserID(ctx context.Context, orgID, userID uuid.UUID) (*Staff, error)
	List(ctx context.Context, filter ListStaffFilter, page, perPage int) ([]*Staff, int, error)
	Update(ctx context.Context, cmd UpdateStaffCmd) (*Staff, error)
	Delete(ctx context.Context, orgID, staffID uuid.UUID) error

	// Locations
	GetStaffLocations(ctx context.Context, staffID uuid.UUID) ([]*StaffLocation, error)
	SetStaffLocations(ctx context.Context, orgID, staffID uuid.UUID, locationIDs []uuid.UUID) error

	// Services
	GetStaffServices(ctx context.Context, staffID uuid.UUID) ([]*StaffService, error)
	SetStaffServices(ctx context.Context, orgID, staffID uuid.UUID, serviceIDs []uuid.UUID) error

	// For availability engine: find staff who can provide a service at a location
	FindAvailableStaff(ctx context.Context, orgID, serviceID, locationID uuid.UUID) ([]*Staff, error)

	// Count for plan limits
	CountActive(ctx context.Context, orgID uuid.UUID) (int, error)
}
