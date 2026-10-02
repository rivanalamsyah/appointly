package integration

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
)

type Provider string

const (
	ProviderGoogleCalendar  Provider = "google_calendar"
	ProviderOutlookCalendar Provider = "outlook_calendar"
	ProviderWhatsAppTwilio  Provider = "whatsapp_twilio"
	ProviderEmailResend     Provider = "email_resend"
	ProviderOutboundWebhook Provider = "outbound_webhook"
)

type Category string

const (
	CategoryCalendar  Category = "calendar"
	CategoryMessaging Category = "messaging"
	CategoryEmail     Category = "email"
	CategoryWebhook   Category = "webhook"
)

type Status string

const (
	StatusConnected    Status = "connected"
	StatusDisconnected Status = "disconnected"
	StatusError        Status = "error"
	StatusSyncing      Status = "syncing"
)

// Integration stores connection credentials and settings for an external provider.
type Integration struct {
	ID             uuid.UUID         `json:"id"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	Provider       Provider          `json:"provider"`
	Category       Category          `json:"category"`
	Status         Status            `json:"status"`
	AccountEmail   string            `json:"account_email,omitempty"`
	AccessToken    string            `json:"-"` // Never exposed in JSON responses
	RefreshToken   string            `json:"-"` // Never exposed in JSON responses
	TokenExpiresAt *time.Time        `json:"token_expires_at,omitempty"`
	WebhookURL     string            `json:"webhook_url,omitempty"`
	WebhookSecret  string            `json:"webhook_secret,omitempty"`
	Settings       map[string]string `json:"settings,omitempty"`
	LastSyncedAt   *time.Time        `json:"last_synced_at,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// SyncState tracks 1-to-1 sync mapping between an internal appointment and an external provider event.
type SyncState struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	IntegrationID   uuid.UUID  `json:"integration_id"`
	AppointmentID   uuid.UUID  `json:"appointment_id"`
	ExternalEventID string     `json:"external_event_id"`
	SyncStatus      string     `json:"sync_status"` // synced | pending | failed
	LastSyncedAt    time.Time  `json:"last_synced_at"`
	SyncError       string     `json:"sync_error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// --- Provider Contracts (Adapter Pattern) ---

// CalendarProvider defines the interface for external calendar sync adapters.
type CalendarProvider interface {
	SyncAppointment(ctx context.Context, integration *Integration, appt *appointment.Appointment) (string, error)
	DeleteAppointment(ctx context.Context, integration *Integration, externalEventID string) error
	RefreshAccessToken(ctx context.Context, integration *Integration) (*Integration, error)
}

// WhatsAppMessage payload for messaging adapter.
type WhatsAppMessage struct {
	RecipientPhone string
	TemplateName   string
	Variables      map[string]string
}

// WhatsAppProvider defines the contract for messaging delivery.
type WhatsAppProvider interface {
	SendMessage(ctx context.Context, integration *Integration, msg WhatsAppMessage) (string, error)
}

// EmailMessage payload for email adapter.
type EmailMessage struct {
	ToEmail string
	Subject string
	BodyHTML string
}

// EmailProvider defines the contract for email delivery.
type EmailProvider interface {
	SendEmail(ctx context.Context, integration *Integration, msg EmailMessage) (string, error)
}

// OutboundWebhookService defines the contract for sending signed webhook events to tenant endpoints.
type OutboundWebhookService interface {
	DeliverEvent(ctx context.Context, integration *Integration, eventType string, payload []byte) error
}

// Repository defines the persistence interface for integration models.
type Repository interface {
	Create(ctx context.Context, integration *Integration) (*Integration, error)
	Update(ctx context.Context, integration *Integration) (*Integration, error)
	GetByProvider(ctx context.Context, orgID uuid.UUID, provider Provider) (*Integration, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Integration, error)
	Delete(ctx context.Context, orgID uuid.UUID, provider Provider) error

	// Sync State
	SaveSyncState(ctx context.Context, state *SyncState) error
	GetSyncState(ctx context.Context, orgID, apptID uuid.UUID) (*SyncState, error)
}
