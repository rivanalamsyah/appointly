package integration

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domain "github.com/appointly/appointly/backend/internal/domain/integration"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/integration"
)

type Handler struct {
	svc *usecase.Service
}

func NewHandler(svc *usecase.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/integrations", func(r chi.Router) {
		r.Get("/", h.ListIntegrations)
		r.Post("/{provider}/connect", h.ConnectIntegration)
		r.Post("/{provider}/disconnect", h.DisconnectIntegration)
		r.Post("/{provider}/reconnect", h.ReconnectIntegration)
	})
}

func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	items, err := h.svc.ListIntegrations(r.Context(), orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"integrations": items,
	})
}

func (h *Handler) ConnectIntegration(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	providerStr := chi.URLParam(r, "provider")
	provider := domain.Provider(providerStr)

	var req struct {
		AccountEmail  string            `json:"account_email"`
		WebhookURL    string            `json:"webhook_url"`
		WebhookSecret string            `json:"webhook_secret"`
		Settings      map[string]string `json:"settings"`
	}

	_ = v1.DecodeJSON(r, &req)

	cmd := usecase.ConnectCmd{
		OrganizationID: orgID,
		Provider:       provider,
		AccountEmail:   req.AccountEmail,
		WebhookURL:     req.WebhookURL,
		WebhookSecret:  req.WebhookSecret,
		Settings:       req.Settings,
	}

	it, err := h.svc.ConnectIntegration(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"integration": it,
		"message":     "Integration connected successfully",
	})
}

func (h *Handler) DisconnectIntegration(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	providerStr := chi.URLParam(r, "provider")
	provider := domain.Provider(providerStr)

	err = h.svc.DisconnectIntegration(r.Context(), orgID, provider)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Integration disconnected successfully",
	})
}

func (h *Handler) ReconnectIntegration(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	providerStr := chi.URLParam(r, "provider")
	provider := domain.Provider(providerStr)

	it, err := h.svc.ReconnectIntegration(r.Context(), orgID, provider)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"integration": it,
		"message":     "Integration reconnected successfully",
	})
}
