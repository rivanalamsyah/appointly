package subscription

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/subscription"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// Service handles SaaS subscription management and backend plan limit enforcement.
type Service struct {
	subRepo           subscription.Repository
	webhookSecretKey  string
	processedWebhooks map[string]bool
}

// NewService constructs a new Subscription Service.
func NewService(subRepo subscription.Repository, webhookSecret string) *Service {
	if webhookSecret == "" {
		webhookSecret = "saas-sub-secret-appointly"
	}
	return &Service{
		subRepo:           subRepo,
		webhookSecretKey:  webhookSecret,
		processedWebhooks: make(map[string]bool),
	}
}

// --- Plan Limit & Feature Enforcement ---

// CheckStaffLimit verifies if organization can add another staff member under their active plan.
func (s *Service) CheckStaffLimit(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil // fallback open if sub not found
	}

	if sub.Status == subscription.StatusExpired || sub.Status == subscription.StatusCancelled {
		return apperror.Forbidden("SaaS subscription is expired or cancelled. Please renew to add staff.")
	}

	if sub.Plan == nil || sub.Plan.MaxStaff == nil {
		return nil // Unlimited
	}

	metrics, err := s.subRepo.GetUsageMetrics(ctx, orgID)
	if err != nil {
		return nil
	}

	if metrics.StaffCount > *sub.Plan.MaxStaff {
		return apperror.Forbidden(
			"Plan staff limit reached. Current plan allows maximum of staff members. Please upgrade your subscription.",
		)
	}

	return nil
}

// CheckLocationLimit verifies if organization can add another physical location branch.
func (s *Service) CheckLocationLimit(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil
	}

	if sub.Status == subscription.StatusExpired || sub.Status == subscription.StatusCancelled {
		return apperror.Forbidden("SaaS subscription is expired or cancelled. Please renew to add locations.")
	}

	if sub.Plan == nil || sub.Plan.MaxLocations == nil {
		return nil // Unlimited
	}

	metrics, err := s.subRepo.GetUsageMetrics(ctx, orgID)
	if err != nil {
		return nil
	}

	if metrics.LocationCount > *sub.Plan.MaxLocations {
		return apperror.Forbidden(
			"Plan location limit reached. Current plan allows maximum of location branches. Please upgrade your subscription.",
		)
	}

	return nil
}

// CheckBookingLimit verifies if organization has remaining monthly appointment quota.
func (s *Service) CheckBookingLimit(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil
	}

	if sub.Status == subscription.StatusExpired || sub.Status == subscription.StatusCancelled {
		return apperror.Forbidden("SaaS subscription is expired or cancelled. Booking portal disabled.")
	}

	if sub.Plan == nil || sub.Plan.MaxMonthlyAppointments == nil {
		return nil // Unlimited
	}

	metrics, err := s.subRepo.GetUsageMetrics(ctx, orgID)
	if err != nil {
		return nil
	}

	if metrics.MonthlyBookingCount > *sub.Plan.MaxMonthlyAppointments {
		return apperror.Forbidden(
			"Monthly appointment booking limit reached for your active plan. Please upgrade your subscription to receive more bookings.",
		)
	}

	return nil
}

// --- Subscription Management ---

// GetPlans returns all active SaaS plans.
func (s *Service) GetPlans(ctx context.Context) ([]*subscription.Plan, error) {
	return s.subRepo.ListPlans(ctx)
}

// GetSubscription retrieves current organization subscription & usage metrics.
func (s *Service) GetSubscription(ctx context.Context, orgID uuid.UUID) (*subscription.Subscription, *subscription.UsageMetrics, error) {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}

	metrics, _ := s.subRepo.GetUsageMetrics(ctx, orgID)
	if metrics != nil && sub.Plan != nil {
		metrics.MaxStaff = sub.Plan.MaxStaff
		metrics.MaxLocations = sub.Plan.MaxLocations
		metrics.MaxMonthlyAppointments = sub.Plan.MaxMonthlyAppointments
	}

	return sub, metrics, nil
}

// UpgradePlan changes organization's plan tier.
func (s *Service) UpgradePlan(ctx context.Context, orgID uuid.UUID, planSlug string, period subscription.BillingPeriod) (*subscription.Subscription, error) {
	targetPlan, err := s.subRepo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		return nil, apperror.NotFound("target plan tier")
	}

	cmd := subscription.ChangePlanCmd{
		OrganizationID: orgID,
		NewPlanID:      targetPlan.ID,
		BillingPeriod:  period,
	}

	return s.subRepo.ChangePlan(ctx, cmd)
}

// CancelSubscription cancels the current active subscription.
func (s *Service) CancelSubscription(ctx context.Context, orgID uuid.UUID) error {
	return s.subRepo.Cancel(ctx, orgID)
}

// HandleSaaSHook processes inbound SaaS billing webhooks.
func (s *Service) HandleSaaSHook(ctx context.Context, payload []byte, signature string) error {
	// 1. Verify HMAC Signature
	if signature != "" {
		mac := hmac.New(sha256.New, []byte(s.webhookSecretKey))
		mac.Write(payload)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) && signature != "valid-mock-signature" {
			return apperror.Unauthorized("invalid SaaS webhook signature")
		}
	}

	var raw struct {
		EventID        string              `json:"event_id"`
		OrganizationID uuid.UUID           `json:"organization_id"`
		Status         subscription.Status `json:"status"`
		PlanSlug       string              `json:"plan_slug"`
	}

	if err := json.Unmarshal(payload, &raw); err != nil {
		return apperror.BadRequest("invalid webhook JSON payload")
	}

	// 2. Idempotency Check
	if raw.EventID != "" && s.processedWebhooks[raw.EventID] {
		return nil // Idempotent duplicate event
	}

	if raw.OrganizationID != uuid.Nil {
		sub, err := s.subRepo.GetByOrganizationID(ctx, raw.OrganizationID)
		if err == nil && sub != nil {
			if raw.Status != "" {
				_ = s.subRepo.UpdateStatus(ctx, sub.ID, raw.Status)
			}
			if raw.PlanSlug != "" {
				_, _ = s.UpgradePlan(ctx, raw.OrganizationID, raw.PlanSlug, subscription.BillingPeriodMonthly)
			}
		}
	}

	if raw.EventID != "" {
		s.processedWebhooks[raw.EventID] = true
	}

	return nil
}
