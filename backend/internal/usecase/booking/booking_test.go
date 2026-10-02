package booking_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
	bookinguc "github.com/appointly/appointly/backend/internal/usecase/booking"
)

func setupTestBookingService(t *testing.T) (
	*bookinguc.Service,
	uuid.UUID, // orgID
	uuid.UUID, // serviceID
	uuid.UUID, // staffID
	uuid.UUID, // locationID
	*memory.AppointmentRepository,
	*memory.ServiceRepository,
	*memory.StaffRepository,
) {
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
		Name:         "Barber Shop VIP",
		Slug:         "barber-vip",
		BusinessType: organization.BusinessTypeSalon,
		Email:        "vip@barber.com",
		Timezone:     "UTC",
		OwnerID:      uuid.New(),
		Settings: organization.Settings{
			AutoConfirmBookings:    true,
			MinAdvanceBookingHours: 0,
			MaxAdvanceDays:         30,
			CancellationNoticehours: 24,
		},
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	// 2. Create Service
	svcItem, err := svcRepo.CreateService(ctx, service.CreateServiceCmd{
		OrganizationID:  org.ID,
		Name:            "Gentleman Haircut",
		DurationMinutes: 45,
		PriceCents:      3500,
		Currency:        "USD",
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// 3. Create Staff
	st, err := staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "Alex",
		LastName:       "Master",
		Email:          "alex@barber.com",
	})
	if err != nil {
		t.Fatalf("failed to create staff: %v", err)
	}
	_ = staffRepo.SetStaffServices(ctx, org.ID, st.ID, []uuid.UUID{svcItem.ID})

	// 4. Setup Schedule
	locID := uuid.New()
	targetDate := time.Now().UTC().AddDate(0, 0, 2)
	dayOfWeek := scheduling.WeekdayToDayOfWeek(targetDate.Weekday())

	_ = schedRepo.UpsertStaffSchedule(ctx, scheduling.UpsertStaffScheduleCmd{
		OrganizationID: org.ID,
		StaffID:        st.ID,
		Schedule: []scheduling.StaffSchedule{
			{
				DayOfWeek: dayOfWeek,
				IsWorking: true,
				StartTime: "08:00",
				EndTime:   "18:00",
			},
		},
	})

	return svc, org.ID, svcItem.ID, st.ID, locID, apptRepo, svcRepo, staffRepo
}

func contextWithAuth(orgID, userID uuid.UUID, role rbac.Role) context.Context {
	return rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  userID,
		OrgID:   orgID,
		Role:    role,
		IsOwner: role == rbac.RoleOwner,
	})
}

func TestBooking_NormalFlow(t *testing.T) {
	svc, orgID, serviceID, staffID, locationID, _, _, _ := setupTestBookingService(t)
	futureTime := time.Now().UTC().AddDate(0, 0, 2).Truncate(time.Hour).Add(10 * time.Hour)

	cmd := appointment.CreateAppointmentCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		ServiceID:      serviceID,
		StaffID:        staffID,
		StartTime:      futureTime,
		GuestName:      "Michael Corleone",
		GuestEmail:     "michael@godfather.com",
		GuestPhone:     "+1555111222",
		Source:         "online",
		CreatedBy:      uuid.New(),
	}

	appt, err := svc.CreateAppointment(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error creating appointment, got %v", err)
	}

	if appt.GuestName != "Michael Corleone" {
		t.Errorf("expected guest name Michael Corleone, got %s", appt.GuestName)
	}
	if appt.Status != appointment.StatusConfirmed {
		t.Errorf("expected status confirmed by auto-confirm setting, got %s", appt.Status)
	}
}

func TestBooking_DoubleBookingRacePrevention(t *testing.T) {
	svc, orgID, serviceID, staffID, locationID, _, _, _ := setupTestBookingService(t)
	futureTime := time.Now().UTC().AddDate(0, 0, 2).Truncate(time.Hour).Add(14 * time.Hour)

	var wg sync.WaitGroup
	concurrentAttempts := 10
	successCount := 0
	conflictCount := 0
	var mu sync.Mutex

	for i := 0; i < concurrentAttempts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmd := appointment.CreateAppointmentCmd{
				OrganizationID: orgID,
				LocationID:     locationID,
				ServiceID:      serviceID,
				StaffID:        staffID,
				StartTime:      futureTime,
				GuestName:      "Concurrent Client",
				GuestEmail:     "concurrent@test.com",
				Source:         "online",
				CreatedBy:      uuid.New(),
			}

			_, err := svc.CreateAppointment(context.Background(), cmd)
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				conflictCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful booking under race conditions, got %d", successCount)
	}
	if conflictCount != concurrentAttempts-1 {
		t.Errorf("expected %d conflict rejections, got %d", concurrentAttempts-1, conflictCount)
	}
}

func TestBooking_InactiveServiceFailure(t *testing.T) {
	svc, orgID, serviceID, staffID, locationID, _, svcRepo, _ := setupTestBookingService(t)
	ctx := context.Background()

	stStatus := service.StatusInactive
	_, _ = svcRepo.UpdateService(ctx, service.UpdateServiceCmd{
		ID:             serviceID,
		OrganizationID: orgID,
		Status:         &stStatus,
	})

	futureTime := time.Now().UTC().AddDate(0, 0, 2).Truncate(time.Hour).Add(11 * time.Hour)
	_, err := svc.CreateAppointment(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		ServiceID:      serviceID,
		StaffID:        staffID,
		StartTime:      futureTime,
	})

	if err == nil {
		t.Error("expected error booking inactive service, got nil")
	}
}

func TestBooking_RescheduleAndCancelWorkflow(t *testing.T) {
	svc, orgID, serviceID, staffID, locationID, _, _, _ := setupTestBookingService(t)
	ctx := contextWithAuth(orgID, uuid.New(), rbac.RoleOwner)

	futureTime := time.Now().UTC().AddDate(0, 0, 2).Truncate(time.Hour).Add(10 * time.Hour)
	appt, err := svc.CreateAppointment(ctx, appointment.CreateAppointmentCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		ServiceID:      serviceID,
		StaffID:        staffID,
		StartTime:      futureTime,
		GuestName:      "Workflow Test",
	})
	if err != nil {
		t.Fatalf("failed to create appointment: %v", err)
	}

	// Reschedule to 15:00
	newTime := time.Date(futureTime.Year(), futureDateMonth(futureTime), futureTime.Day(), 15, 0, 0, 0, time.UTC)
	rescheduled, err := svc.RescheduleAppointment(ctx, appointment.RescheduleCmd{
		AppointmentID: appt.ID,
		NewStartTime:  newTime,
		Reason:        "Customer conflict",
		ChangedBy:     uuid.New(),
	})
	if err != nil {
		t.Fatalf("failed to reschedule appointment: %v", err)
	}
	if rescheduled.Status != appointment.StatusRescheduled {
		t.Errorf("expected status rescheduled, got %s", rescheduled.Status)
	}

	// Cancel appointment
	err = svc.CancelAppointment(ctx, orgID, appt.ID, "Family emergency", uuid.New())
	if err != nil {
		t.Fatalf("failed to cancel appointment: %v", err)
	}

	updated, err := svc.GetAppointmentByID(ctx, orgID, appt.ID)
	if err != nil {
		t.Fatalf("failed to get updated appointment: %v", err)
	}
	if updated.Status != appointment.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", updated.Status)
	}
}

func futureDateMonth(t time.Time) time.Month {
	return t.Month()
}
