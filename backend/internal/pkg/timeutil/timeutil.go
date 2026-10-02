// Package timeutil provides timezone-aware time utilities for the scheduling engine.
// All times in the database are stored in UTC. Conversion to/from tenant timezone
// happens at the API boundary.
package timeutil

import (
	"fmt"
	"time"
)

// WeekdayName maps Go's time.Weekday to lowercase string names
// used in the database (monday, tuesday, etc.)
var WeekdayName = map[time.Weekday]string{
	time.Sunday:    "sunday",
	time.Monday:    "monday",
	time.Tuesday:   "tuesday",
	time.Wednesday: "wednesday",
	time.Thursday:  "thursday",
	time.Friday:    "friday",
	time.Saturday:  "saturday",
}

// ParseLocation parses an IANA timezone string into a *time.Location.
// Returns an error if the timezone is invalid.
func ParseLocation(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("timeutil: invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

// MustParseLocation parses a timezone, panicking on error.
// Only use with known-valid timezone strings.
func MustParseLocation(tz string) *time.Location {
	loc, err := ParseLocation(tz)
	if err != nil {
		panic(err)
	}
	return loc
}

// StartOfDay returns the start of the day (00:00:00) in the given location.
func StartOfDay(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// EndOfDay returns the end of the day (23:59:59.999999999) in the given location.
func EndOfDay(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, loc)
}

// ToUTC converts a time to UTC.
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// InLocation converts a UTC time to the given location.
func InLocation(t time.Time, loc *time.Location) time.Time {
	return t.In(loc)
}

// DateOnly returns a date-only string in YYYY-MM-DD format.
func DateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseDate parses a YYYY-MM-DD date string in the given location.
func ParseDate(dateStr string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil: invalid date %q: %w", dateStr, err)
	}
	return t, nil
}

// ParseTimeOfDay parses an HH:MM time string into hours and minutes.
func ParseTimeOfDay(s string) (hour, minute int, err error) {
	var h, m int
	_, err = fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil {
		return 0, 0, fmt.Errorf("timeutil: invalid time-of-day %q: %w", s, err)
	}
	if h < 0 || h > 23 {
		return 0, 0, fmt.Errorf("timeutil: hour out of range: %d", h)
	}
	if m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("timeutil: minute out of range: %d", m)
	}
	return h, m, nil
}

// ApplyTimeOfDay applies HH:MM to a given date in the specified location.
func ApplyTimeOfDay(date time.Time, timeStr string, loc *time.Location) (time.Time, error) {
	h, m, err := ParseTimeOfDay(timeStr)
	if err != nil {
		return time.Time{}, err
	}
	date = date.In(loc)
	return time.Date(date.Year(), date.Month(), date.Day(), h, m, 0, 0, loc), nil
}

// SlotsInRange generates time slots between start and end with the given duration step.
// Each slot represents the START time of the slot.
func SlotsInRange(start, end time.Time, step time.Duration) []time.Time {
	if step <= 0 || start.After(end) {
		return nil
	}

	var slots []time.Time
	current := start
	for !current.After(end.Add(-step)) {
		slots = append(slots, current)
		current = current.Add(step)
	}
	return slots
}

// Overlaps returns true if the interval [start1, end1) overlaps with [start2, end2).
func Overlaps(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// TruncateToMinute truncates a time to the minute boundary.
func TruncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

// AddMinutes adds minutes to a time.
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}
