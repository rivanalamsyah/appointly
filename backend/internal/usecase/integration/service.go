package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/integration"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	repo            integration.Repository
	auditRepo       audit.Repository
	gcalAdapter     integration.CalendarProvider
	whatsappAdapter integration.WhatsAppProvider
	emailAdapter    integration.EmailProvider
	webhookAdapter  integration.OutboundWebhookService
}

func NewService(
	repo integration.Repository,
	auditRepo audit.Repository,
	gcal integration.CalendarProvider,
	wa integration.WhatsAppProvider,
	email integration.EmailProvider,
	wh integration.OutboundWebhookService,
) *Service {
	return &Service{
		repo:            repo,
		auditRepo:       auditRepo,
		gcalAdapter:     gcal,
		whatsappAdapter: wa,
		emailAdapter:    email,
		webhookAdapter:  wh,
	}
}

// ListIntegrations retrieves all active integrations and connection states for an organization.
func (s *Service) ListIntegrations(ctx context.Context, orgID uuid.UUID) ([]*integration.Integration, error) {
	items, err := s.repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Ensure default providers are present in returned list for UI initialization
	providers := []integration.Provider{
		integration.ProviderGoogleCalendar,
		integration.ProviderWhatsAppTwilio,
		integration.ProviderEmailResend,
		integration.ProviderOutboundWebhook,
	}

	foundMap := make(map[integration.Provider]*integration.Integration)
	for _, it := range items {
		foundMap[it.Provider] = it
	}

	var result []*integration.Integration
	for _, p := range providers {
		if existing, ok := foundMap[p]; ok {
			result = append(result, existing)
		} else {
			// Stub disconnected integration
			cat := integration.CategoryCalendar
			if p == integration.ProviderWhatsAppTwilio {
				cat = integration.CategoryMessaging
			} else if p == integration.ProviderEmailResend {
				cat = integration.CategoryEmail
			} else if p == integration.ProviderOutboundWebhook {
				cat = integration.CategoryWebhook
			}

			result = append(result, &integration.Integration{
				ID:             uuid.Nil,
				OrganizationID: orgID,
				Provider:       p,
				Category:       cat,
				Status:         integration.StatusDisconnected,
			})
		}
	}

	return result, nil
}

type ConnectCmd struct {
	OrganizationID uuid.UUID
	Provider       integration.Provider
	AuthCode       string
	AccountEmail   string
	WebhookURL     string
	WebhookSecret  string
	Settings       map[string]string
}

// ConnectIntegration sets up or connects an external provider.
func (s *Service) ConnectIntegration(ctx context.Context, cmd ConnectCmd) (*integration.Integration, error) {
	if cmd.OrganizationID == uuid.Nil {
		return nil, apperror.BadRequest("organization ID is required")
	}

	existing, err := s.repo.GetByProvider(ctx, cmd.OrganizationID, cmd.Provider)
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days OAuth access token

	cat := integration.CategoryCalendar
	if cmd.Provider == integration.ProviderWhatsAppTwilio {
		cat = integration.CategoryMessaging
	} else if cmd.Provider == integration.ProviderEmailResend {
		cat = integration.CategoryEmail
	} else if cmd.Provider == integration.ProviderOutboundWebhook {
		cat = integration.CategoryWebhook
	}

	item := &integration.Integration{
		OrganizationID: cmd.OrganizationID,
		Provider:       cmd.Provider,
		Category:       cat,
		Status:         integration.StatusConnected,
		AccountEmail:   cmd.AccountEmail,
		AccessToken:    fmt.Sprintf("access_token_%s_%d", cmd.Provider, time.Now().Unix()),
		RefreshToken:   fmt.Sprintf("refresh_token_%s_%d", cmd.Provider, time.Now().Unix()),
		TokenExpiresAt: &expiresAt,
		WebhookURL:     cmd.WebhookURL,
		WebhookSecret:  cmd.WebhookSecret,
		Settings:       cmd.Settings,
		ErrorMessage:   "",
	}

	var saved *integration.Integration
	if err == nil && existing != nil {
		item.ID = existing.ID
		saved, err = s.repo.Update(ctx, item)
	} else {
		saved, err = s.repo.Create(ctx, item)
	}

	if err == nil && s.auditRepo != nil {
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			ID:             uuid.New(),
			OrganizationID: &cmd.OrganizationID,
			Action:         "integration.connected",
			ResourceType:   "integration",
			Metadata: map[string]interface{}{
				"provider": string(cmd.Provider),
			},
		})
	}

	return saved, err
}

