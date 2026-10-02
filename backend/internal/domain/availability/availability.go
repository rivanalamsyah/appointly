// Package availability defines the domain concepts, time-window math,
// and slot generation structures for the Appointly Availability Engine.
package availability

import (
	"time"

	"github.com/google/uuid"
)

// TimeWindow represents a continuous interval [Start, End).
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration returns the window duration.
func (w TimeWindow) Duration() time.Duration {
	if w.End.Before(w.Start) {
		return 0
	}
	return w.End.Sub(w.Start)
}

// Overlaps returns true if two time windows overlap (open interval comparison).
func (w TimeWindow) Overlaps(other TimeWindow) bool {
	return w.Start.Before(other.End) && w.End.After(other.Start)
}

// Contains returns true if other is completely enclosed within w.
func (w TimeWindow) Contains(other TimeWindow) bool {
	return !w.Start.After(other.Start) && !w.End.Before(other.End)
}

// Intersect computes the overlapping time window between w and other.
// Returns (intersectionWindow, ok).
func (w TimeWindow) Intersect(other TimeWindow) (TimeWindow, bool) {
	start := w.Start
	if other.Start.After(start) {
		start = other.Start
	}

	end := w.End
	if other.End.Before(end) {
		end = other.End
	}

	if start.Before(end) {
		return TimeWindow{Start: start, End: end}, true
	}
	return TimeWindow{}, false
}

// Subtract removes the exclude window from w, returning 0, 1, or 2 remaining windows.
func (w TimeWindow) Subtract(exclude TimeWindow) []TimeWindow {
	intersection, ok := w.Intersect(exclude)
	if !ok {
		// No overlap — return original window
		return []TimeWindow{w}
	}

	var result []TimeWindow
	if w.Start.Before(intersection.Start) {
		result = append(result, TimeWindow{Start: w.Start, End: intersection.Start})
	}
	if w.End.After(intersection.End) {
		result = append(result, TimeWindow{Start: intersection.End, End: w.End})
	}
	return result
}

// Slot represents a valid candidate booking time slot returned by the Availability Engine.
type Slot struct {
	StartTime              time.Time   `json:"start_time"`
	EndTime                time.Time   `json:"end_time"`
	FormattedStartTime     string      `json:"formatted_start_time"`
	FormattedEndTime       string      `json:"formatted_end_time"`
	Timezone               string      `json:"timezone"`
	StaffID                uuid.UUID   `json:"staff_id"`
	StaffName              string      `json:"staff_name"`
	ResourceIDs            []uuid.UUID `json:"resource_ids,omitempty"`
	ServiceDurationMinutes int         `json:"service_duration_minutes"`
}

// GetAvailabilityQuery holds query parameters for computing candidate slots.
type GetAvailabilityQuery struct {
	OrganizationID uuid.UUID  `json:"organization_id"`
	LocationID     *uuid.UUID `json:"location_id,omitempty"`
	ServiceID      uuid.UUID  `json:"service_id"`
	StaffID        *uuid.UUID `json:"staff_id,omitempty"`
	DateFrom       time.Time  `json:"date_from"`
	DateTo         time.Time  `json:"date_to"`
	Timezone       string     `json:"timezone"` // IANA timezone e.g. "Asia/Jakarta"
}

// Engine defines the primary domain contract for the Availability Engine.
type Engine interface {
	GetAvailableSlots(query GetAvailabilityQuery) ([]Slot, error)
}
