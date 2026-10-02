// Package payment defines the Payment and Refund domain.
// Payment lifecycle is separate from Appointment lifecycle.
// Payment provider implementations are abstracted behind the Provider interface.
package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle of a payment.
type Status string

const (
	StatusPending            Status = "pending"
	StatusPaid               Status = "paid"
	StatusFailed             Status = "failed"
	StatusRefunded           Status = "refunded"
	StatusPartiallyRefunded  Status = "partially_refunded"
	StatusExpired            Status = "expired"
	StatusCancelled          Status = "cancelled"
)

// Provider represents the payment processor.
type Provider string

const (
	ProviderStripe    Provider = "stripe"
	ProviderMidtrans  Provider = "midtrans"
	ProviderXendit    Provider = "xendit"
	ProviderManual    Provider = "manual" // cash/transfer recorded manually
)

// Payment represents a payment transaction for an appointment.
type Payment struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	AppointmentID  uuid.UUID  `json:"appointment_id"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty"`

	Status   Status   `json:"status"`
	Provider Provider `json:"provider"`

	// Amount in smallest currency unit
	AmountCents       int64  `json:"amount_cents"`
	Currency          string `json:"currency"`
	RefundedAmountCents int64 `json:"refunded_amount_cents"`

	// Provider-specific references
	ExternalID        string `json:"external_id,omitempty"`   // payment ID from provider
	ExternalReference string `json:"external_reference,omitempty"` // order/invoice ID
	PaymentURL        string `json:"payment_url,omitempty"`   // checkout URL for customer
	ReceiptURL        string `json:"receipt_url,omitempty"`

	// Metadata
	Notes     string     `json:"notes,omitempty"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NetAmount returns the amount after refunds.
func (p *Payment) NetAmount() int64 {
	return p.AmountCents - p.RefundedAmountCents
}

// IsFullyRefunded returns true if the payment has been fully refunded.
func (p *Payment) IsFullyRefunded() bool {
	return p.RefundedAmountCents >= p.AmountCents
}

// Refund represents a refund transaction against a payment.
type Refund struct {
	ID         uuid.UUID  `json:"id"`
	PaymentID  uuid.UUID  `json:"payment_id"`
	AmountCents int64     `json:"amount_cents"`
	Reason     string     `json:"reason,omitempty"`
	ExternalID string     `json:"external_id,omitempty"`
	Status     string     `json:"status"` // pending | completed | failed
	CreatedBy  uuid.UUID  `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// --- Commands ----------------------------------------------------------------

type CreatePaymentCmd struct {
	OrganizationID uuid.UUID
	AppointmentID  uuid.UUID
	CustomerID     *uuid.UUID
	Provider       Provider
	AmountCents    int64
	Currency       string
	ExpiresAt      *time.Time
}

type CreateRefundCmd struct {
	PaymentID   uuid.UUID
	AmountCents int64
	Reason      string
	CreatedBy   uuid.UUID
}

// --- Provider Abstraction Interface ------------------------------------------

// PaymentRequest is the input to create a payment intent/link.
type PaymentRequest struct {
	OrderID     string
	AmountCents int64
	Currency    string
	Description string
	CustomerEmail string
	CustomerName  string
	SuccessURL  string
	FailureURL  string
	Metadata    map[string]string
}

// PaymentResponse is the result from creating a payment.
type PaymentResponse struct {
	ExternalID  string
	PaymentURL  string
	ExpiresAt   *time.Time
}

// WebhookEvent represents an inbound payment event from a provider.
type WebhookEvent struct {
	Provider    Provider
	ExternalID  string
	Status      Status
	AmountCents int64
	RawPayload  []byte
}

// ProviderService is the abstraction over payment gateways.
// Implementations are in the infrastructure layer.
type ProviderService interface {
	// CreatePayment creates a payment intent/link and returns the checkout URL.
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)

	// GetPaymentStatus queries the current status from the provider.
	GetPaymentStatus(ctx context.Context, externalID string) (Status, error)

	// CreateRefund initiates a refund at the provider.
	CreateRefund(ctx context.Context, externalID string, amountCents int64, reason string) (string, error)

	// VerifyWebhookSignature verifies the webhook payload signature.
	VerifyWebhookSignature(payload []byte, signature string) error

	// ParseWebhookEvent parses the raw webhook payload into a WebhookEvent.
	ParseWebhookEvent(payload []byte) (*WebhookEvent, error)

	// Provider returns the provider identifier.
	Provider() Provider
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// Payments
	Create(ctx context.Context, cmd CreatePaymentCmd) (*Payment, error)
	GetByID(ctx context.Context, orgID, paymentID uuid.UUID) (*Payment, error)
	GetByAppointmentID(ctx context.Context, orgID, appointmentID uuid.UUID) (*Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*Payment, error)
	UpdateStatus(ctx context.Context, paymentID uuid.UUID, status Status, paidAt *time.Time) error
	UpdateExternalReference(ctx context.Context, paymentID uuid.UUID, externalID, paymentURL string) error
	List(ctx context.Context, orgID uuid.UUID, page, perPage int) ([]*Payment, int, error)

	// Refunds
	CreateRefund(ctx context.Context, cmd CreateRefundCmd) (*Refund, error)
	ListRefunds(ctx context.Context, paymentID uuid.UUID) ([]*Refund, error)
	GetRefundByID(ctx context.Context, refundID uuid.UUID) (*Refund, error)
	UpdateRefundStatus(ctx context.Context, refundID uuid.UUID, status string) error

	// Analytics
	SumByOrganization(ctx context.Context, orgID uuid.UUID, from, to time.Time) (int64, error)
}
