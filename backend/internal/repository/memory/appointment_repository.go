package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type AppointmentRepository struct {
	mu           sync.RWMutex
	appointments map[uuid.UUID]*appointment.Appointment
	history      map[uuid.UUID][]*appointment.StatusHistory
}

func NewAppointmentRepository() *AppointmentRepository {
	return &AppointmentRepository{
		appointments: make(map[uuid.UUID]*appointment.Appointment),
		history:      make(map[uuid.UUID][]*appointment.StatusHistory),
	}
}

func (r *AppointmentRepository) Create(ctx context.Context, cmd appointment.CreateAppointmentCmd) (*appointment.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check conflicts for staff
	for _, a := range r.appointments {
		if a.OrganizationID == cmd.OrganizationID && a.StaffID == cmd.StaffID && a.Status != appointment.StatusCancelled {
			if a.StartTime.Before(cmd.StartTime.Add(time.Minute*30)) && cmd.StartTime.Before(a.EndTime) { // overlap check
				return nil, apperror.Conflict("time slot already booked for staff")
			}
		}
	}

	now := time.Now()
	apptID := uuid.New()

	endTime := cmd.EndTime.UTC()
	if endTime.IsZero() {
		endTime = cmd.StartTime.Add(30 * time.Minute).UTC()
	}

	appt := &appointment.Appointment{
		ID:             apptID,
		OrganizationID: cmd.OrganizationID,
		LocationID:     cmd.LocationID,
		ServiceID:      cmd.ServiceID,
		StaffID:        cmd.StaffID,
		CustomerID:     cmd.CustomerID,
		ResourceID:     cmd.ResourceID,
		StartTime:      cmd.StartTime.UTC(),
		EndTime:        endTime,
		Timezone:       cmd.Timezone,
		Status:         appointment.StatusPending,
		PriceCents:     cmd.PriceCents,
		Currency:       cmd.Currency,
		GuestName:      strings.TrimSpace(cmd.GuestName),
		GuestEmail:     strings.TrimSpace(cmd.GuestEmail),
		GuestPhone:     strings.TrimSpace(cmd.GuestPhone),
		Notes:          strings.TrimSpace(cmd.Notes),
		Source:         cmd.Source,
		CreatedBy:      cmd.CreatedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.appointments[apptID] = appt

	// Record initial status history
	hist := &appointment.StatusHistory{
		ID:            uuid.New(),
		AppointmentID: apptID,
		FromStatus:    "",
		ToStatus:      appointment.StatusPending,
		Reason:        "Appointment created",
		ChangedBy:     cmd.CreatedBy,
		ChangedAt:     now,
	}
	r.history[apptID] = append(r.history[apptID], hist)

	cp := *appt
	return &cp, nil
}

func (r *AppointmentRepository) GetByID(ctx context.Context, orgID, appointmentID uuid.UUID) (*appointment.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	appt, exists := r.appointments[appointmentID]
	if !exists || appt.OrganizationID != orgID {
		return nil, apperror.NotFound("appointment")
	}

	cp := *appt
	return &cp, nil
}

func (r *AppointmentRepository) List(ctx context.Context, filter appointment.ListFilter, page, perPage int) ([]*appointment.Appointment, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var matched []*appointment.Appointment
	for _, a := range r.appointments {
		if a.OrganizationID != filter.OrganizationID {
			continue
		}
		if filter.LocationID != nil && a.LocationID != *filter.LocationID {
			continue
		}
		if filter.StaffID != nil && a.StaffID != *filter.StaffID {
			continue
		}
		if filter.CustomerID != nil && (a.CustomerID == nil || *a.CustomerID != *filter.CustomerID) {
			continue
		}
		if filter.Status != nil && a.Status != *filter.Status {
			continue
		}
		if filter.ServiceID != nil && a.ServiceID != *filter.ServiceID {
			continue
		}
		if filter.DateFrom != nil && a.StartTime.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && a.EndTime.After(*filter.DateTo) {
			continue
		}

		cp := *a
		matched = append(matched, &cp)
	}

	total := len(matched)
	start := (page - 1) * perPage
	if start >= total {
		return []*appointment.Appointment{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *AppointmentRepository) ListForStaff(ctx context.Context, orgID, staffID uuid.UUID, from, to time.Time) ([]*appointment.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*appointment.Appointment
	for _, a := range r.appointments {
		if a.OrganizationID == orgID && a.StaffID == staffID && a.Status != appointment.StatusCancelled {
			if a.EndTime.After(from) && a.StartTime.Before(to) {
				cp := *a
				result = append(result, &cp)
			}
		}
	}
	return result, nil
}

func (r *AppointmentRepository) ListForCalendar(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]*appointment.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*appointment.Appointment
	for _, a := range r.appointments {
		if a.OrganizationID == orgID && a.Status != appointment.StatusCancelled {
			if a.EndTime.After(from) && a.StartTime.Before(to) {
				cp := *a
				result = append(result, &cp)
			}
		}
	}
	return result, nil
}

func (r *AppointmentRepository) UpdateStatus(ctx context.Context, cmd appointment.UpdateStatusCmd) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.appointments[cmd.AppointmentID]
	if !exists {
		return apperror.NotFound("appointment")
	}

	if !appointment.CanTransition(a.Status, cmd.NewStatus) {
		return apperror.ValidationFailed("invalid appointment status transition")
	}

	oldStatus := a.Status
	a.Status = cmd.NewStatus
	a.UpdatedAt = time.Now()

	hist := &appointment.StatusHistory{
		ID:            uuid.New(),
		AppointmentID: a.ID,
		FromStatus:    oldStatus,
		ToStatus:      cmd.NewStatus,
		Reason:        cmd.Reason,
		ChangedBy:     cmd.ChangedBy,
		ChangedAt:     time.Now(),
	}
	r.history[a.ID] = append(r.history[a.ID], hist)
	return nil
}

func (r *AppointmentRepository) Reschedule(ctx context.Context, cmd appointment.RescheduleCmd) (*appointment.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.appointments[cmd.AppointmentID]
	if !exists {
		return nil, apperror.NotFound("appointment")
	}

	dur := a.EndTime.Sub(a.StartTime)
	a.StartTime = cmd.NewStartTime.UTC()
	a.EndTime = cmd.NewStartTime.Add(dur).UTC()
	a.Status = appointment.StatusRescheduled
	a.UpdatedAt = time.Now()

	cp := *a
	return &cp, nil
}

func (r *AppointmentRepository) UpdateNotes(ctx context.Context, cmd appointment.UpdateNotesCmd) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.appointments[cmd.AppointmentID]
	if !exists {
		return apperror.NotFound("appointment")
	}

	if cmd.Notes != nil {
		a.Notes = *cmd.Notes
	}
	if cmd.InternalNotes != nil {
		a.InternalNotes = *cmd.InternalNotes
	}
	a.UpdatedAt = time.Now()
	return nil
}

