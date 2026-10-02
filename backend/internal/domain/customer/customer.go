// Package customer defines the Customer CRM domain.
// Customers are tenant business entities representing people who book appointments.
// They are distinct from Users (who have system auth accounts). A customer may
// optionally be linked to a user account via explicit identity linking.
package customer

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CustomerStatus string

const (
	StatusActive   CustomerStatus = "ACTIVE"
	StatusInactive CustomerStatus = "INACTIVE"
	StatusBlocked  CustomerStatus = "BLOCKED"
)

type CustomerSource string

const (
	SourceWalkIn        CustomerSource = "WALK_IN"
	SourceOnlineBooking CustomerSource = "ONLINE_BOOKING"
	SourceManual        CustomerSource = "MANUAL"
	SourceImport        CustomerSource = "IMPORT"
	SourceOther         CustomerSource = "OTHER"
)

// Customer represents a CRM customer entity scoped to an organization.
type Customer struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	UserID         *uuid.UUID     `json:"user_id,omitempty"` // linked account if registered
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	Notes          string         `json:"notes,omitempty"` // general CRM notes summary
	Status         CustomerStatus `json:"status"`
	Source         CustomerSource `json:"source"`
	Tags           []string       `json:"tags,omitempty"`

	// Denormalized aggregate statistics
	TotalAppointments     int        `json:"total_appointments"`
	CompletedAppointments int        `json:"completed_appointments"`
	NoShowCount           int        `json:"no_show_count"`
	LastAppointmentAt     *time.Time `json:"last_appointment_at,omitempty"`
	TotalSpentCents       int64      `json:"total_spent_cents"`

	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (c *Customer) FullName() string {
	if c.LastName == "" {
		return c.FirstName
	}
	return c.FirstName + " " + c.LastName
}

func (c *Customer) IsDeleted() bool {
	return c.DeletedAt != nil
}

// CustomerNote represents an activity/note entry on the customer's timeline.
type CustomerNote struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	AuthorID       uuid.UUID `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

// --- Commands & Queries ------------------------------------------------------

type CreateCustomerCmd struct {
	OrganizationID uuid.UUID      `json:"organization_id"`
	UserID         *uuid.UUID     `json:"user_id,omitempty"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	Notes          string         `json:"notes,omitempty"`
	Status         CustomerStatus `json:"status,omitempty"`
	Source         CustomerSource `json:"source,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
}

type UpdateCustomerCmd struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	FirstName      *string         `json:"first_name,omitempty"`
	LastName       *string         `json:"last_name,omitempty"`
	Email          *string         `json:"email,omitempty"`
	Phone          *string         `json:"phone,omitempty"`
	Notes          *string         `json:"notes,omitempty"`
	Status         *CustomerStatus `json:"status,omitempty"`
	Source         *CustomerSource `json:"source,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
}

type LinkUserCmd struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	UserID         uuid.UUID `json:"user_id"`
}

type AddNoteCmd struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	AuthorID       uuid.UUID `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	Content        string    `json:"content"`
}

type ListCustomersFilter struct {
	OrganizationID uuid.UUID       `json:"organization_id"`
	Search         string          `json:"search,omitempty"` // name, email, phone
	Status         *CustomerStatus `json:"status,omitempty"`
	Source         *CustomerSource `json:"source,omitempty"`
	IncludeDeleted bool            `json:"include_deleted,omitempty"`
	SortBy         string          `json:"sort_by,omitempty"` // created_at, name, total_appointments
	SortOrder      string          `json:"sort_order,omitempty"` // asc, desc
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	Create(ctx context.Context, cmd CreateCustomerCmd) (*Customer, error)
	GetByID(ctx context.Context, orgID, customerID uuid.UUID) (*Customer, error)
	GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (*Customer, error)
	GetByPhone(ctx context.Context, orgID uuid.UUID, phone string) (*Customer, error)
	GetByUserID(ctx context.Context, orgID, userID uuid.UUID) (*Customer, error)
	List(ctx context.Context, filter ListCustomersFilter, page, perPage int) ([]*Customer, int, error)
	Update(ctx context.Context, cmd UpdateCustomerCmd) (*Customer, error)
	SoftDelete(ctx context.Context, orgID, customerID uuid.UUID) error
	LinkUser(ctx context.Context, cmd LinkUserCmd) (*Customer, error)
	AddNote(ctx context.Context, cmd AddNoteCmd) (*CustomerNote, error)
	ListNotes(ctx context.Context, orgID, customerID uuid.UUID) ([]*CustomerNote, error)
	FindOrCreate(ctx context.Context, cmd CreateCustomerCmd) (*Customer, bool, error)
}
