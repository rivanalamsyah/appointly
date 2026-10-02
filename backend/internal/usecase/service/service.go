package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	svcRepo   service.Repository
	auditRepo audit.Repository
}

func NewService(svcRepo service.Repository, auditRepo audit.Repository) *Service {
	return &Service{
		svcRepo:   svcRepo,
		auditRepo: auditRepo,
	}
}

func (s *Service) CreateService(ctx context.Context, authCtx *rbac.AuthContext, cmd service.CreateServiceCmd) (*service.Service, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceCreate) {
		return nil, apperror.Forbidden("permission service:create required")
	}

	if strings.TrimSpace(cmd.Name) == "" {
		return nil, apperror.ValidationFailed("service name is required")
	}

	if cmd.DurationMinutes <= 0 {
		return nil, apperror.ValidationFailed("duration_minutes must be greater than 0")
	}

	if cmd.BufferBefore < 0 || cmd.BufferAfter < 0 {
		return nil, apperror.ValidationFailed("buffer times cannot be negative")
	}

	if cmd.PriceCents < 0 {
		return nil, apperror.ValidationFailed("price_cents cannot be negative")
	}

	if cmd.CategoryID != nil {
		cat, err := s.svcRepo.GetCategoryByID(ctx, cmd.OrganizationID, *cmd.CategoryID)
		if err != nil || !cat.IsActive {
			return nil, apperror.ValidationFailed("invalid or inactive service category")
		}
	}

	svc, err := s.svcRepo.CreateService(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "service.created",
			ResourceType:   "service",
			ResourceID:     &svc.ID,
			CreatedAt:      time.Now(),
		})
	}

	return svc, nil
}

func (s *Service) GetService(ctx context.Context, authCtx *rbac.AuthContext, orgID, serviceID uuid.UUID) (*service.Service, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceRead) {
		return nil, apperror.Forbidden("permission service:read required")
	}
	return s.svcRepo.GetServiceByID(ctx, orgID, serviceID)
}

func (s *Service) ListServices(ctx context.Context, authCtx *rbac.AuthContext, filter service.ListServicesFilter, page, perPage int) ([]*service.Service, int, error) {
	if authCtx == nil || authCtx.OrgID != filter.OrganizationID {
		return nil, 0, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceRead) {
		return nil, 0, apperror.Forbidden("permission service:read required")
	}
	return s.svcRepo.ListServices(ctx, filter, page, perPage)
}

func (s *Service) UpdateService(ctx context.Context, authCtx *rbac.AuthContext, cmd service.UpdateServiceCmd) (*service.Service, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceUpdate) {
		return nil, apperror.Forbidden("permission service:update required")
	}

	if cmd.DurationMinutes != nil && *cmd.DurationMinutes <= 0 {
		return nil, apperror.ValidationFailed("duration_minutes must be greater than 0")
	}

	if (cmd.BufferBefore != nil && *cmd.BufferBefore < 0) || (cmd.BufferAfter != nil && *cmd.BufferAfter < 0) {
		return nil, apperror.ValidationFailed("buffer times cannot be negative")
	}

	if cmd.PriceCents != nil && *cmd.PriceCents < 0 {
		return nil, apperror.ValidationFailed("price_cents cannot be negative")
	}

	svc, err := s.svcRepo.UpdateService(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "service.updated",
			ResourceType:   "service",
			ResourceID:     &svc.ID,
			CreatedAt:      time.Now(),
		})
	}

	return svc, nil
}

func (s *Service) ToggleServiceStatus(ctx context.Context, authCtx *rbac.AuthContext, orgID, serviceID uuid.UUID, newStatus service.Status) (*service.Service, error) {
	cmd := service.UpdateServiceCmd{
		ID:             serviceID,
		OrganizationID: orgID,
		Status:         &newStatus,
	}
	return s.UpdateService(ctx, authCtx, cmd)
}

func (s *Service) DeleteService(ctx context.Context, authCtx *rbac.AuthContext, orgID, serviceID uuid.UUID) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceDelete) {
		return apperror.Forbidden("permission service:delete required")
	}

	if err := s.svcRepo.DeleteService(ctx, orgID, serviceID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "service.deleted",
			ResourceType:   "service",
			ResourceID:     &serviceID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}

// --- Category Methods ---

func (s *Service) CreateCategory(ctx context.Context, authCtx *rbac.AuthContext, cmd service.CreateCategoryCmd) (*service.ServiceCategory, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceCreate) {
		return nil, apperror.Forbidden("permission service:create required")
	}

	if strings.TrimSpace(cmd.Name) == "" {
		return nil, apperror.ValidationFailed("category name is required")
	}

	return s.svcRepo.CreateCategory(ctx, cmd)
}

func (s *Service) ListCategories(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID) ([]*service.ServiceCategory, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermServiceRead) {
		return nil, apperror.Forbidden("permission service:read required")
	}
	return s.svcRepo.ListCategories(ctx, orgID)
}
