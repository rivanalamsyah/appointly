package admin

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/admin"
)

type Handler struct {
	svc *usecase.Service
}

func NewHandler(svc *usecase.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.RequireSuperAdmin)

		r.Get("/health", h.GetHealthSummary)
		r.Get("/organizations", h.ListOrganizations)
		r.Post("/organizations/{id}/suspend", h.SuspendOrganization)
		r.Post("/organizations/{id}/activate", h.ActivateOrganization)
		r.Post("/organizations/{id}/assign-plan", h.AssignPlanTier)
		r.Post("/organizations/{id}/customers/{customer_id}/breakglass", h.BreakGlassCustomerAccess)
	})
}

func (h *Handler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetSystemHealthSummary(r.Context())
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}
	v1.RespondJSON(w, r, http.StatusOK, summary)
}

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	items, total, err := h.svc.ListPlatformOrganizations(r.Context(), page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"organizations": items,
		"total":         total,
	})
}

func (h *Handler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid organization ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = v1.DecodeJSON(r, &req)

	err = h.svc.SuspendOrganization(r.Context(), orgID, req.Reason)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Organization suspended successfully",
	})
}

func (h *Handler) ActivateOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid organization ID"))
		return
	}

	err = h.svc.ActivateOrganization(r.Context(), orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Organization activated successfully",
	})
}

func (h *Handler) AssignPlanTier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid organization ID"))
		return
	}

	var req struct {
		PlanSlug string `json:"plan_slug"`
	}
	if err := v1.DecodeJSON(r, &req); err != nil || req.PlanSlug == "" {
		v1.RespondError(w, r, apperror.BadRequest("plan_slug is required"))
		return
	}

	err = h.svc.AssignPlanTier(r.Context(), orgID, req.PlanSlug)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Organization SaaS plan updated successfully",
	})
}

func (h *Handler) BreakGlassCustomerAccess(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid organization ID"))
		return
	}

	custIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customer ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := v1.DecodeJSON(r, &req); err != nil || req.Reason == "" {
		v1.RespondError(w, r, apperror.BadRequest("explicit support reason is required"))
		return
	}

	cust, err := h.svc.BreakGlassCustomerAccess(r.Context(), orgID, customerID, req.Reason)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, cust)
}
