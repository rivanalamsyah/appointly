package appointment

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	"github.com/appointly/appointly/backend/internal/pkg/pagination"
	usecase "github.com/appointly/appointly/backend/internal/usecase/booking"
)

type AppointmentHandler struct {
	bookingSvc *usecase.Service
}

func NewAppointmentHandler(bookingSvc *usecase.Service) *AppointmentHandler {
	return &AppointmentHandler{
		bookingSvc: bookingSvc,
	}
}

type CreateAppointmentRequest struct {
	LocationID *string    `json:"location_id,omitempty"`
	ServiceID  string     `json:"service_id"`
	StaffID    string     `json:"staff_id"`
	CustomerID *string    `json:"customer_id,omitempty"`
	ResourceID *string    `json:"resource_id,omitempty"`
	StartTime  time.Time  `json:"start_time"`
	Timezone   string     `json:"timezone,omitempty"`
	GuestName  string     `json:"guest_name,omitempty"`
	GuestEmail string     `json:"guest_email,omitempty"`
	GuestPhone string     `json:"guest_phone,omitempty"`
	Notes      string     `json:"notes,omitempty"`
}

type RescheduleRequest struct {
	NewStartTime time.Time `json:"new_start_time"`
	Reason       string    `json:"reason,omitempty"`
}

type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

func (h *AppointmentHandler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	filter := appointment.ListFilter{
		OrganizationID: orgID,
		Search:         r.URL.Query().Get("search"),
	}

	if locStr := r.URL.Query().Get("location_id"); locStr != "" {
		if parsed, pErr := uuid.Parse(locStr); pErr == nil {
			filter.LocationID = &parsed
		}
	}
	if staffStr := r.URL.Query().Get("staff_id"); staffStr != "" {
		if parsed, pErr := uuid.Parse(staffStr); pErr == nil {
			filter.StaffID = &parsed
		}
	}
	if custStr := r.URL.Query().Get("customer_id"); custStr != "" {
		if parsed, pErr := uuid.Parse(custStr); pErr == nil {
			filter.CustomerID = &parsed
		}
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		st := appointment.Status(statusStr)
		filter.Status = &st
	}

	appts, total, err := h.bookingSvc.ListAppointments(r.Context(), filter, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	paginationMeta := pagination.NewOffsetMeta(pagination.OffsetParams{Page: page, PerPage: perPage}, total)
	v1.RespondList(w, r, http.StatusOK, appts, &paginationMeta)
}

func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateAppointmentRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	svcID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid service_id"))
		return
	}

	staffID, err := uuid.Parse(req.StaffID)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid staff_id"))
		return
	}

	var locID uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		if parsed, pErr := uuid.Parse(*req.LocationID); pErr == nil {
			locID = parsed
		}
	}

	var custID *uuid.UUID
	if req.CustomerID != nil && *req.CustomerID != "" {
		if parsed, pErr := uuid.Parse(*req.CustomerID); pErr == nil {
			custID = &parsed
		}
	}

	var resID *uuid.UUID
	if req.ResourceID != nil && *req.ResourceID != "" {
		if parsed, pErr := uuid.Parse(*req.ResourceID); pErr == nil {
			resID = &parsed
		}
	}

	authCtx := rbac.FromContext(r.Context())
	createdBy := uuid.Nil
	if authCtx != nil {
		createdBy = authCtx.UserID
	}

	cmd := appointment.CreateAppointmentCmd{
		OrganizationID: orgID,
		LocationID:     locID,
		ServiceID:      svcID,
		StaffID:        staffID,
		CustomerID:     custID,
		ResourceID:     resID,
		StartTime:      req.StartTime,
		Timezone:       req.Timezone,
		GuestName:      req.GuestName,
		GuestEmail:     req.GuestEmail,
		GuestPhone:     req.GuestPhone,
		Notes:          req.Notes,
		Source:         "manual",
		CreatedBy:      createdBy,
	}

	created, err := h.bookingSvc.CreateAppointment(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, created)
}

