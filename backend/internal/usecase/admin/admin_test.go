package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/repository/memory"

	adminUseCase "github.com/appointly/appointly/backend/internal/usecase/admin"
)

func setupTestAdminService() (*adminUseCase.Service, *memory.OrgRepository, *memory.SubscriptionRepository, *memory.CustomerRepository, *memory.AuditRepository) {
	orgRepo := memory.NewOrgRepository()
	staffRepo := memory.NewStaffRepository()
	locRepo := memory.NewLocationRepository()
	apptRepo := memory.NewAppointmentRepository()
	subRepo := memory.NewSubscriptionRepository(staffRepo, locRepo, apptRepo)
	custRepo := memory.NewCustomerRepository()
	auditRepo := memory.NewAuditRepository()

	svc := adminUseCase.NewService(orgRepo, subRepo, custRepo, auditRepo)
	return svc, orgRepo, subRepo, custRepo, auditRepo
}

func TestSuperAdmin_AuthorizationProtection(t *testing.T) {
	handler := middleware.RequireSuperAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"super_admin_ok"}`))
	}))

	// 1. Regular organization user (Not Super Admin)
	t.Run("Regular Organization User Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/health", nil)
		authCtx := &rbac.AuthContext{
			UserID:       uuid.New(),
			OrgID:        uuid.New(),
			Role:         rbac.RoleOwner,
			IsOwner:      true,
			IsSuperAdmin: false, // NOT super admin!
		}
		ctx := rbac.NewContext(req.Context(), authCtx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for regular organization user, got %d", rec.Code)
		}
	})

	// 2. Super Admin User
	t.Run("Super Admin Allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/health", nil)
		authCtx := &rbac.AuthContext{
			UserID:       uuid.New(),
			OrgID:        uuid.New(),
			Role:         rbac.RoleAdmin,
			IsSuperAdmin: true, // Super Admin!
		}
		ctx := rbac.NewContext(req.Context(), authCtx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for super admin user, got %d", rec.Code)
		}
	})
}

func TestSuperAdmin_OrganizationSuspensionAndActivation(t *testing.T) {
	svc, orgRepo, _, _, auditRepo := setupTestAdminService()
	ctx := context.Background()

	org, err := orgRepo.Create(ctx, &organization.Organization{
		ID:     uuid.New(),
		Name:   "Target Tenant",
		Slug:   "target-tenant",
		Status: organization.StatusActive,
	})
	if err != nil {
		t.Fatalf("unexpected error creating org: %v", err)
	}

	// 1. Suspend Organization
	err = svc.SuspendOrganization(ctx, org.ID, "TOS violation")
	if err != nil {
		t.Fatalf("unexpected error suspending org: %v", err)
	}

	updated, _ := orgRepo.GetByID(ctx, org.ID)
	if updated.Status != organization.StatusSuspended {
		t.Fatalf("expected status suspended, got %s", updated.Status)
	}

	// Verify Audit Log
	logs, total, _ := auditRepo.List(ctx, org.ID, audit.ListFilter{}, 1, 10)
	if total < 1 || logs[0].Action != "admin.org_suspended" {
		t.Fatalf("expected audit log event admin.org_suspended")
	}

	// 2. Activate Organization
	err = svc.ActivateOrganization(ctx, org.ID)
	if err != nil {
		t.Fatalf("unexpected error activating org: %v", err)
	}

	reactivated, _ := orgRepo.GetByID(ctx, org.ID)
	if reactivated.Status != organization.StatusActive {
		t.Fatalf("expected status active, got %s", reactivated.Status)
	}
}

func TestSuperAdmin_AssignPlanTier(t *testing.T) {
	svc, orgRepo, subRepo, _, auditRepo := setupTestAdminService()
	ctx := context.Background()

	org, _ := orgRepo.Create(ctx, &organization.Organization{
		ID:   uuid.New(),
		Name: "SaaS Plan Test Org",
		Slug: "saas-plan-org",
	})

	err := svc.AssignPlanTier(ctx, org.ID, "professional")
	if err != nil {
		t.Fatalf("unexpected error assigning plan tier: %v", err)
	}

	sub, _ := subRepo.GetByOrganizationID(ctx, org.ID)
	if sub == nil || sub.Plan == nil || sub.Plan.Slug != "professional" {
		t.Fatalf("expected plan professional assigned to organization")
	}

	// Verify Audit Log
	logs, total, _ := auditRepo.List(ctx, org.ID, audit.ListFilter{}, 1, 10)
	if total < 1 || logs[0].Action != "admin.plan_assigned" {
		t.Fatalf("expected audit log event admin.plan_assigned")
	}
}

func TestSuperAdmin_BreakGlassCustomerAccess(t *testing.T) {
	svc, orgRepo, _, custRepo, auditRepo := setupTestAdminService()
	ctx := context.Background()

	org, _ := orgRepo.Create(ctx, &organization.Organization{
		ID:   uuid.New(),
		Name: "Break Glass Org",
		Slug: "breakglass-org",
	})

	cust, _ := custRepo.Create(ctx, customer.CreateCustomerCmd{
		OrganizationID: org.ID,
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@doe.com",
	})

	// 1. Missing reason rejected
	_, err := svc.BreakGlassCustomerAccess(ctx, org.ID, cust.ID, "")
	if err == nil {
		t.Fatalf("expected error when break-glass reason is missing")
	}

	// 2. Valid break-glass access
	inspected, err := svc.BreakGlassCustomerAccess(ctx, org.ID, cust.ID, "Support Ticket #9921")
	if err != nil || inspected == nil {
		t.Fatalf("expected successful break-glass customer retrieval")
	}

	// Verify mandatory audit log entry
	logs, total, _ := auditRepo.List(ctx, org.ID, audit.ListFilter{}, 1, 10)
	if total < 1 || logs[0].Action != "admin.breakglass_customer_access" {
		t.Fatalf("expected audit log event admin.breakglass_customer_access")
	}
}
