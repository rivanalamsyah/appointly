package availability_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/availability"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
)

func TestTimeWindow_Math(t *testing.T) {
	now := time.Now()
	w1 := availability.TimeWindow{Start: now, End: now.Add(2 * time.Hour)}
	w2 := availability.TimeWindow{Start: now.Add(1 * time.Hour), End: now.Add(3 * time.Hour)}

	// Overlap check
	if !w1.Overlaps(w2) {
		t.Error("expected w1 and w2 to overlap, got false")
	}

	// Intersect check
	intersected, ok := w1.Intersect(w2)
	if !ok {
		t.Fatal("expected intersection, got false")
	}
	if !intersected.Start.Equal(w2.Start) || !intersected.End.Equal(w1.End) {
		t.Errorf("unexpected intersection window: %v", intersected)
	}

	// Subtract check (subtracting middle window splits into two)
	wMain := availability.TimeWindow{Start: now, End: now.Add(4 * time.Hour)}
	wMid := availability.TimeWindow{Start: now.Add(1 * time.Hour), End: now.Add(2 * time.Hour)}
	subtracted := wMain.Subtract(wMid)

	if len(subtracted) != 2 {
		t.Fatalf("expected 2 windows after subtraction, got %d", len(subtracted))
	}
	if !subtracted[0].Start.Equal(now) || !subtracted[0].End.Equal(wMid.Start) {
		t.Errorf("unexpected first subtracted window: %v", subtracted[0])
	}
	if !subtracted[1].Start.Equal(wMid.End) || !subtracted[1].End.Equal(wMain.End) {
		t.Errorf("unexpected second subtracted window: %v", subtracted[1])
	}
}

