package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type AuditRepository struct {
	mu   sync.RWMutex
	logs map[uuid.UUID]*audit.AuditLog
}

func NewAuditRepository() *AuditRepository {
	return &AuditRepository{
		logs: make(map[uuid.UUID]*audit.AuditLog),
	}
}

func (r *AuditRepository) Create(ctx context.Context, log audit.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	cp := log
	r.logs[log.ID] = &cp
	return nil
}

func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*audit.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	l, exists := r.logs[id]
	if !exists {
		return nil, apperror.NotFound("audit log")
	}
	cp := *l
	return &cp, nil
}

func (r *AuditRepository) List(ctx context.Context, orgID uuid.UUID, filter audit.ListFilter, page, perPage int) ([]*audit.AuditLog, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*audit.AuditLog
	for _, l := range r.logs {
		if l.OrganizationID != nil && *l.OrganizationID == orgID {
			cp := *l
			result = append(result, &cp)
		}
	}
	total := len(result)
	start := (page - 1) * perPage
	if start >= total {
		return []*audit.AuditLog{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}
