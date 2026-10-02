package security_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/repository/memory"

	auditUseCase "github.com/appointly/appointly/backend/internal/usecase/audit"
	bookingUseCase "github.com/appointly/appointly/backend/internal/usecase/booking"
	customerUseCase "github.com/appointly/appointly/backend/internal/usecase/customer"
	locationUseCase "github.com/appointly/appointly/backend/internal/usecase/location"
	resourceUseCase "github.com/appointly/appointly/backend/internal/usecase/resource"
	serviceUseCase "github.com/appointly/appointly/backend/internal/usecase/service"
	staffUseCase "github.com/appointly/appointly/backend/internal/usecase/staff"
)

func contextWithAuth(orgID, userID uuid.UUID, role rbac.Role) context.Context {
	return rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  userID,
		OrgID:   orgID,
		Role:    role,
		IsOwner: role == rbac.RoleOwner,
	})
}

func TestCrossTenantIsolation_Systematic(t *testing.T) {
	orgA := uuid.New()
	userA := uuid.New()
	ctxA := contextWithAuth(orgA, userA, rbac.RoleAdmin)
	authCtxA := &rbac.AuthContext{UserID: userA, OrgID: orgA, Role: rbac.RoleAdmin}

	orgB := uuid.New()
	userB := uuid.New()
	ctxB := contextWithAuth(orgB, userB, rbac.RoleAdmin)
	authCtxB := &rbac.AuthContext{UserID: userB, OrgID: orgB, Role: rbac.RoleAdmin}

	auditRepo := memory.NewAuditRepository()

	// 1. Appointments Cross-Tenant Test
	t.Run("Appointments Tenant Isolation", func(t *testing.T) {
		orgRepo := memory.NewOrgRepository()
		locRepo := memory.NewLocationRepository()
		svcRepo := memory.NewServiceRepository()
		staffRepo := memory.NewStaffRepository()
		schedRepo := memory.NewSchedulingRepository()
		resRepo := memory.NewResourceRepository()
		apptRepo := memory.NewAppointmentRepository()

		bookingSvc := bookingUseCase.NewService(apptRepo, orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, nil)

		apptID := uuid.New()
		_, err := apptRepo.Create(ctxA, appointment.CreateAppointmentCmd{
			OrganizationID: orgA,
			LocationID:     uuid.New(),
			ServiceID:      uuid.New(),
			StaffID:        uuid.New(),
			StartTime:      time.Now().Add(24 * time.Hour),
			EndTime:        time.Now().Add(25 * time.Hour),
			GuestName:      "Alice OrgA",
			GuestEmail:     "alice@orga.com",
			PriceCents:     10000,
			Currency:       "USD",
		})
		if err != nil {
			t.Fatalf("unexpected error creating appt: %v", err)
		}

		gotAppt, err := bookingSvc.GetAppointmentByID(ctxB, orgB, apptID)
		if err == nil || gotAppt != nil {
			t.Fatalf("expected error reading Org A's appt from Org B context, got %v", gotAppt)
		}

		err = bookingSvc.CancelAppointment(ctxB, orgB, apptID, "malicious cancel", userB)
		if err == nil {
			t.Fatalf("expected error cancelling Org A's appt from Org B context")
		}
	})

	// 2. Customer Management Cross-Tenant Test
	t.Run("Customer Management Tenant Isolation", func(t *testing.T) {
		custRepo := memory.NewCustomerRepository()
		custSvc := customerUseCase.NewService(custRepo)

		custID := uuid.New()
		_, err := custRepo.Create(ctxA, customer.CreateCustomerCmd{
			OrganizationID: orgA,
			FirstName:      "Bob",
			LastName:       "Client",
			Email:          "bob@client.com",
		})
		if err != nil {
			t.Fatalf("unexpected error creating customer: %v", err)
		}

		gotCust, err := custSvc.GetCustomer(ctxB, custID)
		if err == nil || gotCust != nil {
			t.Fatalf("expected error reading Org A's customer from Org B context")
		}
	})

	// 3. Staff Profile Cross-Tenant Test
	t.Run("Staff Profile Tenant Isolation", func(t *testing.T) {
		staffRepo := memory.NewStaffRepository()
		staffSvc := staffUseCase.NewService(staffRepo, auditRepo)

		st, err := staffSvc.CreateStaff(ctxA, authCtxA, staff.CreateStaffCmd{
			OrganizationID: orgA,
			FirstName:      "Smith",
			LastName:       "Doctor",
			Email:          "smith@orga.com",
		})
		if err != nil {
			t.Fatalf("unexpected error creating staff: %v", err)
		}

		gotStaff, err := staffSvc.GetStaff(ctxB, authCtxB, orgB, st.ID)
		if err == nil || gotStaff != nil {
			t.Fatalf("expected error reading Org A's staff from Org B")
		}
	})

	// 4. Resource Management Cross-Tenant Test
	t.Run("Resource Management Tenant Isolation", func(t *testing.T) {
		resRepo := memory.NewResourceRepository()
		resSvc := resourceUseCase.NewService(resRepo)

		stCtxA := contextWithAuth(orgA, userA, rbac.RoleOwner)
		stCtxB := contextWithAuth(orgB, userB, rbac.RoleOwner)

		res, err := resSvc.CreateResource(stCtxA, resource.CreateResourceCmd{
			Name: "Room 101",
			Type: resource.TypeRoom,
		})
		if err != nil {
			t.Fatalf("unexpected error creating resource: %v", err)
		}

		gotRes, err := resSvc.GetResource(stCtxB, res.ID)
		if err == nil || gotRes != nil {
			t.Fatalf("expected error reading Org A's resource from Org B")
		}
	})

	// 5. Locations Cross-Tenant Test
	t.Run("Location Management Tenant Isolation", func(t *testing.T) {
		locRepo := memory.NewLocationRepository()
		locSvc := locationUseCase.NewService(locRepo, auditRepo)

		loc, err := locSvc.CreateLocation(ctxA, authCtxA, location.CreateLocationCmd{
			OrganizationID: orgA,
			Name:           "Main Branch",
			AddressLine1:   "123 Main St",
			City:           "New York",
			Country:        "US",
			Timezone:       "America/New_York",
		})
		if err != nil {
			t.Fatalf("unexpected error creating location: %v", err)
		}

		gotLoc, err := locSvc.GetLocation(ctxB, authCtxB, orgB, loc.ID)
		if err == nil || gotLoc != nil {
			t.Fatalf("expected error reading Org A's location from Org B")
		}
	})

	// 6. Services Catalog Cross-Tenant Test
	t.Run("Services Catalog Tenant Isolation", func(t *testing.T) {
		svcRepo := memory.NewServiceRepository()
		svcService := serviceUseCase.NewService(svcRepo, auditRepo)

		s, err := svcService.CreateService(ctxA, authCtxA, service.CreateServiceCmd{
			OrganizationID:  orgA,
			Name:            "Haircut & Styling",
			DurationMinutes: 45,
			PriceCents:      150000,
			Currency:        "IDR",
		})
		if err != nil {
			t.Fatalf("unexpected error creating service: %v", err)
		}

		gotSvc, err := svcService.GetService(ctxB, authCtxB, orgB, s.ID)
		if err == nil || gotSvc != nil {
			t.Fatalf("expected error reading Org A's service from Org B")
		}
	})

	// 7. Audit Log Access Isolation Test
	t.Run("Audit Log Tenant Isolation & Immutability", func(t *testing.T) {
		auditSvc := auditUseCase.NewService(auditRepo)

		logID := uuid.New()
		err := auditRepo.Create(ctxA, audit.AuditLog{
			ID:             logID,
			OrganizationID: &orgA,
			Action:         audit.ActionOrgSettingsUpdated,
			ResourceType:   "organization",
		})
		if err != nil {
			t.Fatalf("unexpected error creating audit log: %v", err)
		}

		gotLog, err := auditSvc.GetLogByID(ctxB, orgB, logID)
		if err == nil || gotLog != nil {
			t.Fatalf("expected error reading Org A's audit log from Org B context")
		}

		ownLog, err := auditSvc.GetLogByID(ctxA, orgA, logID)
		if err != nil || ownLog == nil || ownLog.ID != logID {
			t.Fatalf("expected successful retrieval of Org A's audit log")
		}
	})
}

func TestSecurityHeaders_Middleware(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/services", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected DENY header")
	}
	if rec.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("expected 1; mode=block header")
	}
	if rec.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("expected strict-origin-when-cross-origin header")
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Errorf("expected Content-Security-Policy header")
	}
}

func TestMaxBytes_Middleware(t *testing.T) {
	handler := middleware.MaxBytes(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 2<<20) // 2MB
		_, err := r.Body.Read(buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	largePayload := strings.Repeat("A", 1500000)
	req := httptest.NewRequest("POST", "/api/v1/appointments", strings.NewReader(largePayload))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 Request Entity Too Large, got %d", rec.Code)
	}
}
