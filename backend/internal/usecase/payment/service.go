package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/payment"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// Service provides use cases for the Payment domain.
type Service struct {
	paymentRepo appointmentPaymentRepo
	apptRepo    appointment.Repository
	orgRepo     organization.Repository
	providers   map[payment.Provider]payment.ProviderService
}

// Interface wrapper extending payment.Repository to include webhook idempotency methods
type appointmentPaymentRepo interface {
	payment.Repository
	IsWebhookProcessed(eventKey string) bool
	MarkWebhookProcessed(eventKey string)
}

// NewService constructs a new Payment Service.
func NewService(
	paymentRepo appointmentPaymentRepo,
	apptRepo appointment.Repository,
	orgRepo organization.Repository,
	providers map[payment.Provider]payment.ProviderService,
) *Service {
	return &Service{
		paymentRepo: paymentRepo,
		apptRepo:    apptRepo,
		orgRepo:     orgRepo,
		providers:   providers,
	}
}

// CreateIntentCmd contains data for initiating a payment checkout intent.
type CreateIntentCmd struct {
	OrganizationID uuid.UUID
	AppointmentID  uuid.UUID
	Provider       payment.Provider
	SuccessURL     string
	CancelURL      string
}

// CreatePaymentIntent initiates a checkout order with the selected gateway.
func (s *Service) CreatePaymentIntent(ctx context.Context, cmd CreateIntentCmd) (*payment.Payment, error) {
	// 1. Fetch organization and appointment
	org, err := s.orgRepo.GetByID(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, apperror.NotFound("organization")
	}

	appt, err := s.apptRepo.GetByID(ctx, cmd.OrganizationID, cmd.AppointmentID)
	if err != nil {
		return nil, apperror.NotFound("appointment")
	}

	// Check existing payment for this appointment
	existing, _ := s.paymentRepo.GetByAppointmentID(ctx, cmd.OrganizationID, cmd.AppointmentID)
	if existing != nil && existing.Status == payment.StatusPaid {
		return existing, nil // Already paid
	}

	// 2. Select Provider
	prov, ok := s.providers[cmd.Provider]
	if !ok {
		// Fallback to first available provider
		for _, p := range s.providers {
			prov = p
			break
		}
	}
	if prov == nil {
		return nil, apperror.ValidationFailed("no payment provider configured")
	}

	// 3. Determine amount based on price snapshot
	amountCents := appt.PriceCents
	if amountCents <= 0 {
		amountCents = 1000 // default minimum test amount
	}

	// 4. Create local pending payment record
	createCmd := payment.CreatePaymentCmd{
		OrganizationID: cmd.OrganizationID,
		AppointmentID:  cmd.AppointmentID,
		CustomerID:     appt.CustomerID,
		Provider:       prov.Provider(),
		AmountCents:    amountCents,
		Currency:       org.Currency,
	}

	payRecord, err := s.paymentRepo.Create(ctx, createCmd)
	if err != nil {
		return nil, err
	}

	// 5. Invoke provider to create intent / order URL
	req := payment.PaymentRequest{
		OrderID:       payRecord.ID.String(),
		AmountCents:   amountCents,
		Currency:      org.Currency,
		Description:   fmt.Sprintf("Booking for %s", org.Name),
		CustomerEmail: appt.GuestEmail,
		CustomerName:  appt.GuestName,
		SuccessURL:    cmd.SuccessURL,
		FailureURL:    cmd.CancelURL,
	}

	resp, err := prov.CreatePayment(ctx, req)
	if err != nil {
		return nil, apperror.Internal("failed to create payment intent with gateway")
	}

	// Update local record with provider external references
	_ = s.paymentRepo.UpdateExternalReference(ctx, payRecord.ID, resp.ExternalID, resp.PaymentURL)
	payRecord.ExternalID = resp.ExternalID
	payRecord.ExternalReference = resp.ExternalID
	payRecord.PaymentURL = resp.PaymentURL

	return payRecord, nil
}

