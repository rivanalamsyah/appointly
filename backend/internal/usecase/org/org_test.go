package org_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/org"
)

func setupOrgService() (*usecase.Service, *memory.OrgRepository, *memory.MemberRepository, *memory.AuditRepository) {
	orgRepo := memory.NewOrgRepository()
	memberRepo := memory.NewMemberRepository()
	auditRepo := memory.NewAuditRepository()
	svc := usecase.NewService(orgRepo, memberRepo, auditRepo)
	return svc, orgRepo, memberRepo, auditRepo
}

func TestCreateOrganization_Success(t *testing.T) {
	svc, _, memberRepo, _ := setupOrgService()
	ctx := context.Background()

	ownerID := uuid.New()
	cmd := organization.CreateOrganizationCmd{
		Name:         "Barber Elite",
		Slug:         "barber-elite",
		BusinessType: organization.BusinessTypeSalon,
		Email:        "contact@barberelite.com",
	}

	org, member, err := svc.CreateOrganization(ctx, ownerID, cmd)
	if err != nil {
		t.Fatalf("expected no error creating org, got %v", err)
	}

	if org.Name != "Barber Elite" || org.Slug != "barber-elite" {
		t.Errorf("unexpected org details: %+v", org)
	}

	if member.Role != rbac.RoleOwner {
		t.Errorf("expected creator role to be OWNER, got %s", member.Role)
	}

	m, err := memberRepo.GetMemberByUserAndOrg(ctx, ownerID, org.ID)
	if err != nil || m.Role != rbac.RoleOwner {
		t.Errorf("expected owner membership in DB, got err=%v", err)
	}
}

func TestTenantIsolation_CrossTenantAccessAttempt(t *testing.T) {
	svc, _, _, _ := setupOrgService()
	ctx := context.Background()

	owner1ID := uuid.New()
	org1, _, _ := svc.CreateOrganization(ctx, owner1ID, organization.CreateOrganizationCmd{
		Name: "Org Alpha",
		Slug: "org-alpha",
	})

	owner2ID := uuid.New()
	org2, _, _ := svc.CreateOrganization(ctx, owner2ID, organization.CreateOrganizationCmd{
		Name: "Org Beta",
		Slug: "org-beta",
	})

	// User 1 attempts to access Org 2
	user1AuthCtx := &rbac.AuthContext{
		UserID:  owner1ID,
		OrgID:   org1.ID, // User 1's active org is Org 1
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	// Read members of Org 2 using User 1's context
	_, err := svc.ListMembers(ctx, user1AuthCtx, org2.ID)
	if err == nil {
		t.Fatalf("EXPECTED SECURITY ERROR on cross-tenant access attempt, got nil")
	}
}

func TestRBAC_PermissionDenial(t *testing.T) {
	svc, _, _, _ := setupOrgService()
	ctx := context.Background()

	ownerID := uuid.New()
	org, _, _ := svc.CreateOrganization(ctx, ownerID, organization.CreateOrganizationCmd{
		Name: "Clinic Health",
		Slug: "clinic-health",
	})

	staffUserID := uuid.New()
	// User has STAFF role in Org
	staffAuthCtx := &rbac.AuthContext{
		UserID:  staffUserID,
		OrgID:   org.ID,
		Role:    rbac.RoleStaff, // STAFF role does NOT have PermMemberInvite or PermMemberUpdate
		IsOwner: false,
	}

	targetMemberID := uuid.New()
	// STAFF attempts to update a member role
	err := svc.UpdateMemberRole(ctx, staffAuthCtx, org.ID, targetMemberID, rbac.RoleAdmin)
	if err == nil {
		t.Fatalf("EXPECTED PERMISSION DENIED error for STAFF role attempting admin action, got nil")
	}
}