func (h *AppointmentHandler) GetAppointment(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	appt, err := h.bookingSvc.GetAppointmentByID(r.Context(), orgID, apptID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, appt)
}

func (h *AppointmentHandler) RescheduleAppointment(w http.ResponseWriter, r *http.Request) {
	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	var req RescheduleRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	authCtx := rbac.FromContext(r.Context())
	changedBy := uuid.Nil
	if authCtx != nil {
		changedBy = authCtx.UserID
	}

	cmd := appointment.RescheduleCmd{
		AppointmentID: apptID,
		NewStartTime:  req.NewStartTime,
		Reason:        req.Reason,
		ChangedBy:     changedBy,
	}

	rescheduled, err := h.bookingSvc.RescheduleAppointment(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, rescheduled)
}

func (h *AppointmentHandler) CancelAppointment(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	var req CancelRequest
	_ = v1.DecodeJSON(r, &req)

	authCtx := rbac.FromContext(r.Context())
	cancelledBy := uuid.Nil
	if authCtx != nil {
		cancelledBy = authCtx.UserID
	}

	if err := h.bookingSvc.CancelAppointment(r.Context(), orgID, apptID, req.Reason, cancelledBy); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "appointment cancelled successfully"})
}

func (h *AppointmentHandler) ConfirmAppointment(w http.ResponseWriter, r *http.Request) {
	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	authCtx := rbac.FromContext(r.Context())
	changedBy := uuid.Nil
	if authCtx != nil {
		changedBy = authCtx.UserID
	}

	cmd := appointment.UpdateStatusCmd{
		AppointmentID: apptID,
		NewStatus:     appointment.StatusConfirmed,
		Reason:        "Confirmed by staff/admin",
		ChangedBy:     changedBy,
	}

	if err := h.bookingSvc.UpdateAppointmentStatus(r.Context(), cmd); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "appointment confirmed"})
}

func (h *AppointmentHandler) CompleteAppointment(w http.ResponseWriter, r *http.Request) {
	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	authCtx := rbac.FromContext(r.Context())
	changedBy := uuid.Nil
	if authCtx != nil {
		changedBy = authCtx.UserID
	}

	cmd := appointment.UpdateStatusCmd{
		AppointmentID: apptID,
		NewStatus:     appointment.StatusCompleted,
		Reason:        "Service delivered",
		ChangedBy:     changedBy,
	}

	if err := h.bookingSvc.UpdateAppointmentStatus(r.Context(), cmd); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "appointment completed"})
}

func (h *AppointmentHandler) NoShowAppointment(w http.ResponseWriter, r *http.Request) {
	apptIDStr := chi.URLParam(r, "appointmentID")
	apptID, err := uuid.Parse(apptIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid appointmentID"))
		return
	}

	authCtx := rbac.FromContext(r.Context())
	changedBy := uuid.Nil
	if authCtx != nil {
		changedBy = authCtx.UserID
	}

	cmd := appointment.UpdateStatusCmd{
		AppointmentID: apptID,
		NewStatus:     appointment.StatusNoShow,
		Reason:        "Customer did not show up",
		ChangedBy:     changedBy,
	}

	if err := h.bookingSvc.UpdateAppointmentStatus(r.Context(), cmd); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "appointment marked as no-show"})
}

// Helper registration for Chi router
func (h *AppointmentHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/organizations/{orgID}/appointments", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.ListAppointments)
		r.Post("/", h.CreateAppointment)
		r.Get("/{appointmentID}", h.GetAppointment)
		r.Post("/{appointmentID}/reschedule", h.RescheduleAppointment)
		r.Post("/{appointmentID}/cancel", h.CancelAppointment)
		r.Post("/{appointmentID}/confirm", h.ConfirmAppointment)
		r.Post("/{appointmentID}/complete", h.CompleteAppointment)
		r.Post("/{appointmentID}/no-show", h.NoShowAppointment)
	})
}