// HandleWebhook processes inbound webhook events from payment gateways.
// SECURITY RULE: Never trust payment status from frontend client; final payment status MUST come from verified webhook!
func (s *Service) HandleWebhook(ctx context.Context, providerName payment.Provider, payload []byte, signature string) error {
	prov, ok := s.providers[providerName]
	if !ok {
		return apperror.BadRequest("unsupported payment provider webhook")
	}

	// 1. Verify Signature
	if err := prov.VerifyWebhookSignature(payload, signature); err != nil {
		return apperror.Unauthorized("invalid webhook signature")
	}

	// 2. Parse Webhook Event
	event, err := prov.ParseWebhookEvent(payload)
	if err != nil {
		return err
	}

	// 3. Idempotency Check — calculate event key
	hasher := sha256.New()
	hasher.Write(payload)
	eventKey := hex.EncodeToString(hasher.Sum(nil))

	if s.paymentRepo.IsWebhookProcessed(eventKey) {
		return nil // Duplicate webhook event gracefully ignored
	}

	// 4. Find linked payment record by ExternalID or Order ID
	payRecord, err := s.paymentRepo.GetByExternalID(ctx, event.ExternalID)
	if err != nil {
		// Attempt fallback by UUID parsed from external ID
		if parsedID, pErr := uuid.Parse(event.ExternalID); pErr == nil {
			payRecord, err = s.paymentRepo.GetByID(ctx, uuid.Nil, parsedID)
		}
	}
	if err != nil {
		return apperror.NotFound("payment record for webhook")
	}

	// 5. Process Payment Status & Keep Appointment State Consistent
	if event.Status == payment.StatusPaid {
		_ = s.paymentRepo.UpdateStatus(ctx, payRecord.ID, payment.StatusPaid, nil)

		// State Consistency: Automatically confirm appointment on verified payment success!
		_ = s.apptRepo.UpdateStatus(ctx, appointment.UpdateStatusCmd{
			AppointmentID: payRecord.AppointmentID,
			NewStatus:     appointment.StatusConfirmed,
			Reason:        "Payment successfully verified via backend webhook",
		})
	} else if event.Status == payment.StatusFailed {
		_ = s.paymentRepo.UpdateStatus(ctx, payRecord.ID, payment.StatusFailed, nil)
	}

	// Mark webhook key as processed for idempotency
	s.paymentRepo.MarkWebhookProcessed(eventKey)
	return nil
}

// ProcessRefund handles refund requests with permission checks.
func (s *Service) ProcessRefund(ctx context.Context, cmd payment.CreateRefundCmd) (*payment.Refund, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil && !authCtx.Can(rbac.PermPaymentRefund) {
		return nil, apperror.Unauthorized("insufficient permissions to process refund")
	}

	payRecord, err := s.paymentRepo.GetByID(ctx, uuid.Nil, cmd.PaymentID)
	if err != nil {
		return nil, apperror.NotFound("payment")
	}

	if cmd.AmountCents <= 0 {
		return nil, apperror.ValidationFailed("refund amount must be greater than zero")
	}

	if cmd.AmountCents > payRecord.NetAmount() {
		return nil, apperror.ValidationFailed("refund amount exceeds remaining net payment amount")
	}

	// Invoke provider refund if external payment exists
	if payRecord.ExternalID != "" {
		prov, ok := s.providers[payRecord.Provider]
		if ok && prov != nil {
			_, _ = prov.CreateRefund(ctx, payRecord.ExternalID, cmd.AmountCents, cmd.Reason)
		}
	}

	ref, err := s.paymentRepo.CreateRefund(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

// GetPaymentByID retrieves a payment by ID with authorization check.
func (s *Service) GetPaymentByID(ctx context.Context, orgID, paymentID uuid.UUID) (*payment.Payment, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil {
		orgID = authCtx.OrgID
		if !authCtx.Can(rbac.PermPaymentRead) {
			return nil, apperror.Unauthorized("insufficient permissions to read payment")
		}
	}

	return s.paymentRepo.GetByID(ctx, orgID, paymentID)
}

// ListPayments returns paginated payments for an organization.
func (s *Service) ListPayments(ctx context.Context, orgID uuid.UUID, page, perPage int) ([]*payment.Payment, int, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil {
		orgID = authCtx.OrgID
		if !authCtx.Can(rbac.PermPaymentRead) {
			return nil, 0, apperror.Unauthorized("insufficient permissions to read payments")
		}
	}

	return s.paymentRepo.List(ctx, orgID, page, perPage)
}
