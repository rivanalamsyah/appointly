package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/subscription"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// SubscriptionRepository is an in-memory thread-safe implementation of subscription.Repository.
type SubscriptionRepository struct {
	mu            sync.RWMutex
	plans         map[uuid.UUID]*subscription.Plan
	planBySlug    map[string]*subscription.Plan
	subscriptions map[uuid.UUID]*subscription.Subscription
	staffRepo     *StaffRepository
	locRepo       *LocationRepository
	apptRepo      *AppointmentRepository
}

// NewSubscriptionRepository creates and seeds default SaaS plans.
func NewSubscriptionRepository(
	staffRepo *StaffRepository,
	locRepo *LocationRepository,
	apptRepo *AppointmentRepository,
) *SubscriptionRepository {
	r := &SubscriptionRepository{
		plans:         make(map[uuid.UUID]*subscription.Plan),
		planBySlug:    make(map[string]*subscription.Plan),
		subscriptions: make(map[uuid.UUID]*subscription.Subscription),
		staffRepo:     staffRepo,
		locRepo:       locRepo,
		apptRepo:      apptRepo,
	}

	r.seedDefaultPlans()
	return r
}

func (r *SubscriptionRepository) seedDefaultPlans() {
	now := time.Now().UTC()
	freeStaff := 2
	freeLoc := 1
	freeAppts := 100

	growthStaff := 10
	growthLoc := 3
	growthAppts := 500

	proStaff := 50
	proLoc := 10
	proAppts := 2000

	defaultPlans := []*subscription.Plan{
		{
			ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:                   "Free Starter",
			Slug:                   "starter",
			Description:            "Ideal for solo practitioners starting out",
			IsActive:               true,
			MaxStaff:               &freeStaff,
			MaxLocations:           &freeLoc,
			MaxMonthlyAppointments: &freeAppts,
			MonthlyPriceCents:      0,
			AnnualPriceCents:       0,
			Currency:               "USD",
			Features:               map[string]bool{"online_booking": true, "basic_reports": true},
			DisplayOrder:           1,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:                   "Growth",
			Slug:                   "growth",
			Description:            "For growing teams and multi-service boutiques",
			IsActive:               true,
			MaxStaff:               &growthStaff,
			MaxLocations:           &growthLoc,
			MaxMonthlyAppointments: &growthAppts,
			MonthlyPriceCents:      4900,
			AnnualPriceCents:       49000,
			Currency:               "USD",
			Features:               map[string]bool{"online_booking": true, "custom_branding": true, "sms_reminders": true},
			DisplayOrder:           2,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Name:                   "Professional",
			Slug:                   "professional",
			Description:            "For established multi-branch organizations",
			IsActive:               true,
			MaxStaff:               &proStaff,
			MaxLocations:           &proLoc,
			MaxMonthlyAppointments: &proAppts,
			MonthlyPriceCents:      9900,
			AnnualPriceCents:       99000,
			Currency:               "USD",
			Features:               map[string]bool{"online_booking": true, "custom_branding": true, "api_access": true, "webhooks": true},
			DisplayOrder:           3,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Name:                   "Enterprise",
			Slug:                   "enterprise",
			Description:            "Unlimited scale with custom SLA & dedicated support",
			IsActive:               true,
			MaxStaff:               nil, // unlimited
			MaxLocations:           nil,
			MaxMonthlyAppointments: nil,
			MonthlyPriceCents:      29900,
			AnnualPriceCents:       299000,
			Currency:               "USD",
			Features:               map[string]bool{"online_booking": true, "api_access": true, "webhooks": true, "custom_branding": true, "sla": true},
			DisplayOrder:           4,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}

	for _, p := range defaultPlans {
		r.plans[p.ID] = p
		r.planBySlug[p.Slug] = p
	}
}

// ListPlans returns available active SaaS subscription tiers.
func (r *SubscriptionRepository) ListPlans(ctx context.Context) ([]*subscription.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*subscription.Plan
	for _, p := range r.plans {
		if p.IsActive {
			cp := *p
			list = append(list, &cp)
		}
	}
	return list, nil
}

// GetPlanByID retrieves a plan by UUID.
func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, planID uuid.UUID) (*subscription.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plans[planID]
	if !exists {
		return nil, apperror.NotFound("plan")
	}
	cp := *p
	return &cp, nil
}

// GetPlanBySlug retrieves a plan by slug.
func (r *SubscriptionRepository) GetPlanBySlug(ctx context.Context, slug string) (*subscription.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.planBySlug[slug]
	if !exists {
		return nil, apperror.NotFound("plan")
	}
	cp := *p
	return &cp, nil
}

