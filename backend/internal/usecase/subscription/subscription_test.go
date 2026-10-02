package subscription_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/domain/subscription"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	subuc "github.com/appointly/appointly/backend/internal/usecase/subscription"
)

func setupTestSubscriptionService(t *testing.T) (
	*subuc.Service,
	*organization.Organization,
	*memory.StaffRepository,
	*memory.LocationRepository,
	*memory.AppointmentRepository,
	*memory.SubscriptionRepository,
) {
	orgRepo := memory.NewOrgRepository()
	staffRepo := memory.NewStaffRepository()
	locRepo := memory.NewLocationRepository()
	apptRepo := memory.NewAppointmentRepository()
	subRepo := memory.NewSubscriptionRepository(staffRepo, locRepo, apptRepo)

	subSvc := subuc.NewService(subRepo, "mock-sub-secret")

	org, err := orgRepo.Create(context.Background(), &organization.Organization{
		ID:       uuid.New(),
		Name:     "SaaS Limit Test Org",
		Slug:     "saas-limit-org",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	return subSvc, org, staffRepo, locRepo, apptRepo, subRepo
}

func TestSubscription_PlanLimitsEnforcement(t *testing.T) {
	subSvc, org, staffRepo, _, _, _ := setupTestSubscriptionService(t)
	ctx := context.Background()

	// Assign Starter Plan (max_staff = 2)
	_, err := subSvc.UpgradePlan(ctx, org.ID, "starter", subscription.BillingPeriodMonthly)
	if err != nil {
		t.Fatalf("failed to assign starter plan: %v", err)
	}

	// 1. Create 2 staff members (allowed by Starter plan max_staff = 2)
	_, _ = staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Staff One",
		LastName:       "Test",
	})
	_, _ = staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Staff Two",
		LastName:       "Test",
	})

	// Check limit check for staff -> should pass with 2 staff
	if err := subSvc.CheckStaffLimit(ctx, org.ID); err != nil {
		t.Errorf("expected 2 staff to pass Starter plan limit, got %v", err)
	}

	// Add 3rd staff member -> Exceeds Starter plan limit
	_, _ = staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Staff Three",
		LastName:       "Exceed",
	})

	// Check limit check for staff -> Should fail with Plan Limit error
	errOver := subSvc.CheckStaffLimit(ctx, org.ID)
	if errOver == nil {
		t.Error("SECURITY FAILURE: Expected plan staff limit error when 3rd staff added to Starter plan, got nil")
	}
}

func TestSubscription_PlanUpgradeResetsLimits(t *testing.T) {
	subSvc, org, staffRepo, _, _, _ := setupTestSubscriptionService(t)
	ctx := context.Background()

	// Add 3 staff members
	for i := 1; i <= 3; i++ {
		_, _ = staffRepo.Create(ctx, staff.CreateStaffCmd{
			OrganizationID: org.ID,
			FirstName:      "Staff",
			LastName:       "Member",
		})
	}

	// Upgrade to Growth Plan (max_staff = 10)
	upgradedSub, err := subSvc.UpgradePlan(ctx, org.ID, "growth", subscription.BillingPeriodMonthly)
	if err != nil {
		t.Fatalf("failed to upgrade plan to growth: %v", err)
	}

	if upgradedSub.Plan.Slug != "growth" {
		t.Errorf("expected plan slug growth, got %s", upgradedSub.Plan.Slug)
	}

	// Staff limit check should now PASS under Growth Plan (3 <= 10)
	if err := subSvc.CheckStaffLimit(ctx, org.ID); err != nil {
		t.Errorf("expected 3 staff to pass Growth plan limit, got %v", err)
	}
}

