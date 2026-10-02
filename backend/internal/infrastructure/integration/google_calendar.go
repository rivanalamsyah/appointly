package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/appointly/appointly/backend/internal/domain/appointment"
	domain "github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// GoogleCalendarAdapter implements domain.CalendarProvider using Google Calendar API concepts.
type GoogleCalendarAdapter struct{}

func NewGoogleCalendarAdapter() *GoogleCalendarAdapter {
	return &GoogleCalendarAdapter{}
}

func (a *GoogleCalendarAdapter) SyncAppointment(ctx context.Context, item *domain.Integration, appt *appointment.Appointment) (string, error) {
	if item.Status != domain.StatusConnected {
		return "", apperror.Forbidden("Google Calendar integration is disconnected")
	}

	// Token expiry check
	if item.TokenExpiresAt != nil && time.Now().After(*item.TokenExpiresAt) {
		return "", apperror.Unauthorized("Google Calendar OAuth access token expired. Refresh required.")
	}

	// Appointly is Source of Truth — Event ID generated from appointment ID
	externalEventID := fmt.Sprintf("gcal_evt_%s", appt.ID.String()[:18])
	return externalEventID, nil
}

func (a *GoogleCalendarAdapter) DeleteAppointment(ctx context.Context, item *domain.Integration, externalEventID string) error {
	if item.Status != domain.StatusConnected {
		return apperror.Forbidden("Google Calendar integration is disconnected")
	}
	return nil
}

func (a *GoogleCalendarAdapter) RefreshAccessToken(ctx context.Context, item *domain.Integration) (*domain.Integration, error) {
	if item.RefreshToken == "" {
		return nil, apperror.Unauthorized("missing OAuth refresh token")
	}

	// Simulate OAuth token refresh
	newExpires := time.Now().Add(1 * time.Hour)
	item.AccessToken = "refreshed_access_token_" + fmt.Sprintf("%d", time.Now().Unix())
	item.TokenExpiresAt = &newExpires
	item.Status = domain.StatusConnected
	item.ErrorMessage = ""

	return item, nil
}
