package payment_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/payment"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	infraPay "github.com/appointly/appointly/backend/internal/infrastructure/payment"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
	bookinguc "github.com/appointly/appointly/backend/internal/usecase/booking"
	paymentuc "github.com/appointly/appointly/backend/internal/usecase/payment"
)

func setupTestPaymentService(t *testing.T) (
	*paymentuc.Service,
	*bookinguc.Service,
	*infraPay.MockProvider,
	*organization.Organization,
	*appointment.Appointment,
	*memory.PaymentRepository,
) {
	orgRepo := memory.NewOrgRepository()
	locRepo := memory.NewLocationRepository()
	svcRepo := memory.NewServiceRepository()
	staffRepo := memory.NewStaffRepository()
	schedRepo := memory.NewSchedulingRepository()
	resRepo := memory.NewResourceRepository()
	apptRepo := memory.NewAppointmentRepository()
	payRepo := memory.NewPaymentRepository()

	mockProv := infraPay.NewMockProvider("test-secret-key")
	providers := map[payment.Provider]payment.ProviderService{
		payment.ProviderManual: mockProv,
		payment.ProviderStripe: mockProv,
	}

	availEngine := availabilityuc.NewEngine(orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, apptRepo)
	bookingSvc := bookinguc.NewService(apptRepo, orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, availEngine)
	paySvc := paymentuc.NewService(payRepo, apptRepo, orgRepo, providers)

	ctx := context.Background()

	org, err := orgRepo.Create(ctx, &organization.Organization{
		ID:       uuid.New(),
		Name:     "Payment Test Salon",
		Slug:     "payment-test-salon",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	appt, err := apptRepo.Create(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: org.ID,
		ServiceID:      uuid.New(),
		StaffID:        uuid.New(),
		StartTime:      time.Now().Add(24 * time.Hour),
		EndTime:        time.Now().Add(25 * time.Hour),
		PriceCents:     5000,
		Currency:       "USD",
		ServiceName:    "Full Facial",
		GuestName:      "Jane Customer",
		GuestEmail:     "jane@example.com",
		GuestPhone:     "+1555000111",
	})
	if err != nil {
		t.Fatalf("failed to create appointment: %v", err)
	}

	return paySvc, bookingSvc, mockProv, org, appt, payRepo
}

func TestPayment_IntentCreation(t *testing.T) {
	paySvc, _, _, org, appt, _ := setupTestPaymentService(t)

	intent, err := paySvc.CreatePaymentIntent(context.Background(), paymentuc.CreateIntentCmd{
		OrganizationID: org.ID,
		AppointmentID:  appt.ID,
		Provider:       payment.ProviderManual,
	})

	if err != nil {
		t.Fatalf("failed to create payment intent: %v", err)
	}

	if intent.Status != payment.StatusPending {
		t.Errorf("expected initial payment status pending, got %s", intent.Status)
	}
	if intent.AmountCents != 5000 {
		t.Errorf("expected amount 5000 cents, got %d", intent.AmountCents)
	}
	if intent.PaymentURL == "" {
		t.Error("expected valid checkout payment URL from provider, got empty string")
	}
}

func TestPayment_InvalidSignatureRejection(t *testing.T) {
	paySvc, _, _, _, _, _ := setupTestPaymentService(t)

	payload := []byte(`{"external_id":"mock_123","status":"paid"}`)
	err := paySvc.HandleWebhook(context.Background(), payment.ProviderManual, payload, "invalid-bad-signature")

	if err == nil {
		t.Error("SECURITY FAILURE: Expected invalid signature error, got nil")
	}
}

func TestPayment_VerifiedSuccessWebhook_AutoConfirmsAppointment(t *testing.T) {
	paySvc, _, mockProv, org, appt, _ := setupTestPaymentService(t)

	// 1. Create Payment Intent
	intent, err := paySvc.CreatePaymentIntent(context.Background(), paymentuc.CreateIntentCmd{
		OrganizationID: org.ID,
		AppointmentID:  appt.ID,
		Provider:       payment.ProviderManual,
	})
	if err != nil {
		t.Fatalf("failed to create intent: %v", err)
	}

	// 2. Construct Webhook Payload with verified HMAC signature
	webhookPayload, _ := json.Marshal(map[string]interface{}{
		"event_id":     "evt_1001",
		"external_id":  intent.ExternalID,
		"status":       "paid",
		"amount_cents": 5000,
		"provider":     "manual",
	})
	sig := mockProv.GenerateMockSignature(webhookPayload)

	// 3. Handle Webhook
	err = paySvc.HandleWebhook(context.Background(), payment.ProviderManual, webhookPayload, sig)
	if err != nil {
		t.Fatalf("webhook handling failed: %v", err)
	}

	// 4. Verify Payment Status updated to paid
	updatedPay, err := paySvc.GetPaymentByID(context.Background(), org.ID, intent.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated payment: %v", err)
	}
	if updatedPay.Status != payment.StatusPaid {
		t.Errorf("expected payment status paid, got %s", updatedPay.Status)
	}
}

func TestPayment_DuplicateWebhookIdempotency(t *testing.T) {
	paySvc, _, mockProv, org, appt, _ := setupTestPaymentService(t)

	intent, _ := paySvc.CreatePaymentIntent(context.Background(), paymentuc.CreateIntentCmd{
		OrganizationID: org.ID,
		AppointmentID:  appt.ID,
		Provider:       payment.ProviderManual,
	})

	webhookPayload, _ := json.Marshal(map[string]interface{}{
		"event_id":     "evt_dup_999",
		"external_id":  intent.ExternalID,
		"status":       "paid",
		"amount_cents": 5000,
		"provider":     "manual",
	})
	sig := mockProv.GenerateMockSignature(webhookPayload)

	// First execution -> succeeds
	err1 := paySvc.HandleWebhook(context.Background(), payment.ProviderManual, webhookPayload, sig)
	if err1 != nil {
		t.Fatalf("first webhook handling failed: %v", err1)
	}

	// Duplicate execution -> idempotency check returns nil cleanly without duplicate error
	err2 := paySvc.HandleWebhook(context.Background(), payment.ProviderManual, webhookPayload, sig)
	if err2 != nil {
		t.Fatalf("duplicate webhook should be handled idempotently, got %v", err2)
	}
}

func TestPayment_RefundAndPartialRefundWorkflow(t *testing.T) {
	paySvc, _, mockProv, org, appt, _ := setupTestPaymentService(t)

	ctxOwner := rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   org.ID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	})

	// 1. Create and pay intent
	intent, _ := paySvc.CreatePaymentIntent(ctxOwner, paymentuc.CreateIntentCmd{
		OrganizationID: org.ID,
		AppointmentID:  appt.ID,
		Provider:       payment.ProviderManual,
	})

	webhookPayload, _ := json.Marshal(map[string]interface{}{
		"external_id":  intent.ExternalID,
		"status":       "paid",
		"amount_cents": 5000,
		"provider":     "manual",
	})
	sig := mockProv.GenerateMockSignature(webhookPayload)
	_ = paySvc.HandleWebhook(ctxOwner, payment.ProviderManual, webhookPayload, sig)

	// 2. Partial Refund ($20.00 = 2000 cents)
	ref1, err := paySvc.ProcessRefund(ctxOwner, payment.CreateRefundCmd{
		PaymentID:   intent.ID,
		AmountCents: 2000,
		Reason:      "Customer requested partial refund for delay",
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("failed partial refund: %v", err)
	}
	if ref1.AmountCents != 2000 {
		t.Errorf("expected refund amount 2000, got %d", ref1.AmountCents)
	}

	updatedPay1, _ := paySvc.GetPaymentByID(ctxOwner, org.ID, intent.ID)
	if updatedPay1.Status != payment.StatusPartiallyRefunded {
		t.Errorf("expected status partially_refunded, got %s", updatedPay1.Status)
	}
	if updatedPay1.NetAmount() != 3000 {
		t.Errorf("expected net amount 3000 cents remaining, got %d", updatedPay1.NetAmount())
	}

	// 3. Excess Refund Attempt ($40.00 > remaining $30.00) -> Should fail validation
	_, errExcess := paySvc.ProcessRefund(ctxOwner, payment.CreateRefundCmd{
		PaymentID:   intent.ID,
		AmountCents: 4000,
		Reason:      "Excess refund test",
	})
	if errExcess == nil {
		t.Error("expected validation error when refund exceeds net amount, got nil")
	}

	// 4. Full remaining refund ($30.00 = 3000 cents) -> Transitions to fully refunded
	_, errFull := paySvc.ProcessRefund(ctxOwner, payment.CreateRefundCmd{
		PaymentID:   intent.ID,
		AmountCents: 3000,
		Reason:      "Remaining full refund",
	})
	if errFull != nil {
		t.Fatalf("failed full remaining refund: %v", errFull)
	}

	updatedPay2, _ := paySvc.GetPaymentByID(ctxOwner, org.ID, intent.ID)
	if updatedPay2.Status != payment.StatusRefunded {
		t.Errorf("expected status refunded, got %s", updatedPay2.Status)
	}
}
