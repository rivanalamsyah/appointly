package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/appointment"
	"github.com/appointly/appointly/backend/internal/domain/location"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/resource"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/domain/service"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	availabilityuc "github.com/appointly/appointly/backend/internal/usecase/availability"
)

type Service struct {
	apptRepo           appointment.Repository
	orgRepo            organization.Repository
	locationRepo       location.Repository
	serviceRepo        service.Repository
	staffRepo          staff.Repository
	schedulingRepo     scheduling.Repository
	resourceRepo       resource.Repository
	availabilityEngine *availabilityuc.Engine
}

func NewService(
	apptRepo appointment.Repository,
	orgRepo organization.Repository,
	locationRepo location.Repository,
	serviceRepo service.Repository,
	staffRepo staff.Repository,
	schedulingRepo scheduling.Repository,
	resourceRepo resource.Repository,
	availabilityEngine *availabilityuc.Engine,
) *Service {
	return &Service{
		apptRepo:           apptRepo,
		orgRepo:            orgRepo,
		locationRepo:       locationRepo,
		serviceRepo:        serviceRepo,
		staffRepo:          staffRepo,
		schedulingRepo:     schedulingRepo,
		resourceRepo:       resourceRepo,
		availabilityEngine: availabilityEngine,
	}
}

func (s *Service) CreateAppointment(ctx context.Context, cmd appointment.CreateAppointmentCmd) (*appointment.Appointment, error) {
	// 1. Verify Organization
	org, err := s.orgRepo.GetByID(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, apperror.NotFound("organization")
	}

	// 2. Verify Service status & duration
	svc, err := s.serviceRepo.GetServiceByID(ctx, cmd.OrganizationID, cmd.ServiceID)
	if err != nil {
		return nil, apperror.NotFound("service")
	}
	if svc.Status != service.StatusActive {
		return nil, apperror.ValidationFailed("selected service is inactive")
	}

	// 3. Verify Staff status & assignment
	st, err := s.staffRepo.GetByID(ctx, cmd.OrganizationID, cmd.StaffID)
	if err != nil {
		return nil, apperror.NotFound("staff member")
	}
	if !st.IsActive {
		return nil, apperror.ValidationFailed("selected staff member is inactive")
	}

	// 4. Calculate timing & snapshots
	serviceDuration := time.Duration(svc.DurationMinutes) * time.Minute
	if cmd.EndTime.IsZero() {
		cmd.EndTime = cmd.StartTime.Add(serviceDuration)
	}

	// Snapshot historical price and currency
	if cmd.PriceCents == 0 {
		cmd.PriceCents = svc.PriceCents
	}
	if cmd.Currency == "" {
		cmd.Currency = svc.Currency
	}
	if cmd.Source == "" {
		cmd.Source = "online"
	}

	// 5. Final Atomic Availability Validation at Transaction Time
	// Check staff conflict
	conflicts, err := s.apptRepo.FindConflicts(ctx, cmd.OrganizationID, cmd.StaffID, cmd.StartTime, cmd.EndTime, nil)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, apperror.Conflict("Selected time slot is no longer available. Please select another slot.")
	}

	// Check resource conflict if required
	if svc.RequiresResource {
		if cmd.ResourceID == nil {
			// Find available resource
			availRes, resErr := s.resourceRepo.GetServiceResources(ctx, cmd.OrganizationID, cmd.ServiceID)
			if resErr != nil || len(availRes) == 0 {
				return nil, apperror.ValidationFailed("No resource assigned to this service")
			}
			for _, r := range availRes {
				if r.Status == resource.StatusActive && r.Capacity > 0 {
					rConflicts, _ := s.apptRepo.FindResourceConflicts(ctx, cmd.OrganizationID, r.ID, cmd.StartTime, cmd.EndTime, nil)
					if len(rConflicts) == 0 {
						cmd.ResourceID = &r.ID
						break
					}
				}
			}
			if cmd.ResourceID == nil {
				return nil, apperror.Conflict("Required physical resource is busy for this time slot")
			}
		} else {
			rConflicts, _ := s.apptRepo.FindResourceConflicts(ctx, cmd.OrganizationID, *cmd.ResourceID, cmd.StartTime, cmd.EndTime, nil)
			if len(rConflicts) > 0 {
				return nil, apperror.Conflict("Selected resource is busy for this time slot")
			}
		}
	}

	// 6. Persist appointment within atomic transaction
	appt, err := s.apptRepo.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Auto-confirm if org settings permit
	if org.Settings.AutoConfirmBookings {
		_ = s.apptRepo.UpdateStatus(ctx, appointment.UpdateStatusCmd{
			AppointmentID: appt.ID,
			NewStatus:     appointment.StatusConfirmed,
			Reason:        "Auto-confirmed by system policy",
			ChangedBy:     cmd.CreatedBy,
		})
		appt.Status = appointment.StatusConfirmed
	}

	return appt, nil
}

