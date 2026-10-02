// Package rbac defines the Role-Based Access Control domain for Appointly.
// Roles are scoped to organizations. A user can have different roles in
// different organizations. Permission checks happen in the use case layer —
// never in the handler or repository.
package rbac

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Role represents a named set of permissions within an organization.
// Roles are hierarchical: OWNER > ADMIN > MANAGER > RECEPTIONIST > STAFF
type Role string

const (
	RoleOwner        Role = "OWNER"        // Full control over the organization
	RoleAdmin        Role = "ADMIN"        // Admin without billing/org deletion
	RoleManager      Role = "MANAGER"      // Manage staff, services, appointments
	RoleReceptionist Role = "RECEPTIONIST" // Create/view appointments, customers
	RoleStaff        Role = "STAFF"        // View own schedule, own appointments
)

// Permission is a fine-grained action that can be allowed or denied.
type Permission string

const (
	// Organization
	PermOrgRead           Permission = "org:read"
	PermOrgUpdate         Permission = "org:update"
	PermOrgDelete         Permission = "org:delete"
	PermOrgManageSettings Permission = "org:manage_settings"

	// Members & Team
	PermMemberInvite  Permission = "member:invite"
	PermMemberRemove  Permission = "member:remove"
	PermMemberRead    Permission = "member:read"
	PermMemberUpdate  Permission = "member:update"

	// Staff
	PermStaffCreate   Permission = "staff:create"
	PermStaffRead     Permission = "staff:read"
	PermStaffUpdate   Permission = "staff:update"
	PermStaffDelete   Permission = "staff:delete"
	PermStaffSchedule Permission = "staff:manage_schedule"

	// Services
	PermServiceCreate Permission = "service:create"
	PermServiceRead   Permission = "service:read"
	PermServiceUpdate Permission = "service:update"
	PermServiceDelete Permission = "service:delete"

	// Locations
	PermLocationCreate Permission = "location:create"
	PermLocationRead   Permission = "location:read"
	PermLocationUpdate Permission = "location:update"
	PermLocationDelete Permission = "location:delete"

	// Resources
	PermResourceCreate Permission = "resource:create"
	PermResourceRead   Permission = "resource:read"
	PermResourceUpdate Permission = "resource:update"
	PermResourceDelete Permission = "resource:delete"

	// Customers
	PermCustomerCreate Permission = "customer:create"
	PermCustomerRead   Permission = "customer:read"
	PermCustomerUpdate Permission = "customer:update"
	PermCustomerDelete Permission = "customer:delete"

	// Appointments
	PermAppointmentCreate   Permission = "appointment:create"
	PermAppointmentRead     Permission = "appointment:read"
	PermAppointmentUpdate   Permission = "appointment:update"
	PermAppointmentCancel   Permission = "appointment:cancel"
	PermAppointmentComplete Permission = "appointment:complete"
	PermAppointmentNoShow   Permission = "appointment:no_show"

	// Payments
	PermPaymentRead   Permission = "payment:read"
	PermPaymentRefund Permission = "payment:refund"

	// Reports
	PermReportRead   Permission = "report:read"
	PermReportExport Permission = "report:export"

	// Billing/Subscription
	PermBillingRead   Permission = "billing:read"
	PermBillingManage Permission = "billing:manage"

	// Integrations
	PermIntegrationManage Permission = "integration:manage"

	// Webhooks
	PermWebhookManage Permission = "webhook:manage"

	// Audit Log
	PermAuditRead Permission = "audit:read"
)

