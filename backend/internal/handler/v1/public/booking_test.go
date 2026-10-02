package public_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	publichandler "github.com/appointly/appointly/backend/internal/handler/v1/public"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
	bookinguc "github.com/appointly/appointly/backend/internal/usecase/booking"
)

func setupTestEnvironment(t *testing.T) (*chi.Mux, *organization.Organization, *service.Service, *staff.Staff) {
	orgRepo := memory.NewOrgRepository()
	serviceRepo := memory.NewServiceRepository()
	staffRepo := memory.NewStaffRepository()
	locRepo := memory.NewLocationRepository()
	schedRepo := memory.NewSchedulingRepository()
	resRepo := memory.NewResourceRepository()
	apptRepo := memory.NewAppointmentRepository()

	org, err := orgRepo.Create(context.Background(), &organization.Organization{
		ID:           uuid.New(),
		Name:         "Luxe Salon",
		Slug:         "luxe-salon",
		BusinessType: organization.BusinessTypeSalon,
		Email:        "contact@luxesalon.com",
		Timezone:     "America/New_York",
		Currency:     "USD",
		OwnerID:      uuid.New(),
		Settings: organization.Settings{
			AllowOnlineBooking:  true,
			AutoConfirmBookings: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	svc, err := serviceRepo.CreateService(context.Background(), service.CreateServiceCmd{
		OrganizationID:  org.ID,
		Name:            "Signature Haircut",
		Description:     "Luxury haircut service",
		DurationMinutes: 45,
		PriceCents:      6500,
		Currency:        "USD",
		IsPublic:        true,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	st, err := staffRepo.Create(context.Background(), staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Alex",
		LastName:       "Wright",
		Email:          "alex.private@salon.com",
		Phone:          "+1555123456",
		Title:          "Master Stylist",
		Bio:            "Top hairdresser with 10 years experience",
		AcceptsOnline:  true,
		ServiceIDs:     []uuid.UUID{svc.ID},
	})
	if err != nil {
		t.Fatalf("failed to create staff: %v", err)
	}
	_ = staffRepo.SetStaffServices(context.Background(), org.ID, st.ID, []uuid.UUID{svc.ID})

	availEngine := availabilityuc.NewEngine(orgRepo, locRepo, serviceRepo, staffRepo, schedRepo, resRepo, apptRepo)
	bookingSvc := bookinguc.NewService(apptRepo, orgRepo, locRepo, serviceRepo, staffRepo, schedRepo, resRepo, availEngine)

	h := publichandler.NewPublicHandler(orgRepo, serviceRepo, staffRepo, locRepo, availEngine, bookingSvc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	return r, org, svc, st
}

type APIEnvelope struct {
	Data json.RawMessage `json:"data"`
}

func TestGetPublicOrganization(t *testing.T) {
	r, _, _, _ := setupTestEnvironment(t)

	req := httptest.NewRequest(http.MethodGet, "/public/orgs/luxe-salon", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var env APIEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)

	var res map[string]interface{}
	_ = json.Unmarshal(env.Data, &res)
	if res["slug"] != "luxe-salon" || res["name"] != "Luxe Salon" {
		t.Errorf("unexpected org profile response: %v", res)
	}
}

func TestGetPublicServices(t *testing.T) {
	r, _, svc, _ := setupTestEnvironment(t)

	req := httptest.NewRequest(http.MethodGet, "/public/orgs/luxe-salon/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var env APIEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)

	var res struct {
		Services []service.Service `json:"services"`
	}
	_ = json.Unmarshal(env.Data, &res)
	if len(res.Services) != 1 || res.Services[0].ID != svc.ID {
		t.Errorf("expected 1 public service, got %v", res.Services)
	}
}

func TestGetPublicStaff_Sanitization(t *testing.T) {
	r, _, _, st := setupTestEnvironment(t)

	req := httptest.NewRequest(http.MethodGet, "/public/orgs/luxe-salon/staff", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var env APIEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)

	var staffList []publichandler.PublicStaffDTO
	_ = json.Unmarshal(env.Data, &staffList)
	if len(staffList) != 1 || staffList[0].ID != st.ID {
		t.Fatalf("expected 1 public staff DTO, got %v", staffList)
	}

	if staffList[0].Name != "Alex Wright" {
		t.Errorf("expected full name Alex Wright, got %s", staffList[0].Name)
	}
}

func TestCreatePublicBooking_PriceTamperingProtection(t *testing.T) {
	r, _, svc, st := setupTestEnvironment(t)

	// 1. Attacker attempts to inject fake "price_cents" field -> Disallowed by strict decoder
	tamperedPayload := map[string]interface{}{
		"service_id":     svc.ID,
		"staff_id":       st.ID,
		"start_time":     time.Now().Add(24 * time.Hour).Truncate(time.Minute).Format(time.RFC3339),
		"customer_name":  "Jane Attacker",
		"customer_email": "jane@example.com",
		"customer_phone": "+1555987654",
		"price_cents":    100, // Unauthorized price parameter
	}

	bodyBytes, _ := json.Marshal(tamperedPayload)
	req := httptest.NewRequest(http.MethodPost, "/public/orgs/luxe-salon/appointments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request when unknown fields are passed, got %d", w.Code)
	}

	// 2. Legitimate booking creation -> Server enforces service price ($65.00)
	validPayload := map[string]interface{}{
		"service_id":     svc.ID,
		"staff_id":       st.ID,
		"start_time":     time.Now().Add(24 * time.Hour).Truncate(time.Minute).Format(time.RFC3339),
		"customer_name":  "Jane Public",
		"customer_email": "jane@example.com",
		"customer_phone": "+1555987654",
		"notes":          "Prefers morning appointment",
	}

	validBodyBytes, _ := json.Marshal(validPayload)
	reqValid := httptest.NewRequest(http.MethodPost, "/public/orgs/luxe-salon/appointments", bytes.NewBuffer(validBodyBytes))
	reqValid.Header.Set("Content-Type", "application/json")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusCreated {
		t.Fatalf("expected booking creation status 201, got %d: %s", wValid.Code, wValid.Body.String())
	}

	var env APIEnvelope
	_ = json.Unmarshal(wValid.Body.Bytes(), &env)

	var created map[string]interface{}
	_ = json.Unmarshal(env.Data, &created)

	if priceCents, ok := created["price_cents"].(float64); !ok || int64(priceCents) != 6500 {
		t.Errorf("SECURITY FAILURE: Expected price snapshot to be 6500 cents, got %v", created["price_cents"])
	}
}

func TestCreatePublicBooking_MissingFieldsValidation(t *testing.T) {
	r, _, svc, _ := setupTestEnvironment(t)

	invalidPayload := map[string]interface{}{
		"service_id":    svc.ID,
		"start_time":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"customer_name": "Jane Public",
	}

	bodyBytes, _ := json.Marshal(invalidPayload)
	req := httptest.NewRequest(http.MethodPost, "/public/orgs/luxe-salon/appointments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected validation error for missing fields, got %d", w.Code)
	}
}
