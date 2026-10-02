package staff_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/staff"
)

func setupStaffService() (*usecase.Service, *memory.StaffRepository) {
	staffRepo := memory.NewStaffRepository()
	auditRepo := memory.NewAuditRepository()
	svc := usecase.NewService(staffRepo, auditRepo)
	return svc, staffRepo
}

func TestCreateStaff_Success(t *testing.T) {
	svc, _ := setupStaffService()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	cmd := staff.CreateStaffCmd{
		OrganizationID: orgID,
		FirstName:      "Alexander",
		LastName:       "Wright",
		Title:          "Master Barber",
		Email:          "alex@barber.com",
		AcceptsOnline:  true,
	}

	st, err := svc.CreateStaff(ctx, authCtx, cmd)
	if err != nil {
		t.Fatalf("expected no error creating staff, got %v", err)
	}

	if st.FullName() != "Alexander Wright" {
		t.Errorf("expected full name Alexander Wright, got %s", st.FullName())
	}
	if !st.IsActive {
		t.Errorf("expected newly created staff to be active")
	}
}

func TestStaff_ManyToManyServiceAssignment(t *testing.T) {
	svc, _ := setupStaffService()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: orgID, Role: rbac.RoleOwner}

	st1, _ := svc.CreateStaff(ctx, authCtx, staff.CreateStaffCmd{
		OrganizationID: orgID,
		FirstName:      "Elena",
		LastName:       "Rostova",
	})

	svc1ID := uuid.New()
	svc2ID := uuid.New()

	// Assign 2 services to Elena
	err := svc.SetStaffServices(ctx, authCtx, orgID, st1.ID, []uuid.UUID{svc1ID, svc2ID})
	if err != nil {
		t.Fatalf("expected successful service assignment, got %v", err)
	}

	assigned, err := svc.GetStaffServices(ctx, authCtx, orgID, st1.ID)
	if err != nil {
		t.Fatalf("failed to retrieve staff services: %v", err)
	}

	if len(assigned) != 2 {
		t.Errorf("expected 2 assigned services for staff, got %d", len(assigned))
	}
}

func TestStaff_TenantIsolation(t *testing.T) {
	svc, _ := setupStaffService()
	ctx := context.Background()

	org1ID := uuid.New()
	authCtx1 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org1ID, Role: rbac.RoleOwner}
	st1, _ := svc.CreateStaff(ctx, authCtx1, staff.CreateStaffCmd{
		OrganizationID: org1ID,
		FirstName:      "Org 1 Staff",
		LastName:       "Member",
	})

	org2ID := uuid.New()
	authCtx2 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org2ID, Role: rbac.RoleOwner}

	// User 2 from Org 2 attempts to view Staff from Org 1
	_, err := svc.GetStaff(ctx, authCtx2, org1ID, st1.ID)
	if err == nil {
		t.Fatalf("EXPECTED CROSS-TENANT SECURITY ERROR on staff lookup, got nil")
	}
}

func TestStaff_RBACPermissionDenial(t *testing.T) {
	svc, _ := setupStaffService()
	ctx := context.Background()

	orgID := uuid.New()
	// User has STAFF role (STAFF lacks PermStaffCreate)
	staffAuthCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleStaff,
		IsOwner: false,
	}

	_, err := svc.CreateStaff(ctx, staffAuthCtx, staff.CreateStaffCmd{
		OrganizationID: orgID,
		FirstName:      "Unauthorized",
		LastName:       "Staff",
	})
	if err == nil {
		t.Fatalf("EXPECTED PERMISSION DENIED error for STAFF role creating staff, got nil")
	}
}

func TestStaff_InactiveHistoricalDataPreservation(t *testing.T) {
	svc, repo := setupStaffService()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: orgID, Role: rbac.RoleOwner}

	st, err := svc.CreateStaff(ctx, authCtx, staff.CreateStaffCmd{
		OrganizationID: orgID,
		FirstName:      "Retired",
		LastName:       "Stylist",
	})
	if err != nil {
		t.Fatalf("create staff failed: %v", err)
	}

	// Deactivate staff
	err = svc.DeactivateStaff(ctx, authCtx, orgID, st.ID)
	if err != nil {
		t.Fatalf("deactivate staff failed: %v", err)
	}

	// Historical lookup by ID MUST still succeed for appointment record integrity
	fetched, err := repo.GetByID(ctx, orgID, st.ID)
	if err != nil || fetched.ID != st.ID {
		t.Errorf("historical appointment lookup must retain staff record even when deactivated")
	}
	if fetched.IsActive {
		t.Errorf("expected staff status to be IsActive=false after deactivation")
	}
}
