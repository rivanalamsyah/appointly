package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/payment"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// MockProvider implements payment.ProviderService for testing and demonstration.
type MockProvider struct {
	secretKey string
}

// NewMockProvider constructs a MockProvider with a secret key.
func NewMockProvider(secretKey string) *MockProvider {
	if secretKey == "" {
		secretKey = "mock-secret-key-appointly"
	}
	return &MockProvider{secretKey: secretKey}
}

func (m *MockProvider) Provider() payment.Provider {
	return payment.ProviderManual
}

func (m *MockProvider) CreatePayment(ctx context.Context, req payment.PaymentRequest) (*payment.PaymentResponse, error) {
	extID := fmt.Sprintf("mock_pay_%s", uuid.New().String()[:8])
	expiresAt := time.Now().UTC().Add(30 * time.Minute)

	return &payment.PaymentResponse{
		ExternalID: extID,
		PaymentURL: fmt.Sprintf("https://checkout.appointly.dev/pay/%s", extID),
		ExpiresAt:  &expiresAt,
	}, nil
}

func (m *MockProvider) GetPaymentStatus(ctx context.Context, externalID string) (payment.Status, error) {
	return payment.StatusPaid, nil
}

func (m *MockProvider) CreateRefund(ctx context.Context, externalID string, amountCents int64, reason string) (string, error) {
	refundExtID := fmt.Sprintf("mock_ref_%s", uuid.New().String()[:8])
	return refundExtID, nil
}

func (m *MockProvider) VerifyWebhookSignature(payload []byte, signature string) error {
	if signature == "" {
		return apperror.Unauthorized("missing webhook signature header")
	}

	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) && signature != "valid-mock-signature" {
		return apperror.Unauthorized("invalid webhook signature")
	}

	return nil
}

func (m *MockProvider) ParseWebhookEvent(payload []byte) (*payment.WebhookEvent, error) {
	var raw struct {
		EventID     string         `json:"event_id"`
		ExternalID  string         `json:"external_id"`
		Status      payment.Status `json:"status"`
		AmountCents int64          `json:"amount_cents"`
		Provider    string         `json:"provider"`
	}

	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, apperror.BadRequest("invalid webhook JSON payload")
	}

	if raw.ExternalID == "" {
		return nil, apperror.ValidationFailed("webhook payload missing external_id")
	}

	return &payment.WebhookEvent{
		Provider:    payment.Provider(raw.Provider),
		ExternalID:  raw.ExternalID,
		Status:      raw.Status,
		AmountCents: raw.AmountCents,
		RawPayload:  payload,
	}, nil
}

// GenerateMockSignature creates a valid HMAC signature for mock testing.
func (m *MockProvider) GenerateMockSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
