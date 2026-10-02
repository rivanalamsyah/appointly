// Package subscription defines the SaaS Subscription & Feature Limit domain.
// SaaS subscription is completely decoupled from customer appointment payment.
package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the operational state of a SaaS subscription.
type Status string

const (
	StatusTrialing Status = "trialing"
	StatusActive   Status = "active"
	StatusPastDue  Status = "past_due"
	StatusCancelled Status = "cancelled"
	StatusExpired  Status = "expired"
	StatusUnpaid   Status = "unpaid"
)

// BillingPeriod represents the subscription billing cycle.
type BillingPeriod string

const (
	BillingPeriodMonthly BillingPeriod = "monthly"
	BillingPeriodAnnual  BillingPeriod = "annual"
)

// Plan represents a SaaS tier (Free, Starter, Growth, Professional, Enterprise).
type Plan struct {
	ID                     uuid.UUID         `json:"id"`
	Name                   string            `json:"name"`
	Slug                   string            `json:"slug"`
	Description            string            `json:"description,omitempty"`
	IsActive               bool              `json:"is_active"`
	MaxStaff               *int              `json:"max_staff,omitempty"` // nil = unlimited
	MaxLocations           *int              `json:"max_locations,omitempty"`
	MaxMonthlyAppointments *int              `json:"max_monthly_appointments,omitempty"`
	MaxCustomers           *int              `json:"max_customers,omitempty"`
	MaxServices            *int              `json:"max_services,omitempty"`
	MonthlyPriceCents      int64             `json:"monthly_price_cents"`
	AnnualPriceCents       int64             `json:"annual_price_cents"`
	Currency               string            `json:"currency"`
	Features               map[string]bool   `json:"features"`
	DisplayOrder           int               `json:"display_order"`
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
}

// Subscription represents an organization's active SaaS tier subscription.
type Subscription struct {
	ID                 uuid.UUID     `json:"id"`
	OrganizationID     uuid.UUID     `json:"organization_id"`
	PlanID             uuid.UUID     `json:"plan_id"`
	Plan               *Plan         `json:"plan,omitempty"`
	Status             Status        `json:"status"`
	BillingPeriod      BillingPeriod `json:"billing_period"`
	CurrentPeriodStart time.Time     `json:"current_period_start"`
	CurrentPeriodEnd   time.Time     `json:"current_period_end"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	TrialEndsAt        *time.Time    `json:"trial_ends_at,omitempty"`
	ExternalID         string        `json:"external_id,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// IsActive returns true if subscription is valid for feature access.
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrialing
}

// UsageMetrics holds current usage vs plan capacity for enforcement.
type UsageMetrics struct {
	StaffCount             int   `json:"staff_count"`
	MaxStaff               *int  `json:"max_staff,omitempty"`
	LocationCount          int   `json:"location_count"`
	MaxLocations           *int  `json:"max_locations,omitempty"`
	MonthlyBookingCount    int   `json:"monthly_booking_count"`
	MaxMonthlyAppointments *int  `json:"max_monthly_appointments,omitempty"`
}

// --- Commands & Filters ------------------------------------------------------

type CreateSubscriptionCmd struct {
	OrganizationID uuid.UUID
	PlanID         uuid.UUID
	BillingPeriod  BillingPeriod
	Status         Status
	ExternalID     string
	TrialDays      int
}

type ChangePlanCmd struct {
	OrganizationID uuid.UUID
	NewPlanID      uuid.UUID
	BillingPeriod  BillingPeriod
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// Plans
	ListPlans(ctx context.Context) ([]*Plan, error)
	GetPlanByID(ctx context.Context, planID uuid.UUID) (*Plan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*Plan, error)

	// Subscriptions
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*Subscription, error)
	Create(ctx context.Context, cmd CreateSubscriptionCmd) (*Subscription, error)
	UpdateStatus(ctx context.Context, subID uuid.UUID, status Status) error
	ChangePlan(ctx context.Context, cmd ChangePlanCmd) (*Subscription, error)
	Cancel(ctx context.Context, orgID uuid.UUID) error

	// Usage tracking
	GetUsageMetrics(ctx context.Context, orgID uuid.UUID) (*UsageMetrics, error)
}
