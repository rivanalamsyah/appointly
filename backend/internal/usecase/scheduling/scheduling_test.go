package scheduling_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/scheduling"
)

func setupSchedulingService() (*usecase.Service, *memory.SchedulingRepository) {
	schedRepo := memory.NewSchedulingRepository()
	auditRepo := memory.NewAuditRepository()
	svc := usecase.NewService(schedRepo, auditRepo)
	return svc, schedRepo
}

func TestBusinessHours_Validation(t *testing.T) {
	svc, _ := setupSchedulingService()
	ctx := context.Background()

	orgID := uuid.New()
	locationID := uuid.New()
	authCtx := &rbac.AuthContext{
		UserID:  uuid.New(),
		OrgID:   orgID,
		Role:    rbac.RoleOwner,
		IsOwner: true,
	}

	// Test 1: Invalid time string (not HH:MM)
	err := svc.UpsertBusinessHours(ctx, authCtx, scheduling.UpsertBusinessHoursCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		Hours: []scheduling.BusinessHours{
			{
				DayOfWeek: scheduling.Monday,
				IsOpen:    true,
				OpenTime:  "25:00", // Invalid!
				CloseTime: "18:00",
			},
		},
	})
	if err == nil {
		t.Errorf("EXPECTED VALIDATION ERROR for invalid HH:MM time string, got nil")
	}

	// Test 2: Close time before open time
	err = svc.UpsertBusinessHours(ctx, authCtx, scheduling.UpsertBusinessHoursCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		Hours: []scheduling.BusinessHours{
			{
				DayOfWeek: scheduling.Monday,
				IsOpen:    true,
				OpenTime:  "18:00", // Close before Open!
				CloseTime: "09:00",
			},
		},
	})
	if err == nil {
		t.Errorf("EXPECTED VALIDATION ERROR for close_time before open_time, got nil")
	}

	// Test 3: Valid Business Hours
	err = svc.UpsertBusinessHours(ctx, authCtx, scheduling.UpsertBusinessHoursCmd{
		OrganizationID: orgID,
		LocationID:     locationID,
		Hours: []scheduling.BusinessHours{
			{
				DayOfWeek: scheduling.Monday,
				IsOpen:    true,
				OpenTime:  "09:00",
				CloseTime: "18:00",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected valid business hours to succeed, got %v", err)
	}
}

func TestStaffSchedule_OverlappingIntervalValidation(t *testing.T) {
	svc, _ := setupSchedulingService()
	ctx := context.Background()

	orgID := uuid.New()
	staffID := uuid.New()
	authCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: orgID, Role: rbac.RoleOwner}

	bStart := "08:00" // Break starts before working hours (09:00)!
	bEnd := "12:00"

	err := svc.UpsertStaffSchedule(ctx, authCtx, scheduling.UpsertStaffScheduleCmd{
		OrganizationID: orgID,
		StaffID:        staffID,
		Schedule: []scheduling.StaffSchedule{
			{
				DayOfWeek:  scheduling.Monday,
				IsWorking:  true,
				StartTime:  "09:00",
				EndTime:    "17:00",
				BreakStart: &bStart,
				BreakEnd:   &bEnd,
			},
		},
	})
	if err == nil {
		t.Fatalf("EXPECTED BREAK INTERVAL OVERLAP ERROR when break is outside working hours, got nil")
	}
}

func TestStaffTimeOff_CollisionAndDateBoundary(t *testing.T) {
	svc, _ := setupSchedulingService()
	ctx := context.Background()

	orgID := uuid.New()
	staffID := uuid.New()
	authCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: orgID, Role: rbac.RoleOwner}

	startDate := time.Now().AddDate(0, 0, 10)
	endDate := time.Now().AddDate(0, 0, 5) // End date before start date!

	_, err := svc.CreateStaffTimeOff(ctx, authCtx, scheduling.CreateTimeOffCmd{
		OrganizationID: orgID,
		StaffID:        staffID,
		StartDate:      startDate,
		EndDate:        endDate,
		Reason:         "Vacation",
		IsAllDay:       true,
	})
	if err == nil {
		t.Fatalf("EXPECTED DATE BOUNDARY ERROR when end_date < start_date, got nil")
	}
}

func TestScheduling_TenantIsolationAndRBAC(t *testing.T) {
	svc, _ := setupSchedulingService()
	ctx := context.Background()

	org1ID := uuid.New()
	org2ID := uuid.New()

	// User 1 in Org 1
	authCtx1 := &rbac.AuthContext{UserID: uuid.New(), OrgID: org1ID, Role: rbac.RoleOwner}

	// User 1 attempts to update business hours for Org 2
	err := svc.UpsertBusinessHours(ctx, authCtx1, scheduling.UpsertBusinessHoursCmd{
		OrganizationID: org2ID, // Cross-tenant!
		LocationID:     uuid.New(),
		Hours:          []scheduling.BusinessHours{},
	})
	if err == nil {
		t.Fatalf("EXPECTED CROSS-TENANT SECURITY ERROR, got nil")
	}

	// User with STAFF role attempts to manage business hours
	staffAuthCtx := &rbac.AuthContext{UserID: uuid.New(), OrgID: org1ID, Role: rbac.RoleStaff}
	err = svc.UpsertBusinessHours(ctx, staffAuthCtx, scheduling.UpsertBusinessHoursCmd{
		OrganizationID: org1ID,
		LocationID:     uuid.New(),
		Hours:          []scheduling.BusinessHours{},
	})
	if err == nil {
		t.Fatalf("EXPECTED PERMISSION DENIED error for STAFF role updating business hours, got nil")
	}
}