func TestSubscription_LifecycleTransitionsAndCancellation(t *testing.T) {
	subSvc, org, _, _, _, subRepo := setupTestSubscriptionService(t)
	ctx := context.Background()

	// 1. Get initial subscription
	sub, metrics, err := subSvc.GetSubscription(ctx, org.ID)
	if err != nil {
		t.Fatalf("failed to get subscription: %v", err)
	}
	if sub.Status != subscription.StatusActive {
		t.Errorf("expected initial status active, got %s", sub.Status)
	}
	if metrics == nil {
		t.Error("expected usage metrics to be populated, got nil")
	}

	// 2. Cancel Subscription
	err = subSvc.CancelSubscription(ctx, org.ID)
	if err != nil {
		t.Fatalf("failed to cancel subscription: %v", err)
	}

	cancelledSub, _, _ := subSvc.GetSubscription(ctx, org.ID)
	if cancelledSub.Status != subscription.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", cancelledSub.Status)
	}

	// 3. Limit Check on Cancelled Subscription -> Should be forbidden
	errLimit := subSvc.CheckStaffLimit(ctx, org.ID)
	if errLimit == nil {
		t.Error("expected limit error on cancelled subscription, got nil")
	}

	// 4. Update status back to active via webhook / admin
	_ = subRepo.UpdateStatus(ctx, cancelledSub.ID, subscription.StatusActive)
	activeSub, _, _ := subSvc.GetSubscription(ctx, org.ID)
	if activeSub.Status != subscription.StatusActive {
		t.Errorf("expected status active after renewal, got %s", activeSub.Status)
	}
}

func TestSubscription_SaaSWebhookProcessing(t *testing.T) {
	subSvc, org, _, _, _, _ := setupTestSubscriptionService(t)
	ctx := context.Background()

	payload, _ := json.Marshal(map[string]interface{}{
		"event_id":        "evt_saas_777",
		"organization_id": org.ID,
		"plan_slug":       "professional",
		"status":          "active",
	})

	err := subSvc.HandleSaaSHook(ctx, payload, "valid-mock-signature")
	if err != nil {
		t.Fatalf("failed to handle SaaS webhook: %v", err)
	}

	sub, _, _ := subSvc.GetSubscription(ctx, org.ID)
	if sub.Plan.Slug != "professional" {
		t.Errorf("expected plan to be upgraded to professional via webhook, got %s", sub.Plan.Slug)
	}

	// Duplicate Webhook Execution -> Handled idempotently
	errDup := subSvc.HandleSaaSHook(ctx, payload, "valid-mock-signature")
	if errDup != nil {
		t.Errorf("duplicate SaaS webhook should be handled idempotently, got %v", errDup)
	}
}

func TestSubscription_BookingLimitEnforcement(t *testing.T) {
	subSvc, org, _, _, apptRepo, _ := setupTestSubscriptionService(t)
	ctx := context.Background()

	// Assign Starter Plan (max_monthly_appointments = 100)
	_, _ = subSvc.UpgradePlan(ctx, org.ID, "starter", subscription.BillingPeriodMonthly)

	// Under 100 bookings -> CheckBookingLimit passes
	if err := subSvc.CheckBookingLimit(ctx, org.ID); err != nil {
		t.Errorf("expected booking limit check to pass under quota, got %v", err)
	}

	// Create 101 appointments for current month
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i <= 100; i++ {
		_, _ = apptRepo.Create(ctx, appointment.CreateAppointmentCmd{
			OrganizationID: org.ID,
			ServiceID:      uuid.New(),
			StaffID:        uuid.New(),
			StartTime:      monthStart.Add(time.Duration(i) * 10 * time.Minute),
			EndTime:        monthStart.Add(time.Duration(i)*10*time.Minute + 5*time.Minute),
			GuestName:      "Quota Client",
		})
	}

	// Check limit check for booking -> Should fail because 100 >= max_monthly_appointments (100)
	errQuota := subSvc.CheckBookingLimit(ctx, org.ID)
	if errQuota == nil {
		t.Error("SECURITY FAILURE: Expected booking quota limit error when 100th booking reached, got nil")
	}
}
