package resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	repo resource.Repository
}

func NewService(repo resource.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateResource(ctx context.Context, cmd resource.CreateResourceCmd) (*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermResourceCreate) {
		return nil, apperror.Unauthorized("insufficient permissions to create resource")
	}

	cmd.OrganizationID = authCtx.OrgID
	return s.repo.CreateResource(ctx, cmd)
}

func (s *Service) GetResource(ctx context.Context, resourceID uuid.UUID) (*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermResourceRead) {
		return nil, apperror.Unauthorized("insufficient permissions to read resource")
	}

	return s.repo.GetResourceByID(ctx, authCtx.OrgID, resourceID)
}

func (s *Service) UpdateResource(ctx context.Context, cmd resource.UpdateResourceCmd) (*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermResourceUpdate) {
		return nil, apperror.Unauthorized("insufficient permissions to update resource")
	}

	return s.repo.UpdateResource(ctx, authCtx.OrgID, cmd)
}

func (s *Service) DeleteResource(ctx context.Context, resourceID uuid.UUID) error {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermResourceDelete) {
		return apperror.Unauthorized("insufficient permissions to delete resource")
	}

	return s.repo.DeleteResource(ctx, authCtx.OrgID, resourceID)
}

func (s *Service) ListResources(ctx context.Context, filter resource.ListResourcesFilter) ([]*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermResourceRead) {
		return nil, apperror.Unauthorized("insufficient permissions to list resources")
	}

	filter.OrganizationID = authCtx.OrgID
	return s.repo.ListResources(ctx, filter)
}

func (s *Service) AssignServiceResources(ctx context.Context, serviceID uuid.UUID, resourceIDs []uuid.UUID) ([]*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || (!authCtx.Can(rbac.PermResourceUpdate) && !authCtx.Can(rbac.PermServiceUpdate)) {
		return nil, apperror.Unauthorized("insufficient permissions to assign service resources")
	}

	cmd := resource.AssignServiceResourcesCmd{
		OrganizationID: authCtx.OrgID,
		ServiceID:      serviceID,
		ResourceIDs:    resourceIDs,
	}

	return s.repo.AssignServiceResources(ctx, cmd)
}

func (s *Service) GetServiceResources(ctx context.Context, serviceID uuid.UUID) ([]*resource.Resource, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || (!authCtx.Can(rbac.PermResourceRead) && !authCtx.Can(rbac.PermServiceRead)) {
		return nil, apperror.Unauthorized("insufficient permissions to get service resources")
	}

	return s.repo.GetServiceResources(ctx, authCtx.OrgID, serviceID)
}