func (s *Service) GetAppointmentByID(ctx context.Context, orgID, apptID uuid.UUID) (*appointment.Appointment, error) {
	return s.apptRepo.GetByID(ctx, orgID, apptID)
}

func (s *Service) ListAppointments(ctx context.Context, filter appointment.ListFilter, page, perPage int) ([]*appointment.Appointment, int, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil {
		filter.OrganizationID = authCtx.OrgID
	}
	return s.apptRepo.List(ctx, filter, page, perPage)
}

func (s *Service) RescheduleAppointment(ctx context.Context, cmd appointment.RescheduleCmd) (*appointment.Appointment, error) {
	authCtx := rbac.FromContext(ctx)
	orgID := uuid.Nil
	if authCtx != nil {
		orgID = authCtx.OrgID
		if !authCtx.Can(rbac.PermAppointmentUpdate) {
			return nil, apperror.Unauthorized("insufficient permissions to reschedule appointment")
		}
	}

	existing, err := s.apptRepo.GetByID(ctx, orgID, cmd.AppointmentID)
	if err != nil {
		return nil, apperror.NotFound("appointment")
	}

	if !appointment.CanTransition(existing.Status, appointment.StatusRescheduled) {
		return nil, apperror.ValidationFailed("cannot reschedule appointment in current status " + string(existing.Status))
	}

	// Calculate new end time based on original duration
	duration := existing.EndTime.Sub(existing.StartTime)
	newEndTime := cmd.NewStartTime.Add(duration)

	// Final availability validation for new slot
	conflicts, err := s.apptRepo.FindConflicts(ctx, existing.OrganizationID, existing.StaffID, cmd.NewStartTime, newEndTime, &existing.ID)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, apperror.Conflict("Target reschedule slot is already booked")
	}

	// Update appointment time & status history
	rescheduled, err := s.apptRepo.Reschedule(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Record status history audit
	_ = s.apptRepo.UpdateStatus(ctx, appointment.UpdateStatusCmd{
		AppointmentID: existing.ID,
		NewStatus:     appointment.StatusRescheduled,
		Reason:        cmd.Reason,
		ChangedBy:     cmd.ChangedBy,
	})

	return rescheduled, nil
}

func (s *Service) CancelAppointment(ctx context.Context, orgID, apptID uuid.UUID, reason string, cancelledBy uuid.UUID) error {
	existing, err := s.apptRepo.GetByID(ctx, orgID, apptID)
	if err != nil {
		return apperror.NotFound("appointment")
	}

	if !appointment.CanTransition(existing.Status, appointment.StatusCancelled) {
		return apperror.ValidationFailed("cannot cancel appointment in current status " + string(existing.Status))
	}

	// Policy notice check
	org, orgErr := s.orgRepo.GetByID(ctx, orgID)
	if orgErr == nil && org.Settings.CancellationNoticehours > 0 {
		noticeDuration := time.Duration(org.Settings.CancellationNoticehours) * time.Hour
		if time.Now().UTC().Add(noticeDuration).After(existing.StartTime) {
			// Cancellation within restricted notice window — allowed but recorded with policy note
			reason = "[Late Cancellation Window] " + reason
		}
	}

	cmd := appointment.UpdateStatusCmd{
		AppointmentID: apptID,
		NewStatus:     appointment.StatusCancelled,
		Reason:        reason,
		ChangedBy:     cancelledBy,
	}

	return s.apptRepo.UpdateStatus(ctx, cmd)
}

func (s *Service) UpdateAppointmentStatus(ctx context.Context, cmd appointment.UpdateStatusCmd) error {
	existing, err := s.apptRepo.GetByID(ctx, uuid.Nil, cmd.AppointmentID)
	if err != nil {
		return apperror.NotFound("appointment")
	}

	if !appointment.CanTransition(existing.Status, cmd.NewStatus) {
		return apperror.ValidationFailed("invalid status transition from " + string(existing.Status) + " to " + string(cmd.NewStatus))
	}

	return s.apptRepo.UpdateStatus(ctx, cmd)
}
