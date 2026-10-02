package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/integration"
	infra "github.com/appointly/appointly/backend/internal/infrastructure/integration"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/integration"
)

func setupTestIntegrationService() (*usecase.Service, *memory.IntegrationRepository, *memory.AuditRepository) {
	repo := memory.NewIntegrationRepository()
	auditRepo := memory.NewAuditRepository()
	gcal := infra.NewGoogleCalendarAdapter()
	wa := infra.NewWhatsAppAdapter()
	email := infra.NewEmailAdapter()
	wh := infra.NewWebhookAdapter()

	svc := usecase.NewService(repo, auditRepo, gcal, wa, email, wh)
	return svc, repo, auditRepo
}

func TestIntegration_ConnectDisconnectReconnect(t *testing.T) {
	svc, repo, auditRepo := setupTestIntegrationService()
	ctx := context.Background()

	orgID := uuid.New()

	// 1. Connect Google Calendar
	cmd := usecase.ConnectCmd{
		OrganizationID: orgID,
		Provider:       integration.ProviderGoogleCalendar,
		AccountEmail:   "owner@orga.com",
	}

	it, err := svc.ConnectIntegration(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error connecting integration, got %v", err)
	}
	if it.Status != integration.StatusConnected {
		t.Fatalf("expected status connected, got %s", it.Status)
	}

	// Verify Audit Log entry
	logs, total, _ := auditRepo.List(ctx, orgID, audit.ListFilter{}, 1, 10)
	if total < 1 || logs[0].Action != "integration.connected" {
		t.Fatalf("expected audit log event integration.connected")
	}

	// 2. Disconnect Integration
	err = svc.DisconnectIntegration(ctx, orgID, integration.ProviderGoogleCalendar)
	if err != nil {
		t.Fatalf("expected no error disconnecting integration, got %v", err)
	}

	disconnected, _ := repo.GetByProvider(ctx, orgID, integration.ProviderGoogleCalendar)
	if disconnected.Status != integration.StatusDisconnected {
		t.Fatalf("expected status disconnected, got %s", disconnected.Status)
	}

	// 3. Reconnect Integration
	reconnected, err := svc.ReconnectIntegration(ctx, orgID, integration.ProviderGoogleCalendar)
	if err != nil {
		t.Fatalf("expected no error reconnecting, got %v", err)
	}
	if reconnected.Status != integration.StatusConnected {
		t.Fatalf("expected status connected after reconnect, got %s", reconnected.Status)
	}
}

func TestIntegration_TokenExpirationAndAutoRefresh(t *testing.T) {
	svc, repo, _ := setupTestIntegrationService()
	ctx := context.Background()

	orgID := uuid.New()

	// Connect Google Calendar
	it, _ := svc.ConnectIntegration(ctx, usecase.ConnectCmd{
		OrganizationID: orgID,
		Provider:       integration.ProviderGoogleCalendar,
		AccountEmail:   "user@orga.com",
	})

	// Manually expire access token
	past := time.Now().Add(-1 * time.Hour)
	it.TokenExpiresAt = &past
	_, _ = repo.Update(ctx, it)

	// Attempt calendar appointment sync
	appt := &appointment.Appointment{
		ID:             uuid.New(),
		OrganizationID: orgID,
		StartTime:      time.Now().Add(24 * time.Hour),
		EndTime:        time.Now().Add(25 * time.Hour),
	}

	syncState, err := svc.SyncAppointmentCalendar(ctx, orgID, appt)
	if err != nil {
		t.Fatalf("expected auto-token refresh during sync, got error %v", err)
	}

	if syncState.ExternalEventID == "" {
		t.Fatalf("expected valid external event ID after sync")
	}

	// Check that access token was updated in repository
	updatedIt, _ := repo.GetByProvider(ctx, orgID, integration.ProviderGoogleCalendar)
	if updatedIt.TokenExpiresAt.Before(time.Now()) {
		t.Fatalf("expected updated token expiration date in future")
	}
}

func TestIntegration_TenantIsolation(t *testing.T) {
	svc, _, _ := setupTestIntegrationService()
	ctx := context.Background()

	orgA := uuid.New()
	orgB := uuid.New()

	// Org A connects WhatsApp
	_, _ = svc.ConnectIntegration(ctx, usecase.ConnectCmd{
		OrganizationID: orgA,
		Provider:       integration.ProviderWhatsAppTwilio,
		AccountEmail:   "wha@orga.com",
	})

	// Org B attempts to disconnect Org A's integration
	err := svc.DisconnectIntegration(ctx, orgB, integration.ProviderWhatsAppTwilio)
	if err == nil {
		t.Fatalf("expected tenant isolation error disconnecting Org A's integration from Org B context")
	}

	// Org B lists integrations
	listB, _ := svc.ListIntegrations(ctx, orgB)
	for _, item := range listB {
		if item.Provider == integration.ProviderWhatsAppTwilio && item.Status == integration.StatusConnected {
			t.Fatalf("Org B should not see Org A's active WhatsApp integration")
		}
	}
}
