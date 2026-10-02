// Package organization defines the Organization domain — the root tenant entity
// for the Appointly multi-tenant SaaS platform. Every business record in the
// system is associated with an organization.
package organization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BusinessType indicates the vertical/industry of the organization.
// This is informational only — the core booking engine is vertical-agnostic.
type BusinessType string

const (
	BusinessTypeSalon       BusinessType = "salon"
	BusinessTypeClinic      BusinessType = "clinic"
	BusinessTypeConsultant  BusinessType = "consultant"
	BusinessTypeLawFirm     BusinessType = "law_firm"
	BusinessTypeTutor       BusinessType = "tutor"
	BusinessTypeGym         BusinessType = "gym"
	BusinessTypeOther       BusinessType = "other"
)

// Status represents the operational status of an organization.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusInactive  Status = "inactive"
)

// Organization is the root tenant entity. All business data belongs to an
// Organization and must be isolated by organization_id in every query.
type Organization struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Slug         string       `json:"slug"`          // URL-safe unique identifier for public pages
	BusinessType BusinessType `json:"business_type"`
	Email        string       `json:"email"`
	Phone        string       `json:"phone,omitempty"`
	Website      string       `json:"website,omitempty"`
	LogoURL      string       `json:"logo_url,omitempty"`
	Description  string       `json:"description,omitempty"`
	Timezone     string       `json:"timezone"`       // IANA timezone (e.g., "Asia/Jakarta")
	Currency     string       `json:"currency"`       // ISO 4217 (e.g., "IDR", "USD")
	Country      string       `json:"country"`        // ISO 3166-1 alpha-2
	Status       Status       `json:"status"`
	Settings     Settings     `json:"settings"`
	OwnerID      uuid.UUID    `json:"owner_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Settings holds organization-level configuration that can be updated by the owner.
type Settings struct {
	// Booking
	AllowOnlineBooking      bool  `json:"allow_online_booking"`
	RequirePaymentUpfront   bool  `json:"require_payment_upfront"`
	AutoConfirmBookings     bool  `json:"auto_confirm_bookings"`
	AllowGuestBooking       bool  `json:"allow_guest_booking"` // customer can book without account
	MinAdvanceBookingHours  int   `json:"min_advance_booking_hours"`  // minimum hours before appointment
	MaxAdvanceDays          int   `json:"max_advance_days"`           // how far ahead customers can book
	CancellationNoticehours int   `json:"cancellation_notice_hours"`  // min hours to cancel
	SlotGranularityMinutes  int   `json:"slot_granularity_minutes"`   // e.g., 15, 30, 60

	// Notifications
	SendConfirmationEmail    bool `json:"send_confirmation_email"`
	SendReminderEmail        bool `json:"send_reminder_email"`
	ReminderHoursBefore      int  `json:"reminder_hours_before"`
	SendSMSReminder          bool `json:"send_sms_reminder"`
	SendWhatsAppReminder     bool `json:"send_whatsapp_reminder"`
}

// DefaultSettings returns sensible defaults for a new organization.
func DefaultSettings() Settings {
	return Settings{
		AllowOnlineBooking:      true,
		RequirePaymentUpfront:   false,
		AutoConfirmBookings:     true,
		AllowGuestBooking:       true,
		MinAdvanceBookingHours:  1,
		MaxAdvanceDays:          60,
		CancellationNoticehours: 24,
		SlotGranularityMinutes:  30,
		SendConfirmationEmail:   true,
		SendReminderEmail:       true,
		ReminderHoursBefore:     24,
		SendSMSReminder:         false,
		SendWhatsAppReminder:    false,
	}
}

// --- Commands ----------------------------------------------------------------

// CreateOrganizationCmd holds data for creating a new organization.
type CreateOrganizationCmd struct {
	Name         string
	Slug         string
	BusinessType BusinessType
	Email        string
	Phone        string
	Timezone     string
	Currency     string
	Country      string
	OwnerID      uuid.UUID
}

// UpdateOrganizationCmd holds data for updating organization profile.
type UpdateOrganizationCmd struct {
	Name        *string
	Phone       *string
	Website     *string
	Description *string
	LogoURL     *string
	Timezone    *string
	Currency    *string
}

// UpdateSettingsCmd holds data for updating organization settings.
type UpdateSettingsCmd = Settings

// --- Queries -----------------------------------------------------------------

// ListOrganizationsFilter holds filter criteria for listing organizations.
type ListOrganizationsFilter struct {
	Status       *Status
	BusinessType *BusinessType
	Search       string
}

// --- Repository Interface ----------------------------------------------------

// Repository defines the persistence contract for the Organization domain.
// Implementations live in the repository/postgres layer.
type Repository interface {
	// Create persists a new organization. Returns the created organization.
	Create(ctx context.Context, org *Organization) (*Organization, error)

	// GetByID retrieves an organization by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)

	// GetBySlug retrieves an organization by its URL slug.
	GetBySlug(ctx context.Context, slug string) (*Organization, error)

	// Update updates organization profile fields.
	Update(ctx context.Context, org *Organization) (*Organization, error)

	// UpdateSettings updates organization booking/notification settings.
	UpdateSettings(ctx context.Context, orgID uuid.UUID, settings Settings) error

	// UpdateStatus changes the operational status of an organization.
	UpdateStatus(ctx context.Context, orgID uuid.UUID, status Status) error

	// SlugExists checks if a slug is already taken.
	SlugExists(ctx context.Context, slug string) (bool, error)

	// List returns organizations matching the filter (for platform admin).
	List(ctx context.Context, filter ListOrganizationsFilter, page, perPage int) ([]*Organization, int, error)

	// Delete permanently removes an organization (platform admin only).
	// This is a hard delete — organization must have no active subscriptions.
	Delete(ctx context.Context, orgID uuid.UUID) error
}

// --- Domain Service ----------------------------------------------------------

// SlugValidator is a pure domain function that validates a slug format.
func SlugValidator(slug string) bool {
	if len(slug) < 3 || len(slug) > 63 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
