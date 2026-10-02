// Package appointment defines the Appointment domain — the core business entity
// of the booking engine. An appointment connects a customer, service, staff,
// location, and optional resource at a specific time.
package appointment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of an appointment.
// Status transitions are controlled by domain rules — not arbitrary updates.
type Status string

const (
	StatusPending     Status = "pending"     // Created, awaiting confirmation
	StatusConfirmed   Status = "confirmed"   // Confirmed (manually or auto)
	StatusRescheduled Status = "rescheduled" // Rescheduled from a previous time
	StatusCompleted   Status = "completed"   // Service delivered
	StatusCancelled   Status = "cancelled"   // Cancelled by customer or staff
	StatusNoShow      Status = "no_show"     // Customer did not show up
)

// ValidTransitions defines allowed status transitions for appointment lifecycle.
// Key: current status, Value: allowed next statuses.
var ValidTransitions = map[Status][]Status{
	StatusPending:     {StatusConfirmed, StatusCancelled},
	StatusConfirmed:   {StatusRescheduled, StatusCompleted, StatusCancelled, StatusNoShow},
	StatusRescheduled: {StatusConfirmed, StatusCancelled},
	StatusCompleted:   {}, // terminal state
	StatusCancelled:   {}, // terminal state
	StatusNoShow:      {}, // terminal state
}

