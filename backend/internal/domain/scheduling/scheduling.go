// Package scheduling defines the availability and scheduling domain.
// This is the core of the booking engine — computing available time slots
// based on business hours, staff schedules, time-off, existing appointments,
// and booking settings. This package is pure domain logic with no side effects.
package scheduling

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DayOfWeek represents a day of the week in lowercase.
type DayOfWeek string

const (
	Monday    DayOfWeek = "monday"
	Tuesday   DayOfWeek = "tuesday"
	Wednesday DayOfWeek = "wednesday"
	Thursday  DayOfWeek = "thursday"
	Friday    DayOfWeek = "friday"
	Saturday  DayOfWeek = "saturday"
	Sunday    DayOfWeek = "sunday"
)

// WeekdayToDayOfWeek converts time.Weekday to DayOfWeek.
func WeekdayToDayOfWeek(w time.Weekday) DayOfWeek {
	switch w {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	default:
		return Sunday
	}
}

// BusinessHours defines the operating hours for a specific day at a location.
type BusinessHours struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	LocationID     uuid.UUID `json:"location_id"` // scoped to location
	DayOfWeek      DayOfWeek `json:"day_of_week"`
	IsOpen         bool      `json:"is_open"`
	OpenTime       string    `json:"open_time"`  // "HH:MM" in organization timezone
	CloseTime      string    `json:"close_time"` // "HH:MM" in organization timezone
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StaffSchedule defines a staff member's working hours for a specific day of week.
// This is the recurring weekly schedule. Exceptions go in StaffTimeOff.
type StaffSchedule struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	StaffID        uuid.UUID `json:"staff_id"`
	DayOfWeek      DayOfWeek `json:"day_of_week"`
	IsWorking      bool      `json:"is_working"`
	StartTime      string    `json:"start_time"` // "HH:MM" in organization timezone
	EndTime        string    `json:"end_time"`   // "HH:MM" in organization timezone
	BreakStart     *string   `json:"break_start,omitempty"` // optional break start "HH:MM"
	BreakEnd       *string   `json:"break_end,omitempty"`   // optional break end "HH:MM"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StaffTimeOff represents a period when a staff member is unavailable.
// Overlaps with recurring schedule take priority (staff is unavailable).
type StaffTimeOff struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	StaffID        uuid.UUID  `json:"staff_id"`
	StartDate      time.Time  `json:"start_date"` // inclusive, UTC
	EndDate        time.Time  `json:"end_date"`   // inclusive, UTC
	Reason         string     `json:"reason,omitempty"`
	IsAllDay       bool       `json:"is_all_day"`
	StartTime      *string    `json:"start_time,omitempty"` // if not all day
	EndTime        *string    `json:"end_time,omitempty"`   // if not all day
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// --- Availability Engine -----------------------------------------------------

// AvailabilityRequest is the input to the availability engine.
type AvailabilityRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id"`
	LocationID     uuid.UUID  `json:"location_id"`
	ServiceID      uuid.UUID  `json:"service_id"`
	StaffID        *uuid.UUID `json:"staff_id,omitempty"` // nil = any available staff
	Date           time.Time  `json:"date"`               // requested date (UTC midnight)
	Timezone       string     `json:"timezone"`           // IANA timezone for display
	ExcludeApptID  *uuid.UUID `json:"-"`                  // exclude when rescheduling
}

// TimeSlot represents a potential appointment slot.
type TimeSlot struct {
	StartTime time.Time  `json:"start_time"` // UTC
	EndTime   time.Time  `json:"end_time"`   // UTC
	StaffID   uuid.UUID  `json:"staff_id"`
	Available bool       `json:"available"`
	// Reason for unavailability (only set when Available=false, for debugging)
	UnavailableReason string `json:"unavailable_reason,omitempty"`
}

// AvailabilityResult contains the computed slots for a given request.
type AvailabilityResult struct {
	Date      string      `json:"date"`      // YYYY-MM-DD in requested timezone
	Timezone  string      `json:"timezone"`
	ServiceID uuid.UUID   `json:"service_id"`
	Slots     []TimeSlot  `json:"slots"`
}

// ServiceDuration holds the timing properties of a service needed for scheduling.
type ServiceDuration struct {
	DurationMinutes  int
	BufferBefore     int // minutes before the appointment (setup/prep)
	BufferAfter      int // minutes after the appointment (cleanup/turnover)
}

// TotalDuration returns the full blocked time including buffers.
func (sd ServiceDuration) TotalDuration() time.Duration {
	total := sd.DurationMinutes + sd.BufferBefore + sd.BufferAfter
	return time.Duration(total) * time.Minute
}

// SlotDuration returns only the visible appointment duration.
func (sd ServiceDuration) SlotDuration() time.Duration {
	return time.Duration(sd.DurationMinutes) * time.Minute
}

// --- Commands ----------------------------------------------------------------

type UpsertBusinessHoursCmd struct {
	OrganizationID uuid.UUID
	LocationID     uuid.UUID
	Hours          []BusinessHours
}

type UpsertStaffScheduleCmd struct {
	OrganizationID uuid.UUID
	StaffID        uuid.UUID
	Schedule       []StaffSchedule
}

type CreateTimeOffCmd struct {
	OrganizationID uuid.UUID
	StaffID        uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	Reason         string
	IsAllDay       bool
	StartTime      *string
	EndTime        *string
}

// --- Repository Interface ----------------------------------------------------

type Repository interface {
	// BusinessHours
	GetBusinessHoursByLocation(ctx context.Context, orgID, locationID uuid.UUID) ([]*BusinessHours, error)
	UpsertBusinessHours(ctx context.Context, cmd UpsertBusinessHoursCmd) error

	// StaffSchedule
	GetStaffSchedule(ctx context.Context, orgID, staffID uuid.UUID) ([]*StaffSchedule, error)
	UpsertStaffSchedule(ctx context.Context, cmd UpsertStaffScheduleCmd) error

	// StaffTimeOff
	CreateTimeOff(ctx context.Context, timeOff *StaffTimeOff) (*StaffTimeOff, error)
	GetTimeOffByID(ctx context.Context, id uuid.UUID) (*StaffTimeOff, error)
	ListTimeOffByStaff(ctx context.Context, orgID, staffID uuid.UUID, from, to time.Time) ([]*StaffTimeOff, error)
	DeleteTimeOff(ctx context.Context, id uuid.UUID) error
	UpdateTimeOff(ctx context.Context, timeOff *StaffTimeOff) (*StaffTimeOff, error)
}

// --- Availability Service Interface ------------------------------------------

// AvailabilityService defines the contract for computing available slots.
// The implementation orchestrates data from multiple repositories.
type AvailabilityService interface {
	// GetAvailableSlots computes available time slots for the given request.
	// This is the core availability engine — pure computation, no side effects.
	GetAvailableSlots(ctx context.Context, req AvailabilityRequest) (*AvailabilityResult, error)

	// IsSlotAvailable checks if a specific slot is still available.
	// Used for last-mile validation before creating an appointment.
	IsSlotAvailable(ctx context.Context, req AvailabilityRequest, startTime time.Time) (bool, error)
}
