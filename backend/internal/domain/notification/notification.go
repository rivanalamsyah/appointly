// Package notification defines the Notification domain.
// Notifications are event-driven and processed asynchronously by workers.
// API requests MUST NOT depend on notification delivery success.
package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Channel represents the delivery channel for a notification.
type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelInApp    Channel = "in_app"
	ChannelWebhook  Channel = "webhook"
)

// EventType represents the business event that triggers a notification.
type EventType string

const (
	EventBookingCreated     EventType = "booking.created"
	EventBookingConfirmed   EventType = "booking.confirmed"
	EventBookingRescheduled EventType = "booking.rescheduled"
	EventBookingCancelled   EventType = "booking.cancelled"
	EventBookingCompleted   EventType = "booking.completed"
	EventBookingReminder    EventType = "booking.reminder"
	EventPaymentSucceeded   EventType = "payment.succeeded"
	EventPaymentFailed      EventType = "payment.failed"
	EventPaymentRefunded    EventType = "payment.refunded"
	EventStaffInvited       EventType = "staff.invited"
	EventMemberInvited      EventType = "member.invited"
)

// Status represents the delivery status of a notification.
type Status string

const (
	StatusPending   Status = "pending"
	StatusSent      Status = "sent"
	StatusFailed    Status = "failed"
	StatusSkipped   Status = "skipped" // e.g., channel disabled for this org
)

// Notification represents a notification queued for delivery.
type Notification struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	EventType      EventType  `json:"event_type"`
	RecipientType  string     `json:"recipient_type"` // customer | staff | member
	RecipientID    *uuid.UUID `json:"recipient_id,omitempty"`
	RecipientEmail string     `json:"recipient_email,omitempty"`
	RecipientPhone string     `json:"recipient_phone,omitempty"`
	RecipientName  string     `json:"recipient_name,omitempty"`

	// Reference to the triggering entity
	ReferenceType string    `json:"reference_type"` // appointment | payment | invitation
	ReferenceID   uuid.UUID `json:"reference_id"`

	// Channels to attempt delivery
	Channels []Channel `json:"channels"`

	// Template data (JSON-serialized payload for template rendering)
	TemplateData map[string]interface{} `json:"template_data,omitempty"`

	ScheduledAt *time.Time `json:"scheduled_at,omitempty"` // nil = send immediately
	CreatedAt   time.Time  `json:"created_at"`
}

// NotificationDelivery records a delivery attempt for a specific channel.
type NotificationDelivery struct {
	ID             uuid.UUID  `json:"id"`
	NotificationID uuid.UUID  `json:"notification_id"`
	Channel        Channel    `json:"channel"`
	Status         Status     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	LastAttemptAt  *time.Time `json:"last_attempt_at,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	ExternalID     string     `json:"external_id,omitempty"` // message ID from provider
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// --- Notification Publisher Interface ----------------------------------------

// Publisher is used by use cases to publish notification events.
// The implementation enqueues the notification as a background job.
// API handlers MUST NOT wait for notification delivery.
type Publisher interface {
	// Publish enqueues a notification for asynchronous delivery.
	Publish(ctx context.Context, notification *Notification) error
}

// --- Template Data Builders --------------------------------------------------
// These helpers create consistent template data payloads for each event type.

// AppointmentTemplateData builds template data for appointment notifications.
type AppointmentTemplateData struct {
	OrganizationName string
	CustomerName     string
	StaffName        string
	ServiceName      string
	LocationName     string
	LocationAddress  string
	AppointmentDate  string // human-readable in customer's timezone
	AppointmentTime  string
	Duration         string
	PriceFormatted   string
	BookingCode      string // short reference code
	CancelURL        string
	RescheduleURL    string
	ManageURL        string
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListPending(ctx context.Context, limit int) ([]*Notification, error)
	ListScheduledReady(ctx context.Context, now time.Time, limit int) ([]*Notification, error)

	// Delivery
	CreateDelivery(ctx context.Context, d *NotificationDelivery) (*NotificationDelivery, error)
	UpdateDelivery(ctx context.Context, d *NotificationDelivery) error
	ListDeliveries(ctx context.Context, notificationID uuid.UUID) ([]*NotificationDelivery, error)
}
