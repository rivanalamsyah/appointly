package location_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/location"
)

func setupLocationService() (*usecase.Service, *memory.LocationRepository) {
	locRepo := memory.NewLocationRepository()
	auditRepo := memory.NewAuditRepository()
	svc := usecase.NewService(locRepo, auditRepo)
	return svc, locRepo
}

func TestCreateLocation_Success(t *testing.T) {
	svc, _ := setupLocationService()
	ctx := context.Background()

	orgID := uuid.New()
	userID := uuid.New()
	authCtx := &rbac.AuthContext{
		UserID:  userID,
		OrgID:   orgID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	cmd := location.CreateLocationCmd{
		OrganizationID: orgID,
		Name:           "Downtown Headquarters",
		AddressLine1:   "123 Main Street",
		City:           "New York",
		Country:        "US",
		Timezone:       "America/New_York",
	}

	loc, err := svc.CreateLocation(ctx, authCtx, cmd)
	if err != nil {
		t.Fatalf("expected no error creating location, got %v", err)
	}

	if loc.Name != "Downtown Headquarters" {
		t.Errorf("expected location name Downtown Headquarters, got %s", loc.Name)
	}
	if !loc.IsDefault {
		t.Errorf("first location created for organization should default to IsDefault=true")
	}
}

func TestCreateLocation_InvalidTimezone(t *testing.T) {
	svc, _ := setupLocationService()
	ctx := context.Background()

	orgID := uuid.New()
	authCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	cmd := location.CreateLocationCmd{
		OrganizationID: orgID,
		Name:           "Invalid TZ Branch",
		Timezone:       "NonExistent/Timezone_Location",
	}

	_, err := svc.CreateLocation(ctx, authCtx, cmd)
	if err == nil {
		t.Fatalf("EXPECTED TIMEZONE VALIDATION ERROR for invalid IANA timezone string, got nil")
	}
}

func TestLocation_TenantIsolation(t *testing.T) {
	svc, _ := setupLocationService()
	ctx := context.Background()

	org1ID := uuid.New()
	authCtx1 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org1ID, Role: rbac.RoleOwner}
	loc1, _ := svc.CreateLocation(ctx, authCtx1, location.CreateLocationCmd{
		OrganizationID: org1ID,
		Name:           "Branch Org 1",
	})

	org2ID := uuid.New()
	authCtx2 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org2ID, Role: rbac.RoleOwner}

	// User 2 from Org 2 attempts to read location of Org 1
	_, err := svc.GetLocation(ctx, authCtx2, org1ID, loc1.ID)
	if err == nil {
		t.Fatalf("EXPECTED CROSS-TENANT ERROR when accessing location from different org, got nil")
	}
}

func TestLocation_PermissionDenial(t *testing.T) {
	svc, _ := setupLocationService()
	ctx := context.Background()

	orgID := uuid.New()
	// User has STAFF role (STAFF lacks PermLocationCreate and PermLocationDelete)
	staffAuthCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleStaff,
		IsOwner: false,
	}

	cmd := location.CreateLocationCmd{
		OrganizationID: orgID,
		Name:           "Unauthorized Branch",
	}

	_, err := svc.CreateLocation(ctx, staffAuthCtx, cmd)
	if err == nil {
		t.Fatalf("EXPECTED PERMISSION DENIED error for STAFF role creating location, got nil")
	}
}
