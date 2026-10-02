package public

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// PublicHandler serves unauthenticated customer-facing API endpoints.
type PublicHandler struct {
	availabilityService scheduling.AvailabilityService
	appointmentService  appointment.Service
}

// NewPublicHandler constructs a PublicHandler instance.
func NewPublicHandler(schedSvc scheduling.AvailabilityService, apptSvc appointment.Service) *PublicHandler {
	return &PublicHandler{
		availabilityService: schedSvc,
		appointmentService:  apptSvc,
	}
}

// SearchSlotsRequest defines query params for dynamic slot search.
type SearchSlotsRequest struct {
	ServiceID string `json:"service_id"`
	StaffID   string `json:"staff_id,omitempty"`
	Date      string `json:"date"` // YYYY-MM-DD
}

// GetAvailableSlots handles GET /api/v1/public/slots
func (h *PublicHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant, err := middleware.TenantFromContext(ctx)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	q := r.URL.Query()
	serviceIDStr := q.Get("service_id")
	if serviceIDStr == "" {
		v1.RespondError(w, r, apperror.BadRequest("service_id query parameter is required"))
		return
	}

	serviceID, parseErr := uuid.Parse(serviceIDStr)
	if parseErr != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid service_id UUID"))
		return
	}

	dateStr := q.Get("date")
	if dateStr == "" {
		v1.RespondError(w, r, apperror.BadRequest("date query parameter is required (YYYY-MM-DD)"))
		return
	}

	targetDate, parseErr := time.Parse("2006-01-02", dateStr)
	if parseErr != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var staffID *uuid.UUID
	if staffStr := q.Get("staff_id"); staffStr != "" {
		id, err := uuid.Parse(staffStr)
		if err != nil {
			v1.RespondError(w, r, apperror.BadRequest("invalid staff_id UUID"))
			return
		}
		staffID = &id
	}

	req := scheduling.AvailabilityRequest{
		OrganizationID: tenant.ID,
		ServiceID:      serviceID,
		StaffID:        staffID,
		Date:           targetDate,
		Timezone:       "UTC",
	}

	result, searchErr := h.availabilityService.GetAvailableSlots(ctx, req)
	if searchErr != nil {
		v1.RespondError(w, r, searchErr)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, result)
}

// CreateBookingRequest defines payload for public appointment creation.
type CreateBookingRequest struct {
	ServiceID        uuid.UUID  `json:"service_id"`
	StaffID          *uuid.UUID `json:"staff_id,omitempty"`
	LocationID       *uuid.UUID `json:"location_id,omitempty"`
	StartTime        time.Time  `json:"start_time"`
	CustomerName     string     `json:"customer_name"`
	CustomerEmail    string     `json:"customer_email"`
	CustomerPhone    string     `json:"customer_phone"`
	Notes            string     `json:"notes,omitempty"`
	IdempotencyKey   string     `json:"idempotency_key,omitempty"`
}

// CreatePublicBooking handles POST /api/v1/public/appointments
func (h *PublicHandler) CreatePublicBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant, err := middleware.TenantFromContext(ctx)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var req CreateBookingRequest
	if appErr := v1.DecodeJSON(r, &req); appErr != nil {
		v1.RespondError(w, r, appErr)
		return
	}

	// Basic validation
	if req.CustomerName == "" || req.CustomerEmail == "" || req.CustomerPhone == "" {
		v1.RespondError(w, r, apperror.ValidationFailed("customer name, email, and phone are required"))
		return
	}

	var staffID uuid.UUID
	if req.StaffID != nil {
		staffID = *req.StaffID
	}

	var locationID uuid.UUID
	if req.LocationID != nil {
		locationID = *req.LocationID
	}

	cmd := appointment.CreateAppointmentCmd{
		OrganizationID: tenant.ID,
		ServiceID:      req.ServiceID,
		StaffID:        staffID,
		LocationID:     locationID,
		StartTime:      req.StartTime,
		GuestName:      req.CustomerName,
		GuestEmail:     req.CustomerEmail,
		GuestPhone:     req.CustomerPhone,
		Notes:          req.Notes,
		Source:         "online",
	}

	appt, createErr := h.appointmentService.CreateAppointment(ctx, cmd)
	if createErr != nil {
		v1.RespondError(w, r, createErr)
		return
	}

	v1.RespondCreated(w, r, appt)
}
