package booking_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
	bookinguc "github.com/appointly/appointly/backend/internal/usecase/booking"
)

func TestCalendar_FiltersAndDateNavigation(t *testing.T) {
	orgRepo := memory.NewOrgRepository()
	locRepo := memory.NewLocationRepository()
	svcRepo := memory.NewServiceRepository()
	staffRepo := memory.NewStaffRepository()
	schedRepo := memory.NewSchedulingRepository()
	resRepo := memory.NewResourceRepository()
	apptRepo := memory.NewAppointmentRepository()

	engine := availabilityuc.NewEngine(orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, apptRepo)
	svc := bookinguc.NewService(apptRepo, orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, engine)

	ctx := context.Background()

	// 1. Create Org
	org, err := orgRepo.Create(ctx, &organization.Organization{
		ID:           uuid.New(),
		Name:         "Salon Calendar Test",
		Slug:         "salon-cal",
		Timezone:     "America/New_York",
		Currency:     "USD",
		OwnerID:      uuid.New(),
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	// 2. Create Service & Staff
	serviceItem, err := svcRepo.CreateService(ctx, service.CreateServiceCmd{
		OrganizationID:  org.ID,
		Name:            "Hair Styling",
		DurationMinutes: 60,
		PriceCents:      8000,
		Currency:        "USD",
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	st1, _ := staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Alice",
		LastName:       "Stylist",
	})
	st2, _ := staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Bob",
		LastName:       "Barber",
	})

	// 3. Create Appointments for Alice and Bob
	t1 := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	t2 := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Hour)

	appt1, err := svc.CreateAppointment(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: org.ID,
		ServiceID:      serviceItem.ID,
		StaffID:        st1.ID,
		StartTime:      t1,
		GuestName:      "Client One",
	})
	if err != nil {
		t.Fatalf("failed to create appt 1: %v", err)
	}

	appt2, err := svc.CreateAppointment(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: org.ID,
		ServiceID:      serviceItem.ID,
		StaffID:        st2.ID,
		StartTime:      t2,
		GuestName:      "Client Two",
	})
	if err != nil {
		t.Fatalf("failed to create appt 2: %v", err)
	}

	// 4. Test Filtering by StaffID
	list1, _, err := svc.ListAppointments(ctx, appointment.ListFilter{
		OrganizationID: org.ID,
		StaffID:        &st1.ID,
	}, 1, 10)
	if err != nil {
		t.Fatalf("failed to filter appointments by staff: %v", err)
	}
	if len(list1) != 1 || list1[0].ID != appt1.ID {
		t.Errorf("expected 1 appointment for Alice, got %d", len(list1))
	}

	// 5. Test Filtering by Date Range
	list2, _, err := svc.ListAppointments(ctx, appointment.ListFilter{
		OrganizationID: org.ID,
		DateFrom:       &t2,
		DateTo:         &t2,
	}, 1, 10)
	if err != nil {
		t.Fatalf("failed to filter appointments by date: %v", err)
	}
	if len(list2) != 1 || list2[0].ID != appt2.ID {
		t.Errorf("expected 1 appointment for target date, got %d", len(list2))
	}
}

func TestCalendar_StatusTransitionAndRescheduleWorkflow(t *testing.T) {
	orgRepo := memory.NewOrgRepository()
	locRepo := memory.NewLocationRepository()
	svcRepo := memory.NewServiceRepository()
	staffRepo := memory.NewStaffRepository()
	schedRepo := memory.NewSchedulingRepository()
	resRepo := memory.NewResourceRepository()
	apptRepo := memory.NewAppointmentRepository()

	engine := availabilityuc.NewEngine(orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, apptRepo)
	svc := bookinguc.NewService(apptRepo, orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, engine)

	org, _ := orgRepo.Create(context.Background(), &organization.Organization{
		ID:       uuid.New(),
		Name:     "Status Test Salon",
		Timezone: "UTC",
	})

	ctx := rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   org.ID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	})

	serviceItem, _ := svcRepo.CreateService(ctx, service.CreateServiceCmd{
		OrganizationID:  org.ID,
		Name:            "Facial Spa",
		DurationMinutes: 30,
		PriceCents:      5000,
	})

	st, _ := staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Clara",
		LastName:       "Esthetician",
	})

	t1 := time.Now().UTC().Add(12 * time.Hour).Truncate(time.Hour)
	appt, err := svc.CreateAppointment(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: org.ID,
		ServiceID:      serviceItem.ID,
		StaffID:        st.ID,
		StartTime:      t1,
		GuestName:      "Status Client",
	})
	if err != nil {
		t.Fatalf("failed to create appointment: %v", err)
	}

	// Confirm Appointment
	err = svc.UpdateAppointmentStatus(ctx, appointment.UpdateStatusCmd{
		AppointmentID: appt.ID,
		NewStatus:     appointment.StatusConfirmed,
		Reason:        "Staff verified deposit",
	})
	if err != nil {
		t.Fatalf("failed to confirm appointment: %v", err)
	}

	// Reschedule to 3 hours later
	newTime := t1.Add(3 * time.Hour)
	rescheduled, err := svc.RescheduleAppointment(ctx, appointment.RescheduleCmd{
		AppointmentID: appt.ID,
		NewStartTime:  newTime,
		Reason:        "Client request",
	})
	if err != nil {
		t.Fatalf("failed to reschedule appointment: %v", err)
	}
	if rescheduled.Status != appointment.StatusRescheduled {
		t.Errorf("expected status rescheduled, got %s", rescheduled.Status)
	}

	// Mark Completed
	err = svc.UpdateAppointmentStatus(ctx, appointment.UpdateStatusCmd{
		AppointmentID: appt.ID,
		NewStatus:     appointment.StatusCompleted,
		Reason:        "Service finished successfully",
	})
	if err != nil {
		t.Fatalf("failed to complete appointment: %v", err)
	}

	fetched, _ := svc.GetAppointmentByID(ctx, org.ID, appt.ID)
	if fetched.Status != appointment.StatusCompleted {
		t.Errorf("expected status completed, got %s", fetched.Status)
	}
}
