package public

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/availability"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
)

// PublicOrgDTO represents sanitized organization metadata for public booking pages.
type PublicOrgDTO struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	LogoURL     string                `json:"logo_url,omitempty"`
	Description string                `json:"description,omitempty"`
	Phone       string                `json:"phone,omitempty"`
	Email       string                `json:"email,omitempty"`
	Timezone    string                `json:"timezone"`
	Currency    string                `json:"currency"`
	Settings    organization.Settings `json:"booking_settings"`
}

// PublicStaffDTO represents sanitized staff profile without sensitive credentials/email.
type PublicStaffDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Title     string    `json:"title,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

// PublicHandler serves unauthenticated customer-facing API endpoints.
type PublicHandler struct {
	orgRepo            organization.Repository
	serviceRepo        service.Repository
	staffRepo          staff.Repository
	locationRepo       location.Repository
	availabilityEngine *availabilityuc.Engine
	appointmentService appointment.Service
}

// NewPublicHandler constructs a PublicHandler instance.
func NewPublicHandler(
	orgRepo organization.Repository,
	serviceRepo service.Repository,
	staffRepo staff.Repository,
	locationRepo location.Repository,
	availabilityEngine *availabilityuc.Engine,
	apptSvc appointment.Service,
) *PublicHandler {
	return &PublicHandler{
		orgRepo:            orgRepo,
		serviceRepo:        serviceRepo,
		staffRepo:          staffRepo,
		locationRepo:       locationRepo,
		availabilityEngine: availabilityEngine,
		appointmentService:  apptSvc,
	}
}

// GetPublicOrganization handles GET /api/v1/public/orgs/{slug}
func (h *PublicHandler) GetPublicOrganization(w http.ResponseWriter, r *http.Request) {
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

	dto := PublicOrgDTO{
		ID:          org.ID,
		Name:        org.Name,
		Slug:        org.Slug,
		LogoURL:     org.LogoURL,
		Description: org.Description,
		Phone:       org.Phone,
		Email:       org.Email,
		Timezone:    org.Timezone,
		Currency:    org.Currency,
		Settings:    org.Settings,
	}

	v1.RespondJSON(w, r, http.StatusOK, dto)
}

// GetPublicServices handles GET /api/v1/public/orgs/{slug}/services
func (h *PublicHandler) GetPublicServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
		return
	}

	services, err := h.serviceRepo.ListPublicServices(ctx, org.ID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	categories, _ := h.serviceRepo.ListPublicCategories(ctx, org.ID)

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"services":   services,
		"categories": categories,
	})
}

// GetPublicStaff handles GET /api/v1/public/orgs/{slug}/staff
func (h *PublicHandler) GetPublicStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
		return
	}

	var serviceID *uuid.UUID
	if sIDStr := r.URL.Query().Get("service_id"); sIDStr != "" {
		if parsed, pErr := uuid.Parse(sIDStr); pErr == nil {
			serviceID = &parsed
		}
	}

	activeTrue := true
	onlineTrue := true
	filter := staff.ListStaffFilter{
		OrganizationID: org.ID,
		ServiceID:      serviceID,
		IsActive:       &activeTrue,
		AcceptsOnline:  &onlineTrue,
	}

	staffList, _, err := h.staffRepo.List(ctx, filter, 1, 100)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	dtos := make([]PublicStaffDTO, 0, len(staffList))
	for _, st := range staffList {
		dtos = append(dtos, PublicStaffDTO{
			ID:        st.ID,
			Name:      st.FullName(),
			Title:     st.Title,
			Bio:       st.Bio,
			AvatarURL: st.AvatarURL,
		})
	}

	v1.RespondJSON(w, r, http.StatusOK, dtos)
}

// GetPublicLocations handles GET /api/v1/public/orgs/{slug}/locations
func (h *PublicHandler) GetPublicLocations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
		return
	}

	locations, err := h.locationRepo.List(ctx, org.ID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, locations)
}

// GetPublicAvailability handles GET /api/v1/public/orgs/{slug}/availability
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
	if tz == "" {
		tz = org.Timezone
	}

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

// CreatePublicBooking handles POST /api/v1/public/orgs/{slug}/appointments
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

	// Server enforces organization_id, price_cents, duration, and status derived from single source of truth
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
		Source:         "online_public",
	}

	appt, createErr := h.appointmentService.CreateAppointment(ctx, cmd)
	if createErr != nil {
		v1.RespondError(w, r, createErr)
		return
	}

	v1.RespondCreated(w, r, appt)
}

// RegisterRoutes mounts all public booking endpoints under chi router.
func (h *PublicHandler) RegisterRoutes(r chi.Router) {
	r.Route("/public/orgs/{slug}", func(r chi.Router) {
		r.Get("/", h.GetPublicOrganization)
		r.Get("/services", h.GetPublicServices)
		r.Get("/staff", h.GetPublicStaff)
		r.Get("/locations", h.GetPublicLocations)
		r.Get("/availability", h.GetPublicAvailability)
		r.Post("/appointments", h.CreatePublicBooking)
	})
	// Backward compatibility alias route
	r.Route("/public/{slug}", func(r chi.Router) {
		r.Get("/", h.GetPublicOrganization)
		r.Get("/services", h.GetPublicServices)
		r.Get("/staff", h.GetPublicStaff)
		r.Get("/locations", h.GetPublicLocations)
		r.Get("/availability", h.GetPublicAvailability)
		r.Post("/appointments", h.CreatePublicBooking)
	})
}
