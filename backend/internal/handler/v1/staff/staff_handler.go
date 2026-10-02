package staff

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/staff"
)

type StaffHandler struct {
	staffSvc *usecase.Service
}

func NewStaffHandler(staffSvc *usecase.Service) *StaffHandler {
	return &StaffHandler{
		staffSvc: staffSvc,
	}
}

type CreateStaffRequest struct {
	UserID        *string  `json:"user_id,omitempty"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	Title         string   `json:"title,omitempty"`
	Bio           string   `json:"bio,omitempty"`
	AcceptsOnline bool     `json:"accepts_online,omitempty"`
	ServiceIDs    []string `json:"service_ids,omitempty"`
}

// CreateStaff handles POST /api/v1/organizations/{orgID}/staff
func (h *StaffHandler) CreateStaff(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateStaffRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		if parsed, parseErr := uuid.Parse(*req.UserID); parseErr == nil {
			userID = &parsed
		}
	}

	var serviceIDs []uuid.UUID
	for _, idStr := range req.ServiceIDs {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			serviceIDs = append(serviceIDs, parsed)
		}
	}

	cmd := staff.CreateStaffCmd{
		OrganizationID: orgID,
		UserID:         userID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		Title:          req.Title,
		Bio:            req.Bio,
		AcceptsOnline:  req.AcceptsOnline,
		ServiceIDs:     serviceIDs,
	}

	st, err := h.staffSvc.CreateStaff(r.Context(), authCtx, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, st)
}

// ListStaff handles GET /api/v1/organizations/{orgID}/staff
func (h *StaffHandler) ListStaff(w http.ResponseWriter, r *http.Request) {
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

	filter := staff.ListStaffFilter{
		OrganizationID: orgID,
		Search:         q.Get("search"),
	}

	staffList, total, err := h.staffSvc.ListStaff(r.Context(), authCtx, filter, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondList(w, r, http.StatusOK, staffList, nil)
	_ = total
}

// GetStaff handles GET /api/v1/organizations/{orgID}/staff/{staffID}
func (h *StaffHandler) GetStaff(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	staffIDStr := chi.URLParam(r, "staffID")
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid staffID"))
		return
	}

	st, err := h.staffSvc.GetStaff(r.Context(), authCtx, orgID, staffID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, st)
}

type AssignServicesRequest struct {
	ServiceIDs []string `json:"service_ids"`
}

// SetStaffServices handles POST /api/v1/organizations/{orgID}/staff/{staffID}/services
func (h *StaffHandler) SetStaffServices(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	staffIDStr := chi.URLParam(r, "staffID")
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid staffID"))
		return
	}

	var req AssignServicesRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var serviceIDs []uuid.UUID
	for _, idStr := range req.ServiceIDs {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			serviceIDs = append(serviceIDs, parsed)
		}
	}

	if err := h.staffSvc.SetStaffServices(r.Context(), authCtx, orgID, staffID, serviceIDs); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}

// DeactivateStaff handles DELETE /api/v1/organizations/{orgID}/staff/{staffID}
func (h *StaffHandler) DeactivateStaff(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	staffIDStr := chi.URLParam(r, "staffID")
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid staffID"))
		return
	}

	if err := h.staffSvc.DeactivateStaff(r.Context(), authCtx, orgID, staffID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}