func setupEngineFixtures(t *testing.T, testDate time.Time) (
	*availabilityuc.Engine,
	uuid.UUID, // orgID
	uuid.UUID, // serviceID
	uuid.UUID, // staffID
	*memory.AppointmentRepository,
	*memory.SchedulingRepository,
	*memory.ResourceRepository,
) {
	orgRepo := memory.NewOrgRepository()
	locRepo := memory.NewLocationRepository()
	svcRepo := memory.NewServiceRepository()
	staffRepo := memory.NewStaffRepository()
	schedRepo := memory.NewSchedulingRepository()
	resRepo := memory.NewResourceRepository()
	apptRepo := memory.NewAppointmentRepository()

	engine := availabilityuc.NewEngine(orgRepo, locRepo, svcRepo, staffRepo, schedRepo, resRepo, apptRepo)

	ctx := context.Background()

	// 1. Create Org
	org, err := orgRepo.Create(ctx, &organization.Organization{
		ID:           uuid.New(),
		Name:         "Test Salon",
		Slug:         "test-salon",
		BusinessType: organization.BusinessTypeSalon,
		Email:        "info@testsalon.com",
		Timezone:     "UTC",
		OwnerID:      uuid.New(),
		Settings: organization.Settings{
			SlotGranularityMinutes: 30,
			MinAdvanceBookingHours: 0,
			MaxAdvanceDays:        30,
		},
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	// 2. Create Service (60 min)
	svc, err := svcRepo.CreateService(ctx, service.CreateServiceCmd{
		OrganizationID:  org.ID,
		Name:            "Haircut & Wash",
		DurationMinutes: 60,
		BufferBefore:    0,
		BufferAfter:     0,
		PriceCents:      5000,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// 3. Create Staff
	st, err := staffRepo.Create(ctx, staff.CreateStaffCmd{
		OrganizationID: org.ID,
		FirstName:      "John",
		LastName:       "Stylist",
		Email:          "john@testsalon.com",
	})
	if err != nil {
		t.Fatalf("failed to create staff: %v", err)
	}
	_ = staffRepo.SetStaffServices(ctx, org.ID, st.ID, []uuid.UUID{svc.ID})

	// 4. Setup Staff Schedule (09:00 to 17:00)
	dayOfWeek := scheduling.WeekdayToDayOfWeek(testDate.Weekday())

	_ = schedRepo.UpsertStaffSchedule(ctx, scheduling.UpsertStaffScheduleCmd{
		OrganizationID: org.ID,
		StaffID:        st.ID,
		Schedule: []scheduling.StaffSchedule{
			{
				DayOfWeek: dayOfWeek,
				IsWorking: true,
				StartTime: "09:00",
				EndTime:   "17:00",
			},
		},
	})

	return engine, org.ID, svc.ID, st.ID, apptRepo, schedRepo, resRepo
}

func TestAvailabilityEngine_BasicSlotGeneration(t *testing.T) {
	futureDate := time.Now().UTC().AddDate(0, 0, 2)
	engine, orgID, svcID, staffID, _, _, _ := setupEngineFixtures(t, futureDate)

	query := availability.GetAvailabilityQuery{
		OrganizationID: orgID,
		ServiceID:      svcID,
		StaffID:        &staffID,
		DateFrom:       futureDate,
		DateTo:         futureDate,
		Timezone:       "UTC",
	}

	slots, err := engine.GetAvailableSlots(context.Background(), query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(slots) == 0 {
		t.Fatalf("expected slots for operating hours 09:00-17:00, got 0")
	}

	// First slot should start at 09:00 and end at 10:00 (60 min service duration)
	first := slots[0]
	if first.StartTime.Hour() != 9 || first.EndTime.Hour() != 10 {
		t.Errorf("expected first slot 09:00-10:00, got %s - %s", first.FormattedStartTime, first.FormattedEndTime)
	}
}

func TestAvailabilityEngine_AppointmentCollision(t *testing.T) {
	futureDate := time.Now().UTC().AddDate(0, 0, 2)
	engine, orgID, svcID, staffID, apptRepo, _, _ := setupEngineFixtures(t, futureDate)

	// Book appointment at 10:00 - 11:00 UTC (60 min duration)
	bookTime := time.Date(futureDate.Year(), futureDate.Month(), futureDate.Day(), 10, 0, 0, 0, time.UTC)
	_, err := apptRepo.Create(context.Background(), appointment.CreateAppointmentCmd{
		OrganizationID: orgID,
		LocationID:     uuid.New(),
		ServiceID:      svcID,
		StaffID:        staffID,
		StartTime:      bookTime,
		EndTime:        bookTime.Add(60 * time.Minute),
		Timezone:       "UTC",
		GuestName:      "Test Client",
	})
	if err != nil {
		t.Fatalf("failed to create appointment: %v", err)
	}

	query := availability.GetAvailabilityQuery{
		OrganizationID: orgID,
		ServiceID:      svcID,
		StaffID:        &staffID,
		DateFrom:       futureDate,
		DateTo:         futureDate,
		Timezone:       "UTC",
	}

	slots, err := engine.GetAvailableSlots(context.Background(), query)
	if err != nil {
		t.Fatalf("failed to compute slots: %v", err)
	}

	// Verify no slot overlaps with 10:00 - 11:00
	for _, s := range slots {
		if s.StartTime.Hour() == 10 {
			t.Errorf("found slot starting at 10:00 despite existing appointment: %v", s)
		}
	}
}

func TestAvailabilityEngine_TimeOffCollision(t *testing.T) {
	futureDate := time.Now().UTC().AddDate(0, 0, 2)
	engine, orgID, svcID, staffID, _, schedRepo, _ := setupEngineFixtures(t, futureDate)

	// Approve staff time-off from 13:00 to 17:00
	timeOffStart := time.Date(futureDate.Year(), futureDate.Month(), futureDate.Day(), 13, 0, 0, 0, time.UTC)
	timeOffEnd := time.Date(futureDate.Year(), futureDate.Month(), futureDate.Day(), 17, 0, 0, 0, time.UTC)

	_, err := schedRepo.CreateTimeOff(context.Background(), &scheduling.StaffTimeOff{
		ID:             uuid.New(),
		OrganizationID: orgID,
		StaffID:        staffID,
		StartDate:      timeOffStart,
		EndDate:        timeOffEnd,
		Reason:         "Personal Leave",
	})
	if err != nil {
		t.Fatalf("failed to create time off: %v", err)
	}

	query := availability.GetAvailabilityQuery{
		OrganizationID: orgID,
		ServiceID:      svcID,
		StaffID:        &staffID,
		DateFrom:       futureDate,
		DateTo:         futureDate,
		Timezone:       "UTC",
	}

	slots, err := engine.GetAvailableSlots(context.Background(), query)
	if err != nil {
		t.Fatalf("failed to compute slots: %v", err)
	}

	for _, s := range slots {
		if s.StartTime.Hour() >= 13 {
			t.Errorf("found slot starting after 13:00 despite staff time-off: %v", s)
		}
	}
}

func TestAvailabilityEngine_ConcurrencySafety(t *testing.T) {
	futureDate := time.Now().UTC().AddDate(0, 0, 2)
	engine, orgID, svcID, staffID, _, _, _ := setupEngineFixtures(t, futureDate)

	query := availability.GetAvailabilityQuery{
		OrganizationID: orgID,
		ServiceID:      svcID,
		StaffID:        &staffID,
		DateFrom:       futureDate,
		DateTo:         futureDate,
		Timezone:       "UTC",
	}

	var wg sync.WaitGroup
	concurrentRequests := 25

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots, err := engine.GetAvailableSlots(context.Background(), query)
			if err != nil || len(slots) == 0 {
				t.Errorf("concurrent availability query failed: %v", err)
			}
		}()
	}

	wg.Wait()
}
