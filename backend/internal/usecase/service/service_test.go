package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/service"
)

func setupServiceCatalog() (*usecase.Service, *memory.ServiceRepository) {
	svcRepo := memory.NewServiceRepository()
	auditRepo := memory.NewAuditRepository()
	svc := usecase.NewService(svcRepo, auditRepo)
	return svc, svcRepo
}

func TestCreateService_ValidationRules(t *testing.T) {
	svc, _ := setupServiceCatalog()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	// Test 1: Duration <= 0 rejected
	_, err := svc.CreateService(ctx, authCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Zero Duration Service",
		DurationMinutes: 0, // Invalid!
	})
	if err == nil {
		t.Errorf("EXPECTED VALIDATION ERROR for zero duration, got nil")
	}

	// Test 2: Negative buffer rejected
	_, err = svc.CreateService(ctx, authCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Negative Buffer Service",
		DurationMinutes: 30,
		BufferBefore:    -10, // Invalid!
	})
	if err == nil {
		t.Errorf("EXPECTED VALIDATION ERROR for negative buffer, got nil")
	}

	// Test 3: Negative price rejected
	_, err = svc.CreateService(ctx, authCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Negative Price Service",
		DurationMinutes: 30,
		PriceCents:      -500, // Invalid!
	})
	if err == nil {
		t.Errorf("EXPECTED VALIDATION ERROR for negative price, got nil")
	}

	// Test 4: Valid Service Success
	validSvc, err := svc.CreateService(ctx, authCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Signature Haircut",
		DurationMinutes: 45,
		BufferBefore:    5,
		BufferAfter:     10,
		PriceCents:      6500,
		Currency:        "USD",
		IsPublic:        true,
	})
	if err != nil {
		t.Fatalf("expected valid service creation to succeed, got %v", err)
	}

	if validSvc.TotalDurationMinutes() != 60 { // 45 + 5 + 10 = 60
		t.Errorf("expected total duration 60, got %d", validSvc.TotalDurationMinutes())
	}
}

func TestService_TenantIsolation(t *testing.T) {
	svc, _ := setupServiceCatalog()
	ctx := context.Background()

	org1ID := uuid.New()
	authCtx1 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org1ID, Role: rbac.RoleOwner}
	svc1, _ := svc.CreateService(ctx, authCtx1, service.CreateServiceCmd{
		OrganizationID:  org1ID,
		Name:            "Org 1 Cut",
		DurationMinutes: 30,
	})

	org2ID := uuid.New()
	authCtx2 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org2ID, Role: rbac.RoleOwner}

	// User 2 from Org 2 attempts to fetch Service 1 from Org 1
	_, err := svc.GetService(ctx, authCtx2, org1ID, svc1.ID)
	if err == nil {
		t.Fatalf("EXPECTED CROSS-TENANT SECURITY ERROR, got nil")
	}
}

func TestService_RBACPermissionDenial(t *testing.T) {
	svc, _ := setupServiceCatalog()
	ctx := context.Background()

	orgID := uuid.New()
	// User has STAFF role (STAFF lacks PermServiceCreate)
	staffAuthCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleStaff,
		IsOwner: false,
	}

	_, err := svc.CreateService(ctx, staffAuthCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Unauthorized Service",
		DurationMinutes: 30,
	})
	if err == nil {
		t.Fatalf("EXPECTED PERMISSION DENIED error for STAFF role creating service, got nil")
	}
}

func TestService_StatusToggleBehavior(t *testing.T) {
	svc, repo := setupServiceCatalog()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: orgID, Role: rbac.RoleOwner}

	s1, err := svc.CreateService(ctx, authCtx, service.CreateServiceCmd{
		OrganizationID:  orgID,
		Name:            "Seasonal Treatment",
		DurationMinutes: 60,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Toggle to Inactive
	updated, err := svc.ToggleServiceStatus(ctx, authCtx, orgID, s1.ID, service.StatusInactive)
	if err != nil {
		t.Fatalf("toggle failed: %v", err)
	}
	if updated.Status != service.StatusInactive {
		t.Errorf("expected status INACTIVE, got %s", updated.Status)
	}

	// Historical query still returns entity by ID
	fetched, err := repo.GetServiceByID(ctx, orgID, s1.ID)
	if err != nil || fetched.ID != s1.ID {
		t.Errorf("historical appointment lookup must retain service record even when inactive")
	}
}
