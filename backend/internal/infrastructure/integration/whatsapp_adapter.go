package integration

import (
	"context"
	"fmt"
	"time"

	domain "github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type WhatsAppAdapter struct{}

func NewWhatsAppAdapter() *WhatsAppAdapter {
	return &WhatsAppAdapter{}
}

func (a *WhatsAppAdapter) SendMessage(ctx context.Context, item *domain.Integration, msg domain.WhatsAppMessage) (string, error) {
	if item.Status != domain.StatusConnected {
		return "", apperror.Forbidden("WhatsApp integration is disconnected")
	}

	if msg.RecipientPhone == "" {
		return "", apperror.BadRequest("recipient phone number is required")
	}

	msgID := fmt.Sprintf("wa_msg_%d", time.Now().UnixNano())
	return msgID, nil
}
