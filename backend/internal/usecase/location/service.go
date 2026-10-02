package location

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	locRepo   location.Repository
	auditRepo audit.Repository
}

func NewService(locRepo location.Repository, auditRepo audit.Repository) *Service {
	return &Service{
		locRepo:   locRepo,
		auditRepo: auditRepo,
	}
}

// ValidateTimezone checks if the string is a valid IANA timezone name.
func ValidateTimezone(tz string) error {
	if tz == "" {
		return nil
	}
	_, err := time.LoadLocation(tz)
	if err != nil {
		return apperror.ValidationFailed("invalid IANA timezone name: " + tz)
	}
	return nil
}

func (s *Service) CreateLocation(ctx context.Context, authCtx *rbac.AuthContext, cmd location.CreateLocationCmd) (*location.Location, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermLocationCreate) {
		return nil, apperror.Forbidden("permission location:create required")
	}

	if strings.TrimSpace(cmd.Name) == "" {
		return nil, apperror.ValidationFailed("location name is required")
	}

	if err := ValidateTimezone(cmd.Timezone); err != nil {
		return nil, err
	}

	loc, err := s.locRepo.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "location.created",
			ResourceType:   "location",
			ResourceID:     &loc.ID,
			CreatedAt:      time.Now(),
		})
	}

	return loc, nil
}

func (s *Service) GetLocation(ctx context.Context, authCtx *rbac.AuthContext, orgID, locationID uuid.UUID) (*location.Location, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermLocationRead) {
		return nil, apperror.Forbidden("permission location:read required")
	}
	return s.locRepo.GetByID(ctx, orgID, locationID)
}

func (s *Service) ListLocations(ctx context.Context, authCtx *rbac.AuthContext, orgID uuid.UUID) ([]*location.Location, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermLocationRead) {
		return nil, apperror.Forbidden("permission location:read required")
	}
	return s.locRepo.List(ctx, orgID)
}

func (s *Service) UpdateLocation(ctx context.Context, authCtx *rbac.AuthContext, cmd location.UpdateLocationCmd) (*location.Location, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermLocationUpdate) {
		return nil, apperror.Forbidden("permission location:update required")
	}

	if cmd.Timezone != nil {
		if err := ValidateTimezone(*cmd.Timezone); err != nil {
			return nil, err
		}
	}

	loc, err := s.locRepo.Update(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "location.updated",
			ResourceType:   "location",
			ResourceID:     &loc.ID,
			CreatedAt:      time.Now(),
		})
	}

	return loc, nil
}

func (s *Service) DeleteLocation(ctx context.Context, authCtx *rbac.AuthContext, orgID, locationID uuid.UUID) error {
	if authCtx == nil || authCtx.OrgID != orgID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermLocationDelete) {
		return apperror.Forbidden("permission location:delete required")
	}

	if err := s.locRepo.Delete(ctx, orgID, locationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "location.deleted",
			ResourceType:   "location",
			ResourceID:     &locationID,
			CreatedAt:      time.Now(),
		})
	}

	return nil
}
