package resource_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	resourceuc "github.com/appointly/appointly/backend/internal/usecase/resource"
)

func setupTestResourceService() (*resourceuc.Service, *memory.ResourceRepository, uuid.UUID, uuid.UUID) {
	repo := memory.NewResourceRepository()
	service := resourceuc.NewService(repo)
	orgID := uuid.New()
	userID := uuid.New()

	return service, repo, orgID, userID
}

func contextWithAuth(orgID, userID uuid.UUID, role rbac.Role) context.Context {
	return rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  userID,
		OrgID:   orgID,
		Role:    role,
		IsOwner: role == rbac.RoleOwner,
	})
}

func TestResource_CreateAndGet(t *testing.T) {
	svc, _, orgID, userID := setupTestResourceService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	cmd := resource.CreateResourceCmd{
		Name:        "Treatment Room 1",
		Type:        resource.TypeRoom,
		Capacity:    2,
		Status:      resource.StatusActive,
		Description: "Main facial room",
	}

	created, err := svc.CreateResource(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.Name != "Treatment Room 1" {
		t.Errorf("expected name Treatment Room 1, got %s", created.Name)
	}
	if created.Capacity != 2 {
		t.Errorf("expected capacity 2, got %d", created.Capacity)
	}

	retrieved, err := svc.GetResource(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error on GetResource, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected resource ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestResource_TenantIsolation(t *testing.T) {
	svc, _, orgA, userA := setupTestResourceService()
	orgB := uuid.New()

	ctxA := contextWithAuth(orgA, userA, rbac.RoleOwner)
	ctxB := contextWithAuth(orgB, uuid.New(), rbac.RoleOwner)

	created, err := svc.CreateResource(ctxA, resource.CreateResourceCmd{
		Name:     "Laser Machine",
		Type:     resource.TypeEquipment,
		Capacity: 1,
		Status:   resource.StatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	// Org B attempting to read Org A's resource
	_, err = svc.GetResource(ctxB, created.ID)
	if err == nil {
		t.Error("expected error when Org B attempts to access Org A resource, got nil")
	}

	// Org B listing resources should return 0 items
	listB, err := svc.ListResources(ctxB, resource.ListResourcesFilter{})
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}
	if len(listB) != 0 {
		t.Errorf("expected 0 resources for Org B, got %d", len(listB))
	}
}

func TestResource_PermissionRules(t *testing.T) {
	svc, _, orgID, userID := setupTestResourceService()

	// Staff role does not have resource:create permission
	ctxStaff := contextWithAuth(orgID, userID, rbac.RoleStaff)

	_, err := svc.CreateResource(ctxStaff, resource.CreateResourceCmd{
		Name: "Forbidden Chair",
		Type: resource.TypeChair,
	})
	if err == nil {
		t.Error("expected error for staff role creating resource, got nil")
	}
}

func TestResource_ServiceAssignment(t *testing.T) {
	svc, _, orgID, userID := setupTestResourceService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	res1, err := svc.CreateResource(ctx, resource.CreateResourceCmd{
		Name:     "Pedicure Chair 1",
		Type:     resource.TypeChair,
		Capacity: 1,
	})
	if err != nil {
		t.Fatalf("failed to create res1: %v", err)
	}

	res2, err := svc.CreateResource(ctx, resource.CreateResourceCmd{
		Name:     "UV Nail Lamp",
		Type:     resource.TypeEquipment,
		Capacity: 1,
	})
	if err != nil {
		t.Fatalf("failed to create res2: %v", err)
	}

	serviceID := uuid.New()
	assigned, err := svc.AssignServiceResources(ctx, serviceID, []uuid.UUID{res1.ID, res2.ID})
	if err != nil {
		t.Fatalf("failed to assign resources to service: %v", err)
	}
	if len(assigned) != 2 {
		t.Errorf("expected 2 assigned resources, got %d", len(assigned))
	}

	svcRes, err := svc.GetServiceResources(ctx, serviceID)
	if err != nil {
		t.Fatalf("failed to get service resources: %v", err)
	}
	if len(svcRes) != 2 {
		t.Errorf("expected 2 service resources, got %d", len(svcRes))
	}
}

func TestResource_AvailabilityContractAndMaintenanceStatus(t *testing.T) {
	svc, repo, orgID, userID := setupTestResourceService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	// Create resource in maintenance status
	maintRes, err := svc.CreateResource(ctx, resource.CreateResourceCmd{
		Name:     "Broken Scanner",
		Type:     resource.TypeEquipment,
		Capacity: 1,
		Status:   resource.StatusMaintenance,
	})
	if err != nil {
		t.Fatalf("failed to create maintenance resource: %v", err)
	}

	now := time.Now()
	later := now.Add(1 * time.Hour)

	avail, err := repo.IsResourceAvailable(context.Background(), orgID, maintRes.ID, now, later)
	if err != nil {
		t.Fatalf("error checking availability: %v", err)
	}
	if avail {
		t.Error("expected resource in MAINTENANCE status to be unavailable, got available = true")
	}

	// Assign maintenance resource to service and check available list
	serviceID := uuid.New()
	_, _ = svc.AssignServiceResources(ctx, serviceID, []uuid.UUID{maintRes.ID})

	availList, err := repo.GetAvailableResourcesForService(context.Background(), orgID, serviceID, nil, now, later)
	if err != nil {
		t.Fatalf("error getting available resources for service: %v", err)
	}
	if len(availList) != 0 {
		t.Errorf("expected 0 available resources for service due to maintenance status, got %d", len(availList))
	}
}
