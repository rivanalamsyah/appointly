package resource

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/resource"
)

type ResourceHandler struct {
	resourceSvc *usecase.Service
}

func NewResourceHandler(resourceSvc *usecase.Service) *ResourceHandler {
	return &ResourceHandler{
		resourceSvc: resourceSvc,
	}
}

type CreateResourceRequest struct {
	LocationID  *string                 `json:"location_id,omitempty"`
	Name        string                  `json:"name"`
	Type        resource.ResourceType   `json:"type"`
	Capacity    int                     `json:"capacity"`
	Status      resource.ResourceStatus `json:"status"`
	Description string                  `json:"description,omitempty"`
}

type UpdateResourceRequest struct {
	LocationID  *string                 `json:"location_id,omitempty"`
	Name        string                  `json:"name"`
	Type        resource.ResourceType   `json:"type"`
	Capacity    int                     `json:"capacity"`
	Status      resource.ResourceStatus `json:"status"`
	Description string                  `json:"description,omitempty"`
}

type AssignServiceResourcesRequest struct {
	ResourceIDs []string `json:"resource_ids"`
}

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	filter := resource.ListResourcesFilter{
		OrganizationID: orgID,
		Search:         r.URL.Query().Get("search"),
	}

	if locStr := r.URL.Query().Get("location_id"); locStr != "" {
		if locID, parseErr := uuid.Parse(locStr); parseErr == nil {
			filter.LocationID = &locID
		}
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		resType := resource.ResourceType(typeStr)
		filter.Type = &resType
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		resStatus := resource.ResourceStatus(statusStr)
		filter.Status = &resStatus
	}

	resources, err := h.resourceSvc.ListResources(r.Context(), filter)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, resources)
}

func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateResourceRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var locationID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		if parsed, parseErr := uuid.Parse(*req.LocationID); parseErr == nil {
			locationID = &parsed
		}
	}

	cmd := resource.CreateResourceCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		Name:           req.Name,
		Type:           req.Type,
		Capacity:       req.Capacity,
		Status:         req.Status,
		Description:    req.Description,
	}

	created, err := h.resourceSvc.CreateResource(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, created)
}

func (h *ResourceHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	resIDStr := chi.URLParam(r, "resourceID")
	resID, err := uuid.Parse(resIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid resourceID"))
		return
	}

	res, err := h.resourceSvc.GetResource(r.Context(), resID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, res)
}

func (h *ResourceHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	resIDStr := chi.URLParam(r, "resourceID")
	resID, err := uuid.Parse(resIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid resourceID"))
		return
	}

	var req UpdateResourceRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var locationID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		if parsed, parseErr := uuid.Parse(*req.LocationID); parseErr == nil {
			locationID = &parsed
		}
	}

	cmd := resource.UpdateResourceCmd{
		ID:          resID,
		LocationID:  locationID,
		Name:        req.Name,
		Type:        req.Type,
		Capacity:    req.Capacity,
		Status:      req.Status,
		Description: req.Description,
	}

	updated, err := h.resourceSvc.UpdateResource(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, updated)
}

func (h *ResourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	resIDStr := chi.URLParam(r, "resourceID")
	resID, err := uuid.Parse(resIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid resourceID"))
		return
	}

	if err := h.resourceSvc.DeleteResource(r.Context(), resID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "resource deleted successfully"})
}

func (h *ResourceHandler) GetServiceResources(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "serviceID")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid serviceID"))
		return
	}

	resources, err := h.resourceSvc.GetServiceResources(r.Context(), serviceID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, resources)
}

func (h *ResourceHandler) AssignServiceResources(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "serviceID")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid serviceID"))
		return
	}

	var req AssignServiceResourcesRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	resourceIDs := make([]uuid.UUID, 0, len(req.ResourceIDs))
	for _, idStr := range req.ResourceIDs {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			resourceIDs = append(resourceIDs, parsed)
		}
	}

	assigned, err := h.resourceSvc.AssignServiceResources(r.Context(), serviceID, resourceIDs)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, assigned)
}

// Helper registration for Chi router
func (h *ResourceHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/organizations/{orgID}/resources", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.ListResources)
		r.Post("/", h.CreateResource)
		r.Get("/{resourceID}", h.GetResource)
		r.Put("/{resourceID}", h.UpdateResource)
		r.Delete("/{resourceID}", h.DeleteResource)
	})

	r.Route("/organizations/{orgID}/services/{serviceID}/resources", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.GetServiceResources)
		r.Put("/", h.AssignServiceResources)
	})
}
