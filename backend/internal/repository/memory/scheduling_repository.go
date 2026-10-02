package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type SchedulingRepository struct {
	mu            sync.RWMutex
	businessHours map[uuid.UUID][]*scheduling.BusinessHours // locationID -> hours
	staffSchedule map[uuid.UUID][]*scheduling.StaffSchedule // staffID -> schedule
	staffTimeOff  map[uuid.UUID]*scheduling.StaffTimeOff    // timeOffID -> timeOff
}

func NewSchedulingRepository() *SchedulingRepository {
	return &SchedulingRepository{
		businessHours: make(map[uuid.UUID][]*scheduling.BusinessHours),
		staffSchedule: make(map[uuid.UUID][]*scheduling.StaffSchedule),
		staffTimeOff:  make(map[uuid.UUID]*scheduling.StaffTimeOff),
	}
}

func (r *SchedulingRepository) GetBusinessHoursByLocation(ctx context.Context, orgID, locationID uuid.UUID) ([]*scheduling.BusinessHours, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hours := r.businessHours[locationID]
	var res []*scheduling.BusinessHours
	for _, h := range hours {
		if h.OrganizationID == orgID {
			cp := *h
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *SchedulingRepository) UpsertBusinessHours(ctx context.Context, cmd scheduling.UpsertBusinessHoursCmd) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var updated []*scheduling.BusinessHours
	now := time.Now()
	for _, h := range cmd.Hours {
		if h.ID == uuid.Nil {
			h.ID = uuid.New()
		}
		h.OrganizationID = cmd.OrganizationID
		h.LocationID = cmd.LocationID
		h.CreatedAt = now
		h.UpdatedAt = now
		cp := h
		updated = append(updated, &cp)
	}
	r.businessHours[cmd.LocationID] = updated
	return nil
}

func (r *SchedulingRepository) GetStaffSchedule(ctx context.Context, orgID, staffID uuid.UUID) ([]*scheduling.StaffSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schedules := r.staffSchedule[staffID]
	var res []*scheduling.StaffSchedule
	for _, s := range schedules {
		if s.OrganizationID == orgID {
			cp := *s
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *SchedulingRepository) UpsertStaffSchedule(ctx context.Context, cmd scheduling.UpsertStaffScheduleCmd) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var updated []*scheduling.StaffSchedule
	now := time.Now()
	for _, s := range cmd.Schedule {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		s.OrganizationID = cmd.OrganizationID
		s.StaffID = cmd.StaffID
		s.CreatedAt = now
		s.UpdatedAt = now
		cp := s
		updated = append(updated, &cp)
	}
	r.staffSchedule[cmd.StaffID] = updated
	return nil
}

func (r *SchedulingRepository) CreateTimeOff(ctx context.Context, timeOff *scheduling.StaffTimeOff) (*scheduling.StaffTimeOff, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if timeOff.ID == uuid.Nil {
		timeOff.ID = uuid.New()
	}
	now := time.Now()
	timeOff.CreatedAt = now
	timeOff.UpdatedAt = now

	cp := *timeOff
	r.staffTimeOff[timeOff.ID] = &cp
	return &cp, nil
}

func (r *SchedulingRepository) GetTimeOffByID(ctx context.Context, id uuid.UUID) (*scheduling.StaffTimeOff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	to, exists := r.staffTimeOff[id]
	if !exists {
		return nil, apperror.NotFound("staff time off")
	}
	cp := *to
	return &cp, nil
}

func (r *SchedulingRepository) ListTimeOffByStaff(ctx context.Context, orgID, staffID uuid.UUID, from, to time.Time) ([]*scheduling.StaffTimeOff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*scheduling.StaffTimeOff
	for _, tOff := range r.staffTimeOff {
		if tOff.OrganizationID == orgID && tOff.StaffID == staffID {
			// Date range overlap check
			if !tOff.EndDate.Before(from) && !tOff.StartDate.After(to) {
				cp := *tOff
				res = append(res, &cp)
			}
		}
	}
	return res, nil
}

func (r *SchedulingRepository) DeleteTimeOff(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.staffTimeOff[id]; !exists {
		return apperror.NotFound("staff time off")
	}
	delete(r.staffTimeOff, id)
	return nil
}

func (r *SchedulingRepository) UpdateTimeOff(ctx context.Context, timeOff *scheduling.StaffTimeOff) (*scheduling.StaffTimeOff, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.staffTimeOff[timeOff.ID]
	if !exists {
		return nil, apperror.NotFound("staff time off")
	}

	timeOff.UpdatedAt = time.Now()
	timeOff.CreatedAt = existing.CreatedAt
	cp := *timeOff
	r.staffTimeOff[timeOff.ID] = &cp
	return &cp, nil
}
