package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/staff"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type StaffRepository struct {
	mu            sync.RWMutex
	staffMembers  map[uuid.UUID]*staff.Staff
	staffLocs     map[uuid.UUID][]uuid.UUID // staffID -> locationIDs
	staffSvcs     map[uuid.UUID][]uuid.UUID // staffID -> serviceIDs
}

func NewStaffRepository() *StaffRepository {
	return &StaffRepository{
		staffMembers: make(map[uuid.UUID]*staff.Staff),
		staffLocs:    make(map[uuid.UUID][]uuid.UUID),
		staffSvcs:    make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *StaffRepository) Create(ctx context.Context, cmd staff.CreateStaffCmd) (*staff.Staff, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	staffID := uuid.New()
	now := time.Now()

	s := &staff.Staff{
		ID:             staffID,
		OrganizationID: cmd.OrganizationID,
		UserID:         cmd.UserID,
		FirstName:      strings.TrimSpace(cmd.FirstName),
		LastName:       strings.TrimSpace(cmd.LastName),
		Email:          strings.TrimSpace(cmd.Email),
		Phone:          strings.TrimSpace(cmd.Phone),
		Title:          strings.TrimSpace(cmd.Title),
		Bio:            strings.TrimSpace(cmd.Bio),
		IsActive:       true,
		AcceptsOnline:  cmd.AcceptsOnline,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.staffMembers[staffID] = s
	r.staffLocs[staffID] = cmd.LocationIDs
	r.staffSvcs[staffID] = cmd.ServiceIDs

	cp := *s
	return &cp, nil
}

func (r *StaffRepository) GetByID(ctx context.Context, orgID, staffID uuid.UUID) (*staff.Staff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.staffMembers[staffID]
	if !exists || s.OrganizationID != orgID {
		return nil, apperror.NotFound("staff member")
	}
	cp := *s
	return &cp, nil
}

func (r *StaffRepository) GetByUserID(ctx context.Context, orgID, userID uuid.UUID) (*staff.Staff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.staffMembers {
		if s.OrganizationID == orgID && s.UserID != nil && *s.UserID == userID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("staff member for user")
}

func (r *StaffRepository) List(ctx context.Context, filter staff.ListStaffFilter, page, perPage int) ([]*staff.Staff, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*staff.Staff
	for _, s := range r.staffMembers {
		if s.OrganizationID != filter.OrganizationID {
			continue
		}
		if filter.IsActive != nil && s.IsActive != *filter.IsActive {
			continue
		}
		if filter.AcceptsOnline != nil && s.AcceptsOnline != *filter.AcceptsOnline {
			continue
		}
		if filter.Search != "" {
			query := strings.ToLower(filter.Search)
			fullName := strings.ToLower(s.FullName())
			if !strings.Contains(fullName, query) && !strings.Contains(strings.ToLower(s.Email), query) {
				continue
			}
		}
		cp := *s
		matched = append(matched, &cp)
	}

	total := len(matched)
	start := (page - 1) * perPage
	if start >= total {
		return []*staff.Staff{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *StaffRepository) Update(ctx context.Context, cmd staff.UpdateStaffCmd) (*staff.Staff, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, exists := r.staffMembers[cmd.ID]
	if !exists || s.OrganizationID != cmd.OrganizationID {
		return nil, apperror.NotFound("staff member")
	}

	if cmd.FirstName != nil {
		s.FirstName = strings.TrimSpace(*cmd.FirstName)
	}
	if cmd.LastName != nil {
		s.LastName = strings.TrimSpace(*cmd.LastName)
	}
	if cmd.Email != nil {
		s.Email = strings.TrimSpace(*cmd.Email)
	}
	if cmd.Phone != nil {
		s.Phone = strings.TrimSpace(*cmd.Phone)
	}
	if cmd.AvatarURL != nil {
		s.AvatarURL = *cmd.AvatarURL
	}
	if cmd.Title != nil {
		s.Title = strings.TrimSpace(*cmd.Title)
	}
	if cmd.Bio != nil {
		s.Bio = strings.TrimSpace(*cmd.Bio)
	}
	if cmd.AcceptsOnline != nil {
		s.AcceptsOnline = *cmd.AcceptsOnline
	}
	if cmd.IsActive != nil {
		s.IsActive = *cmd.IsActive
	}

	s.UpdatedAt = time.Now()
	cp := *s
	return &cp, nil
}

func (r *StaffRepository) Delete(ctx context.Context, orgID, staffID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, exists := r.staffMembers[staffID]
	if !exists || s.OrganizationID != orgID {
		return apperror.NotFound("staff member")
	}
	// Soft delete: deactivate staff so historical appointments stay intact
	s.IsActive = false
	s.UpdatedAt = time.Now()
	return nil
}

func (r *StaffRepository) GetStaffLocations(ctx context.Context, staffID uuid.UUID) ([]*staff.StaffLocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	locIDs := r.staffLocs[staffID]
	var res []*staff.StaffLocation
	for i, lid := range locIDs {
		res = append(res, &staff.StaffLocation{
			StaffID:    staffID,
			LocationID: lid,
			IsPrimary:  i == 0,
		})
	}
	return res, nil
}

func (r *StaffRepository) SetStaffLocations(ctx context.Context, orgID, staffID uuid.UUID, locationIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, exists := r.staffMembers[staffID]
	if !exists || s.OrganizationID != orgID {
		return apperror.NotFound("staff member")
	}
	r.staffLocs[staffID] = locationIDs
	return nil
}

func (r *StaffRepository) GetStaffServices(ctx context.Context, staffID uuid.UUID) ([]*staff.StaffService, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svcIDs := r.staffSvcs[staffID]
	var res []*staff.StaffService
	s := r.staffMembers[staffID]
	orgID := uuid.Nil
	if s != nil {
		orgID = s.OrganizationID
	}
	for _, sid := range svcIDs {
		res = append(res, &staff.StaffService{
			StaffID:        staffID,
			ServiceID:      sid,
			OrganizationID: orgID,
		})
	}
	return res, nil
}

func (r *StaffRepository) SetStaffServices(ctx context.Context, orgID, staffID uuid.UUID, serviceIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, exists := r.staffMembers[staffID]
	if !exists || s.OrganizationID != orgID {
		return apperror.NotFound("staff member")
	}
	r.staffSvcs[staffID] = serviceIDs
	return nil
}

func (r *StaffRepository) FindAvailableStaff(ctx context.Context, orgID, serviceID, locationID uuid.UUID) ([]*staff.Staff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*staff.Staff
	for sid, s := range r.staffMembers {
		if s.OrganizationID != orgID || !s.IsActive {
			continue
		}
		// check if staff serves serviceID
		svcIDs := r.staffSvcs[sid]
		servesSvc := false
		for _, svcid := range svcIDs {
			if svcid == serviceID {
				servesSvc = true
				break
			}
		}
		if servesSvc {
			cp := *s
			matched = append(matched, &cp)
		}
	}
	return matched, nil
}

func (r *StaffRepository) CountActive(ctx context.Context, orgID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, s := range r.staffMembers {
		if s.OrganizationID == orgID && s.IsActive {
			count++
		}
	}
	return count, nil
}
