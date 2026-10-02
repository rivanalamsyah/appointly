// Package audit defines the AuditLog domain.
// AuditLog provides an immutable append-only trail of significant operations.
// This is critical for compliance, debugging, and tenant trust.
// Audit logs are NEVER updated or deleted.
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Action represents the type of audited operation.
type Action string

const (
	// Authentication
	ActionUserLogin            Action = "user.login"
	ActionUserLogout           Action = "user.logout"
	ActionUserLoginFailed      Action = "user.login_failed"
	ActionPasswordChanged      Action = "user.password_changed"
	ActionPasswordReset        Action = "user.password_reset"

	// Organization
	ActionOrgCreated           Action = "org.created"
	ActionOrgUpdated           Action = "org.updated"
	ActionOrgSettingsUpdated   Action = "org.settings_updated"

	// Members
	ActionMemberInvited        Action = "member.invited"
	ActionMemberJoined         Action = "member.joined"
	ActionMemberRemoved        Action = "member.removed"
	ActionMemberRoleChanged    Action = "member.role_changed"

	// Appointments
	ActionAppointmentCreated   Action = "appointment.created"
	ActionAppointmentConfirmed Action = "appointment.confirmed"
	ActionAppointmentCancelled Action = "appointment.cancelled"
	ActionAppointmentCompleted Action = "appointment.completed"
	ActionAppointmentRescheduled Action = "appointment.rescheduled"
	ActionAppointmentNoShow    Action = "appointment.no_show"

	// Payments
	ActionPaymentCreated       Action = "payment.created"
	ActionPaymentCompleted     Action = "payment.completed"
	ActionPaymentFailed        Action = "payment.failed"
	ActionRefundCreated        Action = "refund.created"

	// Staff
	ActionStaffCreated         Action = "staff.created"
	ActionStaffUpdated         Action = "staff.updated"
	ActionStaffDeleted         Action = "staff.deleted"

	// Services
	ActionServiceCreated       Action = "service.created"
	ActionServiceUpdated       Action = "service.updated"
	ActionServiceDeleted       Action = "service.deleted"

	// Subscription
	ActionSubscriptionChanged  Action = "subscription.changed"
	ActionSubscriptionCancelled Action = "subscription.cancelled"

	// Integrations
	ActionWebhookCreated       Action = "webhook.created"
	ActionWebhookDeleted       Action = "webhook.deleted"
	ActionAPIKeyCreated        Action = "api_key.created"
	ActionAPIKeyRevoked        Action = "api_key.revoked"
)

// AuditLog is an immutable record of a significant operation.
type AuditLog struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"` // nil for platform-level events
	ActorID        *uuid.UUID             `json:"actor_id,omitempty"`        // nil for system events
	ActorEmail     string                 `json:"actor_email,omitempty"`
	ActorType      string                 `json:"actor_type"` // user | system | api_key
	Action         Action                 `json:"action"`
	ResourceType   string                 `json:"resource_type,omitempty"` // appointment | payment | staff ...
	ResourceID     *uuid.UUID             `json:"resource_id,omitempty"`
	Changes        map[string]interface{} `json:"changes,omitempty"` // before/after for updates
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	RequestID      string                 `json:"request_id,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// --- Entry Builder -----------------------------------------------------------

// Entry is a builder for creating audit log entries.
type Entry struct {
	log AuditLog
}

// NewEntry creates a new audit log entry builder.
func NewEntry(action Action) *Entry {
	return &Entry{
		log: AuditLog{
			Action:    action,
			ActorType: "user",
			CreatedAt: time.Now().UTC(),
		},
	}
}

func (e *Entry) ForOrg(orgID uuid.UUID) *Entry {
	e.log.OrganizationID = &orgID
	return e
}

func (e *Entry) ByUser(userID uuid.UUID, email string) *Entry {
	e.log.ActorID = &userID
	e.log.ActorEmail = email
	e.log.ActorType = "user"
	return e
}

func (e *Entry) BySystem() *Entry {
	e.log.ActorType = "system"
	return e
}

func (e *Entry) OnResource(resourceType string, resourceID uuid.UUID) *Entry {
	e.log.ResourceType = resourceType
	e.log.ResourceID = &resourceID
	return e
}

func (e *Entry) WithChanges(changes map[string]interface{}) *Entry {
	e.log.Changes = changes
	return e
}

func (e *Entry) WithMetadata(meta map[string]interface{}) *Entry {
	e.log.Metadata = meta
	return e
}

func (e *Entry) WithIP(ip string) *Entry {
	e.log.IPAddress = ip
	return e
}

func (e *Entry) WithRequestID(reqID string) *Entry {
	e.log.RequestID = reqID
	return e
}

func (e *Entry) Build() AuditLog {
	if e.log.ID == uuid.Nil {
		e.log.ID = uuid.New()
	}
	return e.log
}

// --- Repository Interface ----------------------------------------------------

// Repository defines the persistence contract for audit logs.
// There is intentionally NO Update or Delete method — audit logs are immutable.
type Repository interface {
	// Create appends a new audit log entry. This should never fail silently.
	Create(ctx context.Context, log AuditLog) error

	// List returns audit log entries with filtering and pagination.
	List(ctx context.Context, orgID uuid.UUID, filter ListFilter, page, perPage int) ([]*AuditLog, int, error)

	// GetByID retrieves a specific audit log entry.
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)
}

// ListFilter holds filter criteria for listing audit logs.
type ListFilter struct {
	ActorID      *uuid.UUID
	Action       *Action
	ResourceType *string
	ResourceID   *uuid.UUID
	DateFrom     *time.Time
	DateTo       *time.Time
}

// --- Logger Interface --------------------------------------------------------

// Logger provides a convenient way to write audit log entries from use cases.
type Logger interface {
	Log(ctx context.Context, entry AuditLog) error
}
