package org

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	orgRepo    organization.Repository
	memberRepo rbac.Repository
	auditRepo  audit.Repository
}

func NewService(orgRepo organization.Repository, memberRepo rbac.Repository, auditRepo audit.Repository) *Service {
	return &Service{
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		auditRepo:  auditRepo,
	}
}

func (s *Service) CreateOrganization(ctx context.Context, ownerID uuid.UUID, cmd organization.CreateOrganizationCmd) (*organization.Organization, *rbac.OrganizationMember, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, nil, apperror.ValidationFailed("organization name is required")
	}

	slug := strings.ToLower(strings.TrimSpace(cmd.Slug))
	if slug == "" {
		// Generate default slug from name
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	if !organization.SlugValidator(slug) {
		return nil, nil, apperror.ValidationFailed("invalid slug format")
	}

	exists, err := s.orgRepo.SlugExists(ctx, slug)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, apperror.AlreadyExists("organization with slug " + slug)
	}

	org := &organization.Organization{
		ID:           uuid.New(),
		Name:         name,
		Slug:         slug,
		BusinessType: cmd.BusinessType,
		Email:        cmd.Email,
		Phone:        cmd.Phone,
		Timezone:     cmd.Timezone,
		Currency:     cmd.Currency,
		Country:      cmd.Country,
		Status:       organization.StatusActive,
		OwnerID:      ownerID,
		Settings:     organization.DefaultSettings(),
	}

	if org.Timezone == "" {
		org.Timezone = "UTC"
	}
	if org.Currency == "" {
		org.Currency = "USD"
	}
	if org.Country == "" {
		org.Country = "US"
	}
	if org.BusinessType == "" {
		org.BusinessType = organization.BusinessTypeOther
	}

	createdOrg, err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, nil, err
	}

	// Add creator as OWNER member
	now := time.Now()
	ownerMember := &rbac.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: createdOrg.ID,
		UserID:         ownerID,
		Role:           rbac.RoleOwner,
		JoinedAt:       &now,
		IsActive:       true,
	}

	createdMember, err := s.memberRepo.CreateMember(ctx, ownerMember)
	if err != nil {
		return nil, nil, err
	}

	// Audit log
	if s.auditRepo != nil {
		orgID := createdOrg.ID
		actorID := ownerID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			ActorType:      "user",
			Action:         "org.created",
			ResourceType:   "organization",
			ResourceID:     &orgID,
			CreatedAt:      time.Now(),
		})
	}

	return createdOrg, createdMember, nil
}

func (s *Service) GetOrganization(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID) (*organization.Organization, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("access denied to tenant organization")
	}
	if !authCtx.Can(rbac.PermOrgRead) {
		return nil, apperror.Forbidden("insufficient permissions to view organization")
	}
	return s.orgRepo.GetByID(ctx, orgID)
}

func (s *Service) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error) {
	// Simple lookup for user's organizations
	// In real setup, we retrieve memberships for user and fetch org objects
	var result []*organization.Organization
	orgs, _, err := s.orgRepo.List(ctx, organization.ListOrganizationsFilter{}, 1, 100)
	if err != nil {
		return nil, err
	}
	for _, o := range orgs {
		mem, err := s.memberRepo.GetMemberByUserAndOrg(ctx, userID, o.ID)
		if err == nil && mem != nil && mem.IsActive {
			result = append(result, o)
		}
	}
	return result, nil
}

func (s *Service) ListMembers(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID) ([]*rbac.OrganizationMember, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermMemberRead) {
		return nil, apperror.Forbidden("permission member:read required")
	}
	return s.memberRepo.ListMembersByOrg(ctx, orgID)
}

func (s *Service) AddMember(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID, targetUserID uuid.UUID, role rbac.Role) (*rbac.OrganizationMember, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermMemberInvite) {
		return nil, apperror.Forbidden("permission member:invite required")
	}

	now := time.Now()
	newMember := &rbac.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         targetUserID,
		Role:           role,
		InvitedBy:      &authCtx.UserID,
		InvitedAt:      &now,
		JoinedAt:       &now,
		IsActive:       true,
	}

	created, err := s.memberRepo.CreateMember(ctx, newMember)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "member.added",
			ResourceType:   "organization_member",
			ResourceID:     &created.ID,
			CreatedAt:      time.Now(),
		})
	}

	return created, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID, memberID uuid.UUID, newRole rbac.Role) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermMemberUpdate) {
		return apperror.Forbidden("permission member:update required")
	}

	targetMember, err := s.memberRepo.GetMemberByID(ctx, memberID)
	if err != nil || targetMember.OrganizationID != orgID {
		return apperror.NotFound("organization member")
	}

	if err := s.memberRepo.UpdateMemberRole(ctx, memberID, newRole); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "member.role_updated",
			ResourceType:   "organization_member",
			ResourceID:     &memberID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}

func (s *Service) RemoveMember(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID, memberID uuid.UUID) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermMemberRemove) {
		return apperror.Forbidden("permission member:remove required")
	}

	targetMember, err := s.memberRepo.GetMemberByID(ctx, memberID)
	if err != nil || targetMember.OrganizationID != orgID {
		return apperror.NotFound("organization member")
	}

	if err := s.memberRepo.DeactivateMember(ctx, memberID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "member.deactivated",
			ResourceType:   "organization_member",
			ResourceID:     &memberID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}
