package location

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/location"
)

type LocationHandler struct {
	locService *usecase.Service
}

func NewLocationHandler(locService *usecase.Service) *LocationHandler {
	return &LocationHandler{
		locService: locService,
	}
}

type CreateLocationRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Email        string   `json:"email,omitempty"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 string   `json:"address_line2,omitempty"`
	City         string   `json:"city"`
	State        string   `json:"state,omitempty"`
	PostalCode   string   `json:"postal_code,omitempty"`
	Country      string   `json:"country"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Timezone     string   `json:"timezone,omitempty"`
	IsDefault    bool     `json:"is_default,omitempty"`
}

// CreateLocation handles POST /api/v1/organizations/{orgID}/locations
func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateLocationRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := location.CreateLocationCmd{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		Phone:          req.Phone,
		Email:          req.Email,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Timezone:       req.Timezone,
		IsDefault:      req.IsDefault,
	}

	loc, err := h.locService.CreateLocation(r.Context(), authCtx, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, loc)
}

// ListLocations handles GET /api/v1/organizations/{orgID}/locations
func (h *LocationHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	locations, err := h.locService.ListLocations(r.Context(), authCtx, orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, locations)
}

// GetLocation handles GET /api/v1/organizations/{orgID}/locations/{locationID}
func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	locationIDStr := chi.URLParam(r, "locationID")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid locationID"))
		return
	}

	loc, err := h.locService.GetLocation(r.Context(), authCtx, orgID, locationID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, loc)
}

type UpdateLocationRequest struct {
	Name         *string          `json:"name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Phone        *string          `json:"phone,omitempty"`
	Email        *string          `json:"email,omitempty"`
	AddressLine1 *string          `json:"address_line1,omitempty"`
	AddressLine2 *string          `json:"address_line2,omitempty"`
	City         *string          `json:"city,omitempty"`
	State        *string          `json:"state,omitempty"`
	PostalCode   *string          `json:"postal_code,omitempty"`
	Timezone     *string          `json:"timezone,omitempty"`
	Status       *location.Status `json:"status,omitempty"`
	IsDefault    *bool            `json:"is_default,omitempty"`
}

// UpdateLocation handles PUT /api/v1/organizations/{orgID}/locations/{locationID}
func (h *LocationHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	locationIDStr := chi.URLParam(r, "locationID")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid locationID"))
		return
	}

	var req UpdateLocationRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := location.UpdateLocationCmd{
		ID:             locationID,
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		Phone:          req.Phone,
		Email:          req.Email,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Timezone:       req.Timezone,
		Status:         req.Status,
		IsDefault:      req.IsDefault,
	}

	loc, err := h.locService.UpdateLocation(r.Context(), authCtx, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, loc)
}

// DeleteLocation handles DELETE /api/v1/organizations/{orgID}/locations/{locationID}
func (h *LocationHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	locationIDStr := chi.URLParam(r, "locationID")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid locationID"))
		return
	}

	if err := h.locService.DeleteLocation(r.Context(), authCtx, orgID, locationID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}
