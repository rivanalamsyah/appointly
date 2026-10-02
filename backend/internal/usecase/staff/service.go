package staff

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	staffRepo staff.Repository
	auditRepo audit.Repository
}

func NewService(staffRepo staff.Repository, auditRepo audit.Repository) *Service {
	return &Service{
		staffRepo: staffRepo,
		auditRepo: auditRepo,
	}
}

func (s *Service) CreateStaff(ctx context.Context, authCtx *rbac.AuthContext, cmd staff.CreateStaffCmd) (*staff.Staff, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffCreate) {
		return nil, apperror.Forbidden("permission staff:create required")
	}

	if strings.TrimSpace(cmd.FirstName) == "" || strings.TrimSpace(cmd.LastName) == "" {
		return nil, apperror.ValidationFailed("first_name and last_name are required")
	}

	st, err := s.staffRepo.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.created",
			ResourceType:   "staff",
			ResourceID:     &st.ID,
			CreatedAt:      time.Now(),
		})
	}

	return st, nil
}

func (s *Service) GetStaff(ctx context.Context, authCtx *rbac.AuthContext, orgID, staffID uuid.UUID) (*staff.Staff, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffRead) {
		return nil, apperror.Forbidden("permission staff:read required")
	}
	return s.staffRepo.GetByID(ctx, orgID, staffID)
}

func (s *Service) ListStaff(ctx context.Context, authCtx *rbac.AuthContext, filter staff.ListStaffFilter, page, perPage int) ([]*staff.Staff, int, error) {
	if authCtx == nil || authCtx.OrgID != filter.OrganizationID {
		return nil, 0, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffRead) {
		return nil, 0, apperror.Forbidden("permission staff:read required")
	}
	return s.staffRepo.List(ctx, filter, page, perPage)
}

func (s *Service) UpdateStaff(ctx context.Context, authCtx *rbac.AuthContext, cmd staff.UpdateStaffCmd) (*staff.Staff, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffUpdate) {
		return nil, apperror.Forbidden("permission staff:update required")
	}

	st, err := s.staffRepo.Update(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.updated",
			ResourceType:   "staff",
			ResourceID:     &st.ID,
			CreatedAt:      time.Now(),
		})
	}

	return st, nil
}

func (s *Service) SetStaffServices(ctx context.Context, authCtx *rbac.AuthContext, orgID, staffID uuid.UUID, serviceIDs []uuid.UUID) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffUpdate) {
		return apperror.Forbidden("permission staff:update required")
	}

	if err := s.staffRepo.SetStaffServices(ctx, orgID, staffID, serviceIDs); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.services_assigned",
			ResourceType:   "staff",
			ResourceID:     &staffID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}

func (s *Service) GetStaffServices(ctx context.Context, authCtx *rbac.AuthContext, orgID, staffID uuid.UUID) ([]*staff.StaffService, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffRead) {
		return nil, apperror.Forbidden("permission staff:read required")
	}
	return s.staffRepo.GetStaffServices(ctx, staffID)
}

func (s *Service) DeactivateStaff(ctx context.Context, authCtx *rbac.AuthContext, orgID, staffID uuid.UUID) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffDelete) {
		return apperror.Forbidden("permission staff:delete required")
	}

	if err := s.staffRepo.Delete(ctx, orgID, staffID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.deactivated",
			ResourceType:   "staff",
			ResourceID:     &staffID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}
