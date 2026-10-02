package scheduling

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/scheduling"
)

type SchedulingHandler struct {
	schedSvc *usecase.Service
}

func NewSchedulingHandler(schedSvc *usecase.Service) *SchedulingHandler {
	return &SchedulingHandler{
		schedSvc: schedSvc,
	}
}

type UpsertBusinessHoursRequest struct {
	Hours []scheduling.BusinessHours `json:"hours"`
}

// GetBusinessHours handles GET /api/v1/organizations/{orgID}/locations/{locationID}/business-hours
func (h *SchedulingHandler) GetBusinessHours(w http.ResponseWriter, r *http.Request) {
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

	hours, err := h.schedSvc.GetBusinessHours(r.Context(), authCtx, orgID, locationID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, hours)
}

// UpsertBusinessHours handles PUT /api/v1/organizations/{orgID}/locations/{locationID}/business-hours
func (h *SchedulingHandler) UpsertBusinessHours(w http.ResponseWriter, r *http.Request) {
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

	var req UpsertBusinessHoursRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := scheduling.UpsertBusinessHoursCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		Hours:          req.Hours,
	}

	if err := h.schedSvc.UpsertBusinessHours(r.Context(), authCtx, cmd); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}

type UpsertStaffScheduleRequest struct {
	Schedule []scheduling.StaffSchedule `json:"schedule"`
}

// GetStaffSchedule handles GET /api/v1/organizations/{orgID}/staff/{staffID}/schedules
func (h *SchedulingHandler) GetStaffSchedule(w http.ResponseWriter, r *http.Request) {
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

	scheds, err := h.schedSvc.GetStaffSchedule(r.Context(), authCtx, orgID, staffID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, scheds)
}

// UpsertStaffSchedule handles PUT /api/v1/organizations/{orgID}/staff/{staffID}/schedules
func (h *SchedulingHandler) UpsertStaffSchedule(w http.ResponseWriter, r *http.Request) {
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

	var req UpsertStaffScheduleRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := scheduling.UpsertStaffScheduleCmd{
		OrganizationID: orgID,
		StaffID:        staffID,
		Schedule:       req.Schedule,
	}

	if err := h.schedSvc.UpsertStaffSchedule(r.Context(), authCtx, cmd); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}

type CreateTimeOffRequest struct {
	StartDate string  `json:"start_date"` // YYYY-MM-DD
	EndDate   string  `json:"end_date"`   // YYYY-MM-DD
	Reason    string  `json:"reason,omitempty"`
	IsAllDay  bool    `json:"is_all_day"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
}

// CreateTimeOff handles POST /api/v1/organizations/{orgID}/staff/{staffID}/time-offs
func (h *SchedulingHandler) CreateTimeOff(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTimeOffRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	sDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid start_date format (YYYY-MM-DD)"))
		return
	}
	eDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid end_date format (YYYY-MM-DD)"))
		return
	}

	cmd := scheduling.CreateTimeOffCmd{
		OrganizationID: orgID,
		StaffID:        staffID,
		StartDate:      sDate,
		EndDate:        eDate,
		Reason:         req.Reason,
		IsAllDay:       req.IsAllDay,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
	}

	tOff, err := h.schedSvc.CreateStaffTimeOff(r.Context(), authCtx, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, tOff)
}
