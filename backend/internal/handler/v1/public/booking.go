package public

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/availability"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
)

// PublicHandler serves unauthenticated customer-facing API endpoints.
type PublicHandler struct {
	orgRepo            organization.Repository
	availabilityEngine *availabilityuc.Engine
	appointmentService appointment.Service
}

// NewPublicHandler constructs a PublicHandler instance.
func NewPublicHandler(
	orgRepo organization.Repository,
	availabilityEngine *availabilityuc.Engine,
	apptSvc appointment.Service,
) *PublicHandler {
	return &PublicHandler{
		orgRepo:            orgRepo,
		availabilityEngine: availabilityEngine,
		appointmentService:  apptSvc,
	}
}

// GetPublicAvailability handles GET /api/v1/public/{slug}/availability
func (h *PublicHandler) GetPublicAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		v1.RespondError(w, r, apperror.BadRequest("organization slug is required"))
		return
	}

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
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

	dateFromStr := q.Get("date_from")
	if dateFromStr == "" {
		dateFromStr = q.Get("date") // fallback
	}

	now := time.Now().UTC()
	dateFrom := now
	if dateFromStr != "" {
		if parsed, pErr := time.Parse("2006-01-02", dateFromStr); pErr == nil {
			dateFrom = parsed
		}
	}

	dateToStr := q.Get("date_to")
	dateTo := dateFrom
	if dateToStr != "" {
		if parsed, pErr := time.Parse("2006-01-02", dateToStr); pErr == nil {
			dateTo = parsed
		}
	}

	var locationID *uuid.UUID
	if locStr := q.Get("location_id"); locStr != "" {
		if id, pErr := uuid.Parse(locStr); pErr == nil {
			locationID = &id
		}
	}

	var staffID *uuid.UUID
	if staffStr := q.Get("staff_id"); staffStr != "" {
		if id, pErr := uuid.Parse(staffStr); pErr == nil {
			staffID = &id
		}
	}

	tz := q.Get("timezone")

	availQuery := availability.GetAvailabilityQuery{
		OrganizationID: org.ID,
		LocationID:     locationID,
		ServiceID:      serviceID,
		StaffID:        staffID,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		Timezone:       tz,
	}

	slots, searchErr := h.availabilityEngine.GetAvailableSlots(ctx, availQuery)
	if searchErr != nil {
		v1.RespondError(w, r, searchErr)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, slots)
}

// CreateBookingRequest defines payload for public appointment creation.
type CreateBookingRequest struct {
	ServiceID      uuid.UUID  `json:"service_id"`
	StaffID        *uuid.UUID `json:"staff_id,omitempty"`
	LocationID     *uuid.UUID `json:"location_id,omitempty"`
	StartTime      time.Time  `json:"start_time"`
	CustomerName   string     `json:"customer_name"`
	CustomerEmail  string     `json:"customer_email"`
	CustomerPhone  string     `json:"customer_phone"`
	Notes          string     `json:"notes,omitempty"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
}

// CreatePublicBooking handles POST /api/v1/public/{slug}/appointments
func (h *PublicHandler) CreatePublicBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		v1.RespondError(w, r, apperror.BadRequest("organization slug is required"))
		return
	}

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
		return
	}

	var req CreateBookingRequest
	if appErr := v1.DecodeJSON(r, &req); appErr != nil {
		v1.RespondError(w, r, appErr)
		return
	}

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
		OrganizationID: org.ID,
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

// Helper registration for Chi router
func (h *PublicHandler) RegisterRoutes(r chi.Router) {
	r.Route("/public/{slug}", func(r chi.Router) {
		r.Get("/availability", h.GetPublicAvailability)
		r.Post("/appointments", h.CreatePublicBooking)
	})
}
