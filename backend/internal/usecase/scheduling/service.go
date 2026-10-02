package scheduling

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/scheduling"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

var hhmmRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// ParseHHMMMinutes converts "HH:MM" to minutes from midnight (0 to 1439).
func ParseHHMMMinutes(timeStr string) (int, error) {
	if !hhmmRegex.MatchString(timeStr) {
		return 0, fmt.Errorf("invalid time format %q, expected HH:MM (00:00-23:59)", timeStr)
	}
	parts := regexp.MustCompile(`:`).Split(timeStr, 2)
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m, nil
}

// ValidateInterval ensures start < end for HH:MM time strings.
func ValidateInterval(startStr, endStr string) error {
	sMin, err := ParseHHMMMinutes(startStr)
	if err != nil {
		return apperror.ValidationFailed(err.Error())
	}
	eMin, err := ParseHHMMMinutes(endStr)
	if err != nil {
		return apperror.ValidationFailed(err.Error())
	}
	if sMin >= eMin {
		return apperror.ValidationFailed(fmt.Sprintf("start_time (%s) must be strictly before end_time (%s)", startStr, endStr))
	}
	return nil
}

type Service struct {
	schedRepo scheduling.Repository
	auditRepo audit.Repository
}

func NewService(schedRepo scheduling.Repository, auditRepo audit.Repository) *Service {
	return &Service{
		schedRepo: schedRepo,
		auditRepo: auditRepo,
	}
}

func (s *Service) GetBusinessHours(ctx context.Context, authCtx *rbac.AuthContext, orgID, locationID uuid.UUID) ([]*scheduling.BusinessHours, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermOrgRead) {
		return nil, apperror.Forbidden("permission org:read required")
	}
	return s.schedRepo.GetBusinessHoursByLocation(ctx, orgID, locationID)
}

func (s *Service) UpsertBusinessHours(ctx context.Context, authCtx *rbac.AuthContext, cmd scheduling.UpsertBusinessHoursCmd) error {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermOrgManageSettings) {
		return apperror.Forbidden("permission org:manage_settings required")
	}

	for _, bh := range cmd.Hours {
		if bh.IsOpen {
			if err := ValidateInterval(bh.OpenTime, bh.CloseTime); err != nil {
				return err
			}
		}
	}

	if err := s.schedRepo.UpsertBusinessHours(ctx, cmd); err != nil {
		return err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "org.settings_updated",
			ResourceType:   "business_hours",
			CreatedAt:      time.Now(),
		})
	}

	return nil
}

func (s *Service) GetStaffSchedule(ctx context.Context, authCtx *rbac.AuthContext, orgID, staffID uuid.UUID) ([]*scheduling.StaffSchedule, error) {
	if authCtx == nil || authCtx.OrgID != orgID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffRead) {
		return nil, apperror.Forbidden("permission staff:read required")
	}
	return s.schedRepo.GetStaffSchedule(ctx, orgID, staffID)
}

func (s *Service) UpsertStaffSchedule(ctx context.Context, authCtx *rbac.AuthContext, cmd scheduling.UpsertStaffScheduleCmd) error {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffSchedule) {
		return apperror.Forbidden("permission staff:manage_schedule required")
	}

	for _, ss := range cmd.Schedule {
		if ss.IsWorking {
			if err := ValidateInterval(ss.StartTime, ss.EndTime); err != nil {
				return err
			}
			if ss.BreakStart != nil && ss.BreakEnd != nil && *ss.BreakStart != "" && *ss.BreakEnd != "" {
				if err := ValidateInterval(*ss.BreakStart, *ss.BreakEnd); err != nil {
					return err
				}
				sMin, _ := ParseHHMMMinutes(ss.StartTime)
				eMin, _ := ParseHHMMMinutes(ss.EndTime)
				bsMin, _ := ParseHHMMMinutes(*ss.BreakStart)
				beMin, _ := ParseHHMMMinutes(*ss.BreakEnd)

				if bsMin < sMin || beMin > eMin {
					return apperror.ValidationFailed(fmt.Sprintf("break interval (%s - %s) must be inside working hours (%s - %s)", *ss.BreakStart, *ss.BreakEnd, ss.StartTime, ss.EndTime))
				}
			}
		}
	}

	if err := s.schedRepo.UpsertStaffSchedule(ctx, cmd); err != nil {
		return err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.updated",
			ResourceType:   "staff_schedule",
			ResourceID:     &cmd.StaffID,
		})
	}

	return nil
}

func (s *Service) CreateStaffTimeOff(ctx context.Context, authCtx *rbac.AuthContext, cmd scheduling.CreateTimeOffCmd) (*scheduling.StaffTimeOff, error) {
	if authCtx == nil || authCtx.OrgID != cmd.OrganizationID {
		return nil, apperror.Forbidden("cross-tenant access attempt denied")
	}
	if !authCtx.Can(rbac.PermStaffSchedule) {
		return nil, apperror.Forbidden("permission staff:manage_schedule required")
	}

	if cmd.EndDate.Before(cmd.StartDate) {
		return nil, apperror.ValidationFailed("end_date cannot be before start_date")
	}

	if !cmd.IsAllDay {
		if cmd.StartTime == nil || cmd.EndTime == nil {
			return nil, apperror.ValidationFailed("start_time and end_time are required when is_all_day is false")
		}
		if err := ValidateInterval(*cmd.StartTime, *cmd.EndTime); err != nil {
			return nil, err
		}
	}

	tOff := &scheduling.StaffTimeOff{
		ID:             uuid.New(),
		OrganizationID: cmd.OrganizationID,
		StaffID:        cmd.StaffID,
		StartDate:      cmd.StartDate,
		EndDate:        cmd.EndDate,
		Reason:         cmd.Reason,
		IsAllDay:       cmd.IsAllDay,
		StartTime:      cmd.StartTime,
		EndTime:        cmd.EndTime,
		ApprovedBy:     &authCtx.UserID,
	}

	created, err := s.schedRepo.CreateTimeOff(ctx, tOff)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		orgID := cmd.OrganizationID
		actorID := authCtx.UserID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			OrganizationID: &orgID,
			ActorID:        &actorID,
			Action:         "staff.updated",
			ResourceType:   "staff_time_off",
			ResourceID:     &created.ID,
		})
	}

	return created, nil
}