// GetByOrganizationID retrieves the active subscription for an organization.
func (r *SubscriptionRepository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*subscription.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, sub := range r.subscriptions {
		if sub.OrganizationID == orgID {
			cp := *sub
			if p, ok := r.plans[sub.PlanID]; ok {
				planCopy := *p
				cp.Plan = &planCopy
			}
			return &cp, nil
		}
	}

	// Fallback default: assign Starter Plan
	starterPlan, ok := r.planBySlug["starter"]
	if !ok {
		return nil, apperror.NotFound("default starter plan")
	}

	now := time.Now().UTC()
	defaultSub := &subscription.Subscription{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		PlanID:             starterPlan.ID,
		Plan:               starterPlan,
		Status:             subscription.StatusActive,
		BillingPeriod:      subscription.BillingPeriodMonthly,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(1, 0, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	r.subscriptions[defaultSub.ID] = defaultSub
	cp := *defaultSub
	return &cp, nil
}

// Create creates a new subscription.
func (r *SubscriptionRepository) Create(ctx context.Context, cmd subscription.CreateSubscriptionCmd) (*subscription.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plans[cmd.PlanID]
	if !exists {
		return nil, apperror.NotFound("plan")
	}

	now := time.Now().UTC()
	subID := uuid.New()

	status := cmd.Status
	if status == "" {
		status = subscription.StatusActive
	}

	sub := &subscription.Subscription{
		ID:                 subID,
		OrganizationID:     cmd.OrganizationID,
		PlanID:             cmd.PlanID,
		Plan:               p,
		Status:             status,
		BillingPeriod:      cmd.BillingPeriod,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		ExternalID:         cmd.ExternalID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	r.subscriptions[subID] = sub
	return sub, nil
}

// UpdateStatus changes subscription lifecycle state.
func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, subID uuid.UUID, status subscription.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, exists := r.subscriptions[subID]
	if !exists {
		return apperror.NotFound("subscription")
	}

	sub.Status = status
	sub.UpdatedAt = time.Now().UTC()
	return nil
}

// ChangePlan upgrades or downgrades subscription plan.
func (r *SubscriptionRepository) ChangePlan(ctx context.Context, cmd subscription.ChangePlanCmd) (*subscription.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plans[cmd.NewPlanID]
	if !exists {
		return nil, apperror.NotFound("plan")
	}

	now := time.Now().UTC()

	var existingSub *subscription.Subscription
	for _, s := range r.subscriptions {
		if s.OrganizationID == cmd.OrganizationID {
			existingSub = s
			break
		}
	}

	if existingSub == nil {
		subID := uuid.New()
		existingSub = &subscription.Subscription{
			ID:                 subID,
			OrganizationID:     cmd.OrganizationID,
			PlanID:             cmd.NewPlanID,
			Plan:               p,
			Status:             subscription.StatusActive,
			BillingPeriod:      cmd.BillingPeriod,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now.AddDate(0, 1, 0),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		r.subscriptions[subID] = existingSub
	} else {
		existingSub.PlanID = cmd.NewPlanID
		existingSub.Plan = p
		existingSub.BillingPeriod = cmd.BillingPeriod
		existingSub.Status = subscription.StatusActive
		existingSub.UpdatedAt = now
	}

	cp := *existingSub
	return &cp, nil
}

// Cancel sets subscription status to cancelled.
func (r *SubscriptionRepository) Cancel(ctx context.Context, orgID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for _, s := range r.subscriptions {
		if s.OrganizationID == orgID {
			s.Status = subscription.StatusCancelled
			s.CancelledAt = &now
			s.UpdatedAt = now
			return nil
		}
	}
	return apperror.NotFound("subscription")
}

// GetUsageMetrics computes current resource counts for limit enforcement.
func (r *SubscriptionRepository) GetUsageMetrics(ctx context.Context, orgID uuid.UUID) (*subscription.UsageMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := &subscription.UsageMetrics{}

	// Staff count
	if r.staffRepo != nil {
		cnt, _ := r.staffRepo.CountActive(ctx, orgID)
		metrics.StaffCount = cnt
	}

	// Location count
	if r.locRepo != nil {
		cnt, _ := r.locRepo.CountActive(ctx, orgID)
		metrics.LocationCount = cnt
	}

	// Monthly booking count
	if r.apptRepo != nil {
		now := time.Now().UTC()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		appts, _ := r.apptRepo.ListForCalendar(ctx, orgID, monthStart, monthEnd)
		metrics.MonthlyBookingCount = len(appts)
	}

	return metrics, nil
}
