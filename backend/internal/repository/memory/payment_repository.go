package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/payment"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// PaymentRepository is an in-memory thread-safe implementation of payment.Repository.
type PaymentRepository struct {
	mu                sync.RWMutex
	payments          map[uuid.UUID]*payment.Payment
	refunds           map[uuid.UUID]*payment.Refund
	processedWebhooks map[string]bool // idempotency tracker for webhook payloads
}

// NewPaymentRepository constructs a new in-memory PaymentRepository.
func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{
		payments:          make(map[uuid.UUID]*payment.Payment),
		refunds:           make(map[uuid.UUID]*payment.Refund),
		processedWebhooks: make(map[string]bool),
	}
}

// Create persists a new payment.
func (r *PaymentRepository) Create(ctx context.Context, cmd payment.CreatePaymentCmd) (*payment.Payment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	payID := uuid.New()

	p := &payment.Payment{
		ID:             payID,
		OrganizationID: cmd.OrganizationID,
		AppointmentID:  cmd.AppointmentID,
		CustomerID:     cmd.CustomerID,
		Status:         payment.StatusPending,
		Provider:       cmd.Provider,
		AmountCents:    cmd.AmountCents,
		Currency:       cmd.Currency,
		ExpiresAt:      cmd.ExpiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.payments[payID] = p
	return p, nil
}

// GetByID retrieves a payment by ID.
func (r *PaymentRepository) GetByID(ctx context.Context, orgID, paymentID uuid.UUID) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.payments[paymentID]
	if !exists || (orgID != uuid.Nil && p.OrganizationID != orgID) {
		return nil, apperror.NotFound("payment")
	}

	cp := *p
	return &cp, nil
}

// GetByAppointmentID retrieves a payment associated with an appointment.
func (r *PaymentRepository) GetByAppointmentID(ctx context.Context, orgID, appointmentID uuid.UUID) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.AppointmentID == appointmentID && (orgID == uuid.Nil || p.OrganizationID == orgID) {
			cp := *p
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("payment")
}

// GetByExternalID retrieves a payment by external gateway ID.
func (r *PaymentRepository) GetByExternalID(ctx context.Context, externalID string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.ExternalID == externalID || p.ExternalReference == externalID {
			cp := *p
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("payment")
}

// UpdateStatus updates the status of a payment.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status payment.Status, paidAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.payments[paymentID]
	if !exists {
		return apperror.NotFound("payment")
	}

	p.Status = status
	if paidAt != nil {
		p.PaidAt = paidAt
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateExternalReference attaches provider external reference and checkout URL.
func (r *PaymentRepository) UpdateExternalReference(ctx context.Context, paymentID uuid.UUID, externalID, paymentURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.payments[paymentID]
	if !exists {
		return apperror.NotFound("payment")
	}

	p.ExternalID = externalID
	p.ExternalReference = externalID
	p.PaymentURL = paymentURL
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// List retrieves paginated payments for an organization.
func (r *PaymentRepository) List(ctx context.Context, orgID uuid.UUID, page, perPage int) ([]*payment.Payment, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*payment.Payment
	for _, p := range r.payments {
		if orgID == uuid.Nil || p.OrganizationID == orgID {
			cp := *p
			matched = append(matched, &cp)
		}
	}

	total := len(matched)
	start := (page - 1) * perPage
	if start >= total {
		return []*payment.Payment{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

// CreateRefund creates a refund record and updates the payment's refunded amount.
func (r *PaymentRepository) CreateRefund(ctx context.Context, cmd payment.CreateRefundCmd) (*payment.Refund, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.payments[cmd.PaymentID]
	if !exists {
		return nil, apperror.NotFound("payment")
	}

	if cmd.AmountCents <= 0 {
		return nil, apperror.ValidationFailed("refund amount must be greater than zero")
	}

	if cmd.AmountCents > p.NetAmount() {
		return nil, apperror.ValidationFailed("refund amount exceeds remaining net payment amount")
	}

	now := time.Now().UTC()
	refundID := uuid.New()

	ref := &payment.Refund{
		ID:          refundID,
		PaymentID:   cmd.PaymentID,
		AmountCents: cmd.AmountCents,
		Reason:      cmd.Reason,
		Status:      "completed",
		CreatedBy:   cmd.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	p.RefundedAmountCents += cmd.AmountCents
	if p.IsFullyRefunded() {
		p.Status = payment.StatusRefunded
	} else {
		p.Status = payment.StatusPartiallyRefunded
	}
	p.UpdatedAt = now

	r.refunds[refundID] = ref
	return ref, nil
}

// ListRefunds lists all refunds for a payment.
func (r *PaymentRepository) ListRefunds(ctx context.Context, paymentID uuid.UUID) ([]*payment.Refund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*payment.Refund
	for _, ref := range r.refunds {
		if ref.PaymentID == paymentID {
			cp := *ref
			result = append(result, &cp)
		}
	}
	return result, nil
}

// GetRefundByID retrieves a refund by ID.
func (r *PaymentRepository) GetRefundByID(ctx context.Context, refundID uuid.UUID) (*payment.Refund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ref, exists := r.refunds[refundID]
	if !exists {
		return nil, apperror.NotFound("refund")
	}
	cp := *ref
	return &cp, nil
}

// UpdateRefundStatus updates status of a refund.
func (r *PaymentRepository) UpdateRefundStatus(ctx context.Context, refundID uuid.UUID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref, exists := r.refunds[refundID]
	if !exists {
		return apperror.NotFound("refund")
	}

	ref.Status = status
	ref.UpdatedAt = time.Now().UTC()
	return nil
}

// SumByOrganization calculates total revenue for an organization in date range.
func (r *PaymentRepository) SumByOrganization(ctx context.Context, orgID uuid.UUID, from, to time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total int64
	for _, p := range r.payments {
		if p.OrganizationID == orgID && p.Status == payment.StatusPaid {
			if (from.IsZero() || !p.CreatedAt.Before(from)) && (to.IsZero() || !p.CreatedAt.After(to)) {
				total += p.NetAmount()
			}
		}
	}
	return total, nil
}

// IsWebhookProcessed checks if a webhook event ID / signature has been processed.
func (r *PaymentRepository) IsWebhookProcessed(eventKey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processedWebhooks[eventKey]
}

// MarkWebhookProcessed records that a webhook event has been processed.
func (r *PaymentRepository) MarkWebhookProcessed(eventKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processedWebhooks[eventKey] = true
}