// RolePermissions maps each Role to its allowed Permissions.
// This is the source of truth for the RBAC policy.
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		PermOrgRead, PermOrgUpdate, PermOrgDelete, PermOrgManageSettings,
		PermMemberInvite, PermMemberRemove, PermMemberRead, PermMemberUpdate,
		PermStaffCreate, PermStaffRead, PermStaffUpdate, PermStaffDelete, PermStaffSchedule,
		PermServiceCreate, PermServiceRead, PermServiceUpdate, PermServiceDelete,
		PermLocationCreate, PermLocationRead, PermLocationUpdate, PermLocationDelete,
		PermResourceCreate, PermResourceRead, PermResourceUpdate, PermResourceDelete,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate, PermCustomerDelete,
		PermAppointmentCreate, PermAppointmentRead, PermAppointmentUpdate,
		PermAppointmentCancel, PermAppointmentComplete, PermAppointmentNoShow,
		PermPaymentRead, PermPaymentRefund,
		PermReportRead, PermReportExport,
		PermBillingRead, PermBillingManage,
		PermIntegrationManage, PermWebhookManage,
		PermAuditRead,
	},
	RoleAdmin: {
		PermOrgRead, PermOrgUpdate, PermOrgManageSettings,
		PermMemberInvite, PermMemberRemove, PermMemberRead, PermMemberUpdate,
		PermStaffCreate, PermStaffRead, PermStaffUpdate, PermStaffDelete, PermStaffSchedule,
		PermServiceCreate, PermServiceRead, PermServiceUpdate, PermServiceDelete,
		PermLocationCreate, PermLocationRead, PermLocationUpdate, PermLocationDelete,
		PermResourceCreate, PermResourceRead, PermResourceUpdate, PermResourceDelete,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate, PermCustomerDelete,
		PermAppointmentCreate, PermAppointmentRead, PermAppointmentUpdate,
		PermAppointmentCancel, PermAppointmentComplete, PermAppointmentNoShow,
		PermPaymentRead, PermPaymentRefund,
		PermReportRead, PermReportExport,
		PermBillingRead,
		PermIntegrationManage, PermWebhookManage,
		PermAuditRead,
	},
	RoleManager: {
		PermOrgRead,
		PermMemberRead,
		PermStaffCreate, PermStaffRead, PermStaffUpdate, PermStaffSchedule,
		PermServiceCreate, PermServiceRead, PermServiceUpdate,
		PermLocationRead,
		PermResourceRead,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate,
		PermAppointmentCreate, PermAppointmentRead, PermAppointmentUpdate,
		PermAppointmentCancel, PermAppointmentComplete, PermAppointmentNoShow,
		PermPaymentRead,
		PermReportRead,
		PermBillingRead,
	},
	RoleReceptionist: {
		PermOrgRead,
		PermStaffRead,
		PermServiceRead,
		PermLocationRead,
		PermResourceRead,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate,
		PermAppointmentCreate, PermAppointmentRead, PermAppointmentUpdate,
		PermAppointmentCancel, PermAppointmentComplete, PermAppointmentNoShow,
		PermPaymentRead,
	},
	RoleStaff: {
		PermOrgRead,
		PermServiceRead,
		PermLocationRead,
		PermCustomerRead,
		PermAppointmentRead,
		PermAppointmentComplete,
		PermAppointmentNoShow,
	},
}

// HasPermission checks if a role has a specific permission.
// This is a pure function — no database access required.
func HasPermission(role Role, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// --- Member ------------------------------------------------------------------

// OrganizationMember links a User to an Organization with a specific Role.
type OrganizationMember struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Role           Role       `json:"role"`
	InvitedBy      *uuid.UUID `json:"invited_by,omitempty"`
	InvitedAt      *time.Time `json:"invited_at,omitempty"`
	JoinedAt       *time.Time `json:"joined_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// MemberInvitation represents a pending invitation to join an organization.
type MemberInvitation struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Email          string     `json:"email"`
	Role           Role       `json:"role"`
	InvitedBy      uuid.UUID  `json:"invited_by"`
	TokenHash      string     `json:"-"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// IsValid checks if the invitation is still valid.
func (i *MemberInvitation) IsValid() bool {
	return i.AcceptedAt == nil && time.Now().Before(i.ExpiresAt)
}

// --- Commands ----------------------------------------------------------------

type InviteMemberCmd struct {
	OrganizationID uuid.UUID
	Email          string
	Role           Role
	InvitedBy      uuid.UUID
}

type UpdateMemberRoleCmd struct {
	OrganizationID uuid.UUID
	MemberID       uuid.UUID
	NewRole        Role
	UpdatedBy      uuid.UUID
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// Members
	CreateMember(ctx context.Context, member *OrganizationMember) (*OrganizationMember, error)
	GetMemberByID(ctx context.Context, id uuid.UUID) (*OrganizationMember, error)
	GetMemberByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*OrganizationMember, error)
	ListMembersByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, id uuid.UUID, role Role) error
	DeactivateMember(ctx context.Context, id uuid.UUID) error
	CountActiveMembers(ctx context.Context, orgID uuid.UUID) (int, error)

	// Invitations
	CreateInvitation(ctx context.Context, inv *MemberInvitation) (*MemberInvitation, error)
	GetInvitationByToken(ctx context.Context, tokenHash string) (*MemberInvitation, error)
	GetPendingInvitationByEmail(ctx context.Context, orgID uuid.UUID, email string) (*MemberInvitation, error)
	AcceptInvitation(ctx context.Context, id uuid.UUID) error
	ListPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]*MemberInvitation, error)
	DeleteInvitation(ctx context.Context, id uuid.UUID) error
}

// --- Authorization Context ---------------------------------------------------

// AuthContext holds the authenticated user and their role in the active organization.
// This is extracted from the JWT and attached to the request context.
type AuthContext struct {
	UserID    uuid.UUID
	OrgID     uuid.UUID
	Role      Role
	IsOwner   bool
	IsSuperAdmin bool
}

// Can checks if the authenticated user has a specific permission.
func (a *AuthContext) Can(perm Permission) bool {
	if a.IsSuperAdmin {
		return true
	}
	return HasPermission(a.Role, perm)
}

// contextKey is an unexported type for context keys.
type contextKey string

const authContextKey contextKey = "auth_context"

// NewContext stores AuthContext in the request context.
func NewContext(ctx context.Context, auth *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// FromContext extracts AuthContext from the request context.
// Returns nil if no auth context is present.
func FromContext(ctx context.Context) *AuthContext {
	if v, ok := ctx.Value(authContextKey).(*AuthContext); ok {
		return v
	}
	return nil
}
