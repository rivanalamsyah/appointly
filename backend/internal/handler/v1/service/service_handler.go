package service

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/service"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/service"
)

type ServiceHandler struct {
	serviceSvc *usecase.Service
}

func NewServiceHandler(serviceSvc *usecase.Service) *ServiceHandler {
	return &ServiceHandler{
		serviceSvc: serviceSvc,
	}
}

type CreateServiceRequest struct {
	CategoryID      *string `json:"category_id,omitempty"`
	Name            string  `json:"name"`
	Description     string  `json:"description,omitempty"`
	DurationMinutes int     `json:"duration_minutes"`
	BufferBefore    int     `json:"buffer_before_minutes,omitempty"`
	BufferAfter     int     `json:"buffer_after_minutes,omitempty"`
	PriceCents      int64   `json:"price_cents"`
	Currency        string  `json:"currency,omitempty"`
	MaxCapacity     int     `json:"max_capacity,omitempty"`
	RequiresResource bool   `json:"requires_resource,omitempty"`
	Color           string  `json:"color,omitempty"`
	IsPublic        bool    `json:"is_public,omitempty"`
}

// CreateService handles POST /api/v1/organizations/{orgID}/services
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateServiceRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		parsed, parseErr := uuid.Parse(*req.CategoryID)
		if parseErr == nil {
			categoryID = &parsed
		}
	}

	cmd := service.CreateServiceCmd{
		OrganizationID:  orgID,
		CategoryID:      categoryID,
		Name:            req.Name,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		BufferBefore:    req.BufferBefore,
		BufferAfter:     req.BufferAfter,
		PriceCents:      req.PriceCents,
		Currency:        req.Currency,
		MaxCapacity:     req.MaxCapacity,
		RequiresResource: req.RequiresResource,
		Color:           req.Color,
		IsPublic:        req.IsPublic,
	}

	svc, err := h.serviceSvc.CreateService(r.Context(), authCtx, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, svc)
}

// ListServices handles GET /api/v1/organizations/{orgID}/services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := service.ListServicesFilter{
		OrganizationID: orgID,
		Search:         q.Get("search"),
	}

	if catStr := q.Get("category_id"); catStr != "" {
		if parsed, err := uuid.Parse(catStr); err == nil {
			filter.CategoryID = &parsed
		}
	}

	if statusStr := q.Get("status"); statusStr != "" {
		st := service.Status(statusStr)
		filter.Status = &st
	}

	services, total, err := h.serviceSvc.ListServices(r.Context(), authCtx, filter, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondList(w, r, http.StatusOK, services, nil)
	_ = total
}

// GetService handles GET /api/v1/organizations/{orgID}/services/{serviceID}
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	serviceIDStr := chi.URLParam(r, "serviceID")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid serviceID"))
		return
	}

	svc, err := h.serviceSvc.GetService(r.Context(), authCtx, orgID, serviceID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, svc)
}

type UpdateStatusRequest struct {
	Status service.Status `json:"status"`
}

// ToggleStatus handles PATCH /api/v1/organizations/{orgID}/services/{serviceID}/status
func (h *ServiceHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	serviceIDStr := chi.URLParam(r, "serviceID")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid serviceID"))
		return
	}

	var req UpdateStatusRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	svc, err := h.serviceSvc.ToggleServiceStatus(r.Context(), authCtx, orgID, serviceID, req.Status)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, svc)
}

// DeleteService handles DELETE /api/v1/organizations/{orgID}/services/{serviceID}
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	serviceIDStr := chi.URLParam(r, "serviceID")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid serviceID"))
		return
	}

	if err := h.serviceSvc.DeleteService(r.Context(), authCtx, orgID, serviceID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}
