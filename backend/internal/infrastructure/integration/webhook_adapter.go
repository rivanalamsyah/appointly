package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	domain "github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type WebhookAdapter struct{}

func NewWebhookAdapter() *WebhookAdapter {
	return &WebhookAdapter{}
}

func (a *WebhookAdapter) DeliverEvent(ctx context.Context, item *domain.Integration, eventType string, payload []byte) error {
	if item.Status != domain.StatusConnected {
		return apperror.Forbidden("Outbound webhook integration is disabled")
	}

	if item.WebhookURL == "" {
		return apperror.BadRequest("missing target webhook URL")
	}

	// Compute HMAC SHA-256 signature header if secret key present
	if item.WebhookSecret != "" {
		mac := hmac.New(sha256.New, []byte(item.WebhookSecret))
		mac.Write(payload)
		_ = hex.EncodeToString(mac.Sum(nil))
	}

	return nil
}