// DisconnectIntegration disconnects an active provider.
func (s *Service) DisconnectIntegration(ctx context.Context, orgID uuid.UUID, provider integration.Provider) error {
	existing, err := s.repo.GetByProvider(ctx, orgID, provider)
	if err != nil {
		return apperror.NotFound("integration provider to disconnect")
	}

	existing.Status = integration.StatusDisconnected
	existing.AccessToken = ""
	existing.RefreshToken = ""
	_, err = s.repo.Update(ctx, existing)

	if err == nil && s.auditRepo != nil {
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			ID:             uuid.New(),
			OrganizationID: &orgID,
			Action:         "integration.disconnected",
			ResourceType:   "integration",
			Metadata: map[string]interface{}{
				"provider": string(provider),
			},
		})
	}

	return err
}

// ReconnectIntegration refreshes tokens and restores connected status.
func (s *Service) ReconnectIntegration(ctx context.Context, orgID uuid.UUID, provider integration.Provider) (*integration.Integration, error) {
	existing, err := s.repo.GetByProvider(ctx, orgID, provider)
	if err != nil {
		return nil, apperror.NotFound("integration to reconnect")
	}

	if existing.RefreshToken == "" {
		existing.RefreshToken = fmt.Sprintf("refresh_token_%s_%d", provider, time.Now().Unix())
	}

	if s.gcalAdapter != nil && provider == integration.ProviderGoogleCalendar {
		reconnected, err := s.gcalAdapter.RefreshAccessToken(ctx, existing)
		if err != nil {
			existing.Status = integration.StatusError
			existing.ErrorMessage = err.Error()
			_, _ = s.repo.Update(ctx, existing)

			if s.auditRepo != nil {
				_ = s.auditRepo.Create(ctx, audit.AuditLog{
					ID:             uuid.New(),
					OrganizationID: &orgID,
					Action:         "integration.sync_failed",
					ResourceType:   "integration",
				})
			}
			return nil, err
		}
		existing = reconnected
	} else {
		newExpires := time.Now().Add(30 * 24 * time.Hour)
		existing.Status = integration.StatusConnected
		existing.AccessToken = "reconnected_token_" + fmt.Sprintf("%d", time.Now().Unix())
		existing.TokenExpiresAt = &newExpires
		existing.ErrorMessage = ""
	}

	saved, err := s.repo.Update(ctx, existing)
	if err == nil && s.auditRepo != nil {
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			ID:             uuid.New(),
			OrganizationID: &orgID,
			Action:         "integration.reconnected",
			ResourceType:   "integration",
			Metadata: map[string]interface{}{
				"provider": string(provider),
			},
		})
	}

	return saved, err
}

// SyncAppointmentCalendar synchronizes a single appointment to Google Calendar.
func (s *Service) SyncAppointmentCalendar(ctx context.Context, orgID uuid.UUID, appt *appointment.Appointment) (*integration.SyncState, error) {
	it, err := s.repo.GetByProvider(ctx, orgID, integration.ProviderGoogleCalendar)
	if err != nil || it.Status != integration.StatusConnected {
		return nil, apperror.Forbidden("Google Calendar integration is not connected")
	}

	// Check token expiration
	if it.TokenExpiresAt != nil && time.Now().After(*it.TokenExpiresAt) {
		refreshed, refErr := s.ReconnectIntegration(ctx, orgID, integration.ProviderGoogleCalendar)
		if refErr != nil {
			return nil, apperror.Unauthorized("Google Calendar token expired and refresh failed")
		}
		it = refreshed
	}

	externalID, err := s.gcalAdapter.SyncAppointment(ctx, it, appt)
	if err != nil {
		return nil, err
	}

	syncState := &integration.SyncState{
		OrganizationID:  orgID,
		IntegrationID:   it.ID,
		AppointmentID:   appt.ID,
		ExternalEventID: externalID,
		SyncStatus:      "synced",
		LastSyncedAt:    time.Now(),
	}

	_ = s.repo.SaveSyncState(ctx, syncState)
	now := time.Now()
	it.LastSyncedAt = &now
	_, _ = s.repo.Update(ctx, it)

	return syncState, nil
}
