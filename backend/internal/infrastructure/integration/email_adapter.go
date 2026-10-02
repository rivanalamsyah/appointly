package integration

import (
	"context"
	"fmt"
	"time"

	domain "github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type EmailAdapter struct{}

func NewEmailAdapter() *EmailAdapter {
	return &EmailAdapter{}
}

func (a *EmailAdapter) SendEmail(ctx context.Context, item *domain.Integration, msg domain.EmailMessage) (string, error) {
	if item.Status != domain.StatusConnected {
		return "", apperror.Forbidden("Email integration is disconnected")
	}

	if msg.ToEmail == "" {
		return "", apperror.BadRequest("recipient email is required")
	}

	emailID := fmt.Sprintf("email_resend_%d", time.Now().UnixNano())
	return emailID, nil
}
