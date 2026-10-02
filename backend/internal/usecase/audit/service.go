package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// Service manages audit log queries for tenant organizations.
// Audit logs are strictly immutable — no Update or Delete operations are exposed.
type Service struct {
	repo audit.Repository
}

// NewService constructs a new Audit Service.
func NewService(repo audit.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ListLogs retrieves paginated audit log records for an organization.
func (s *Service) ListLogs(ctx context.Context, orgID uuid.UUID, filter audit.ListFilter, page, perPage int) ([]*audit.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.repo.List(ctx, orgID, filter, page, perPage)
}

// GetLogByID retrieves a specific audit record by ID with organization verification.
func (s *Service) GetLogByID(ctx context.Context, orgID, logID uuid.UUID) (*audit.AuditLog, error) {
	entry, err := s.repo.GetByID(ctx, logID)
	if err != nil {
		return nil, err
	}
	if entry.OrganizationID != nil && *entry.OrganizationID != orgID {
		return nil, apperror.Forbidden("access denied to audit log from another tenant")
	}
	return entry, nil
}