// CanTransition checks if transitioning from current to next status is valid.
func CanTransition(current, next Status) bool {
	allowed, ok := ValidTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// Appointment is the central booking entity.
// It connects all booking entities: customer, service, staff, location, resource.
type Appointment struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	LocationID     uuid.UUID  `json:"location_id"`
	ServiceID      uuid.UUID  `json:"service_id"`
	StaffID        uuid.UUID  `json:"staff_id"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty"` // nil for guest bookings
	ResourceID     *uuid.UUID `json:"resource_id,omitempty"` // optional resource

	// Timing (always stored in UTC)
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timezone  string    `json:"timezone"` // IANA timezone the customer booked in

	// Status
	Status Status `json:"status"`

	// Snapshot of service price at booking time (price may change later)
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`

	// Guest booking info (used when customer_id is nil)
	GuestName  string `json:"guest_name,omitempty"`
	GuestEmail string `json:"guest_email,omitempty"`
	GuestPhone string `json:"guest_phone,omitempty"`

	// Booking metadata
	Notes         string `json:"notes,omitempty"`   // customer notes
	InternalNotes string `json:"internal_notes,omitempty"` // staff/admin notes
	CancelReason  string `json:"cancel_reason,omitempty"`
	Source        string `json:"source"` // "online", "manual", "api"

	// References
	PaymentID *uuid.UUID `json:"payment_id,omitempty"`

	CreatedBy uuid.UUID `json:"created_by"` // user who created the booking
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Duration returns the appointment duration.
func (a *Appointment) Duration() time.Duration {
	return a.EndTime.Sub(a.StartTime)
}

// IsGuestBooking returns true if the appointment was made without a customer account.
func (a *Appointment) IsGuestBooking() bool {
	return a.CustomerID == nil
}

// StatusHistory records every status change for audit purposes.
type StatusHistory struct {
	ID            uuid.UUID `json:"id"`
	AppointmentID uuid.UUID `json:"appointment_id"`
	FromStatus    Status    `json:"from_status"`
	ToStatus      Status    `json:"to_status"`
	Reason        string    `json:"reason,omitempty"`
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedAt     time.Time `json:"changed_at"`
}

// --- Commands ----------------------------------------------------------------

// CreateAppointmentCmd holds data for creating a new appointment.
type CreateAppointmentCmd struct {
	OrganizationID uuid.UUID
	LocationID     uuid.UUID
	ServiceID      uuid.UUID
	StaffID        uuid.UUID
	CustomerID     *uuid.UUID
	ResourceID     *uuid.UUID
	StartTime      time.Time
	EndTime        time.Time
	Timezone       string
	PriceCents     int64
	Currency       string
	ServiceName    string
	GuestName      string
	GuestEmail     string
	GuestPhone     string
	Notes          string
	Source         string
	CreatedBy      uuid.UUID
}

// RescheduleCmd holds data for rescheduling an appointment.
type RescheduleCmd struct {
	AppointmentID uuid.UUID
	NewStartTime  time.Time
	Reason        string
	ChangedBy     uuid.UUID
}

// UpdateStatusCmd holds data for updating appointment status.
type UpdateStatusCmd struct {
	AppointmentID uuid.UUID
	NewStatus     Status
	Reason        string
	ChangedBy     uuid.UUID
}

// UpdateNotesCmd holds data for updating appointment notes.
type UpdateNotesCmd struct {
	AppointmentID uuid.UUID
	Notes         *string
	InternalNotes *string
	UpdatedBy     uuid.UUID
}

// --- Queries / Filters -------------------------------------------------------

// ListFilter holds filter criteria for listing appointments.
type ListFilter struct {
	OrganizationID uuid.UUID
	LocationID     *uuid.UUID
	StaffID        *uuid.UUID
	CustomerID     *uuid.UUID
	Status         *Status
	ServiceID      *uuid.UUID
	DateFrom       *time.Time
	DateTo         *time.Time
	Search         string // search on guest name/email or customer name
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// Create creates a new appointment within a database transaction.
	// The implementation MUST use SELECT FOR UPDATE or equivalent locking
	// to prevent double-booking the same staff/time slot.
	Create(ctx context.Context, cmd CreateAppointmentCmd) (*Appointment, error)

	// GetByID retrieves an appointment by ID, scoped to organization.
	GetByID(ctx context.Context, orgID, appointmentID uuid.UUID) (*Appointment, error)

	// List returns appointments matching the filter with pagination.
	List(ctx context.Context, filter ListFilter, page, perPage int) ([]*Appointment, int, error)

	// ListForStaff returns appointments for a specific staff member in a time range.
	ListForStaff(ctx context.Context, orgID, staffID uuid.UUID, from, to time.Time) ([]*Appointment, error)

	// ListForCalendar returns all non-cancelled appointments in a date range for calendar view.
	ListForCalendar(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]*Appointment, error)

	// UpdateStatus changes the appointment status and records history.
	UpdateStatus(ctx context.Context, cmd UpdateStatusCmd) error

	// Reschedule updates the appointment time and sets status to rescheduled.
	Reschedule(ctx context.Context, cmd RescheduleCmd) (*Appointment, error)

	// UpdateNotes updates internal or customer-facing notes.
	UpdateNotes(ctx context.Context, cmd UpdateNotesCmd) error

	// GetStatusHistory returns the full status history for an appointment.
	GetStatusHistory(ctx context.Context, appointmentID uuid.UUID) ([]*StatusHistory, error)

	// CountByOrganization counts appointments for usage/billing tracking.
	CountByOrganization(ctx context.Context, orgID uuid.UUID, from, to time.Time) (int, error)

	// FindConflicts returns existing appointments that overlap with the proposed time slot.
	// Used by the availability engine to detect conflicts.
	FindConflicts(ctx context.Context, orgID, staffID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) ([]*Appointment, error)

	// FindResourceConflicts finds conflicts for a specific resource.
	FindResourceConflicts(ctx context.Context, orgID, resourceID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) ([]*Appointment, error)
}

// --- Service Interface --------------------------------------------------------

type Service interface {
	CreateAppointment(ctx context.Context, cmd CreateAppointmentCmd) (*Appointment, error)
	GetAppointmentByID(ctx context.Context, orgID, apptID uuid.UUID) (*Appointment, error)
	ListAppointments(ctx context.Context, filter ListFilter, page, perPage int) ([]*Appointment, int, error)
	RescheduleAppointment(ctx context.Context, cmd RescheduleCmd) (*Appointment, error)
	CancelAppointment(ctx context.Context, orgID, apptID uuid.UUID, reason string, cancelledBy uuid.UUID) error
	UpdateAppointmentStatus(ctx context.Context, cmd UpdateStatusCmd) error
}

