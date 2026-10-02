package availability

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/availability"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Engine struct {
	orgRepo        organization.Repository
	locationRepo   location.Repository
	serviceRepo    service.Repository
	staffRepo      staff.Repository
	schedulingRepo scheduling.Repository
	resourceRepo   resource.Repository
	apptRepo       appointment.Repository
}

func NewEngine(
	orgRepo organization.Repository,
	locationRepo location.Repository,
	serviceRepo service.Repository,
	staffRepo staff.Repository,
	schedulingRepo scheduling.Repository,
	resourceRepo resource.Repository,
	apptRepo appointment.Repository,
) *Engine {
	return &Engine{
		orgRepo:        orgRepo,
		locationRepo:   locationRepo,
		serviceRepo:    serviceRepo,
		staffRepo:      staffRepo,
		schedulingRepo: schedulingRepo,
		resourceRepo:   resourceRepo,
		apptRepo:       apptRepo,
	}
}

func (e *Engine) GetAvailableSlots(ctx context.Context, query availability.GetAvailabilityQuery) ([]availability.Slot, error) {
	// 1. Load Service
	svc, err := e.serviceRepo.GetServiceByID(ctx, query.OrganizationID, query.ServiceID)
	if err != nil {
		return nil, apperror.NotFound("service")
	}
	if svc.Status != service.StatusActive {
		return []availability.Slot{}, nil // Inactive service returns 0 slots
	}

	// 2. Load Organization Settings & Timezone
	org, err := e.orgRepo.GetByID(ctx, query.OrganizationID)
	if err != nil {
		return nil, apperror.NotFound("organization")
	}

	tzName := query.Timezone
	if tzName == "" {
		tzName = org.Timezone
	}
	if tzName == "" {
		tzName = "UTC"
	}

	locTZ, err := time.LoadLocation(tzName)
	if err != nil {
		locTZ = time.UTC
	}

	granularityMinutes := org.Settings.SlotGranularityMinutes
	if granularityMinutes < 5 {
		granularityMinutes = 15 // default 15 minutes
	}

	minAdvanceNoticeHours := org.Settings.MinAdvanceBookingHours
	if minAdvanceNoticeHours < 0 {
		minAdvanceNoticeHours = 2
	}
	maxAdvanceDays := org.Settings.MaxAdvanceDays
	if maxAdvanceDays < 1 {
		maxAdvanceDays = 60
	}

	now := time.Now().UTC()
	earliestAllowed := now.Add(time.Duration(minAdvanceNoticeHours) * time.Hour)
	latestAllowed := now.AddDate(0, 0, maxAdvanceDays)

	// 3. Find Eligible Staff Candidates
	var eligibleStaff []*staff.Staff
	if query.StaffID != nil {
		st, err := e.staffRepo.GetByID(ctx, query.OrganizationID, *query.StaffID)
		if err != nil || !st.IsActive {
			return []availability.Slot{}, nil
		}
		eligibleStaff = []*staff.Staff{st}
	} else {
		allStaff, _, err := e.staffRepo.List(ctx, staff.ListStaffFilter{OrganizationID: query.OrganizationID}, 1, 100)
		if err != nil {
			return nil, err
		}
		for _, st := range allStaff {
			if st.IsActive {
				assignedSvcs, err := e.staffRepo.GetStaffServices(ctx, st.ID)
				if err == nil {
					for _, ss := range assignedSvcs {
						if ss.ServiceID == query.ServiceID {
							eligibleStaff = append(eligibleStaff, st)
							break
						}
					}
				}
			}
		}
	}

	if len(eligibleStaff) == 0 {
		return []availability.Slot{}, nil
	}

	serviceDuration := time.Duration(svc.DurationMinutes) * time.Minute
	bufferBefore := time.Duration(svc.BufferBefore) * time.Minute
	bufferAfter := time.Duration(svc.BufferAfter) * time.Minute
	totalSlotDuration := bufferBefore + serviceDuration + bufferAfter

	startDate := query.DateFrom.In(locTZ)
	endDate := query.DateTo.In(locTZ)

	var allSlots []availability.Slot

	for currDate := startDate; !currDate.After(endDate); currDate = currDate.AddDate(0, 0, 1) {
		dayOfWeek := scheduling.WeekdayToDayOfWeek(currDate.Weekday())

		// Check location operating hours for this day of week
		var locationWindow *availability.TimeWindow
		if query.LocationID != nil {
			locHours, err := e.schedulingRepo.GetBusinessHoursByLocation(ctx, query.OrganizationID, *query.LocationID)
			if err == nil {
				for _, bh := range locHours {
					if bh.DayOfWeek == dayOfWeek {
						if !bh.IsOpen {
							break
						}
						bizStart := parseTimeOnDate(currDate, bh.OpenTime, locTZ)
						bizEnd := parseTimeOnDate(currDate, bh.CloseTime, locTZ)
						if bizEnd.Before(bizStart) {
							bizEnd = bizEnd.AddDate(0, 0, 1) // Cross-midnight
						}
						locationWindow = &availability.TimeWindow{Start: bizStart.UTC(), End: bizEnd.UTC()}
						break
					}
				}
			}
		} else {
			// Default 24h or org-level window if no location specified
			bizStart := parseTimeOnDate(currDate, "08:00", locTZ)
			bizEnd := parseTimeOnDate(currDate, "20:00", locTZ)
			locationWindow = &availability.TimeWindow{Start: bizStart.UTC(), End: bizEnd.UTC()}
		}

		if locationWindow == nil {
			continue // Location is closed on this day
		}

		// Process each eligible staff member
		for _, st := range eligibleStaff {
			staffSchedules, err := e.schedulingRepo.GetStaffSchedule(ctx, query.OrganizationID, st.ID)
			if err != nil {
				continue
			}

			var staffWindows []availability.TimeWindow
			for _, sch := range staffSchedules {
				if sch.DayOfWeek == dayOfWeek && sch.IsWorking {
					sStart := parseTimeOnDate(currDate, sch.StartTime, locTZ).UTC()
					sEnd := parseTimeOnDate(currDate, sch.EndTime, locTZ).UTC()
					if sEnd.Before(sStart) {
						sEnd = sEnd.AddDate(0, 0, 1)
					}

					shiftWindow := availability.TimeWindow{Start: sStart, End: sEnd}
					if intersected, ok := locationWindow.Intersect(shiftWindow); ok {
						staffWindows = append(staffWindows, intersected)
					}

					// Subtract break if defined
					if sch.BreakStart != nil && sch.BreakEnd != nil {
						bStart := parseTimeOnDate(currDate, *sch.BreakStart, locTZ).UTC()
						bEnd := parseTimeOnDate(currDate, *sch.BreakEnd, locTZ).UTC()
						if bEnd.Before(bStart) {
							bEnd = bEnd.AddDate(0, 0, 1)
						}
						breakWindow := availability.TimeWindow{Start: bStart, End: bEnd}

						var nextWindows []availability.TimeWindow
						for _, w := range staffWindows {
							nextWindows = append(nextWindows, w.Subtract(breakWindow)...)
						}
						staffWindows = nextWindows
					}
				}
			}

			if len(staffWindows) == 0 {
				continue
			}

			// Subtract staff time-off
			timeOffs, err := e.schedulingRepo.ListTimeOffByStaff(ctx, query.OrganizationID, st.ID, currDate.AddDate(0, 0, -1).UTC(), currDate.AddDate(0, 0, 2).UTC())
			if err == nil {
				for _, to := range timeOffs {
					toWindow := availability.TimeWindow{Start: to.StartDate.UTC(), End: to.EndDate.UTC()}
					var nextWindows []availability.TimeWindow
					for _, w := range staffWindows {
						nextWindows = append(nextWindows, w.Subtract(toWindow)...)
					}
					staffWindows = nextWindows
				}
			}

			// Subtract existing appointments & buffers
			existingAppts, err := e.apptRepo.ListForStaff(ctx, query.OrganizationID, st.ID, currDate.AddDate(0, 0, -1).UTC(), currDate.AddDate(0, 0, 2).UTC())
			if err == nil {
				for _, appt := range existingAppts {
					if appt.Status == appointment.StatusCancelled {
						continue
					}
					conflictWindow := availability.TimeWindow{
						Start: appt.StartTime.Add(-bufferAfter),
						End:   appt.EndTime.Add(bufferBefore),
					}
					var nextWindows []availability.TimeWindow
					for _, w := range staffWindows {
						nextWindows = append(nextWindows, w.Subtract(conflictWindow)...)
					}
					staffWindows = nextWindows
				}
			}

			// Slot Generation within remaining staffWindows
			step := time.Duration(granularityMinutes) * time.Minute
			for _, w := range staffWindows {
				for t := w.Start; !t.Add(totalSlotDuration).After(w.End); t = t.Add(step) {
					slotStart := t
					slotEnd := t.Add(totalSlotDuration)

					if slotStart.Before(earliestAllowed) || slotStart.After(latestAllowed) {
						continue
					}

					var assignedResourceIDs []uuid.UUID
					if svc.RequiresResource {
						svcResources, resErr := e.resourceRepo.GetServiceResources(ctx, query.OrganizationID, query.ServiceID)
						if resErr != nil || len(svcResources) == 0 {
							continue
						}
						for _, r := range svcResources {
							if r.Status != resource.StatusActive || r.Capacity <= 0 {
								continue
							}
							if query.LocationID != nil && r.LocationID != nil && *r.LocationID != *query.LocationID {
								continue
							}
							rConflicts, _ := e.apptRepo.FindResourceConflicts(ctx, query.OrganizationID, r.ID, slotStart, slotEnd, nil)
							if len(rConflicts) == 0 {
								assignedResourceIDs = append(assignedResourceIDs, r.ID)
								break
							}
						}
						if len(assignedResourceIDs) == 0 {
							continue
						}
					}

					slotStartInTZ := slotStart.In(locTZ)
					slotEndInTZ := slotEnd.In(locTZ)

					allSlots = append(allSlots, availability.Slot{
						StartTime:              slotStartInTZ,
						EndTime:                slotEndInTZ,
						FormattedStartTime:     slotStartInTZ.Format(time.RFC3339),
						FormattedEndTime:       slotEndInTZ.Format(time.RFC3339),
						Timezone:               tzName,
						StaffID:                st.ID,
						StaffName:              st.FullName(),
						ResourceIDs:            assignedResourceIDs,
						ServiceDurationMinutes: svc.DurationMinutes,
					})
				}
			}
		}
	}

	sort.Slice(allSlots, func(i, j int) bool {
		return allSlots[i].StartTime.Before(allSlots[j].StartTime)
	})

	return allSlots, nil
}

func parseTimeOnDate(date time.Time, hhmm string, loc *time.Location) time.Time {
	var hour, min int
	fmt.Sscanf(hhmm, "%d:%d", &hour, &min)
	return time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, loc)
}