func (r *AppointmentRepository) GetStatusHistory(ctx context.Context, appointmentID uuid.UUID) ([]*appointment.StatusHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hist, exists := r.history[appointmentID]
	if !exists {
		return []*appointment.StatusHistory{}, nil
	}

	var result []*appointment.StatusHistory
	for _, h := range hist {
		cp := *h
		result = append(result, &cp)
	}
	return result, nil
}

func (r *AppointmentRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID, from, to time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, a := range r.appointments {
		if a.OrganizationID == orgID {
			if a.StartTime.After(from) && a.StartTime.Before(to) {
				count++
			}
		}
	}
	return count, nil
}

func (r *AppointmentRepository) FindConflicts(ctx context.Context, orgID, staffID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) ([]*appointment.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var conflicts []*appointment.Appointment
	for _, a := range r.appointments {
		if a.OrganizationID != orgID || a.StaffID != staffID || a.Status == appointment.StatusCancelled {
			continue
		}
		if excludeID != nil && a.ID == *excludeID {
			continue
		}

		// Overlap condition: start < a.EndTime AND end > a.StartTime
		if start.Before(a.EndTime) && end.After(a.StartTime) {
			cp := *a
			conflicts = append(conflicts, &cp)
		}
	}
	return conflicts, nil
}

func (r *AppointmentRepository) FindResourceConflicts(ctx context.Context, orgID, resourceID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) ([]*appointment.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var conflicts []*appointment.Appointment
	for _, a := range r.appointments {
		if a.OrganizationID != orgID || a.ResourceID == nil || *a.ResourceID != resourceID || a.Status == appointment.StatusCancelled {
			continue
		}
		if excludeID != nil && a.ID == *excludeID {
			continue
		}

		if start.Before(a.EndTime) && end.After(a.StartTime) {
			cp := *a
			conflicts = append(conflicts, &cp)
		}
	}
	return conflicts, nil
}
