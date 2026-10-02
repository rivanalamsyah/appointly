package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// MemberRepository is a thread-safe in-memory implementation of rbac.Repository.
type MemberRepository struct {
	mu          sync.RWMutex
	members     map[uuid.UUID]*rbac.OrganizationMember
	userOrgKey  map[string]uuid.UUID // "userID:orgID" -> memberID
	invitations map[uuid.UUID]*rbac.MemberInvitation
}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		members:     make(map[uuid.UUID]*rbac.OrganizationMember),
		userOrgKey:  make(map[string]uuid.UUID),
		invitations: make(map[uuid.UUID]*rbac.MemberInvitation),
	}
}

func (r *MemberRepository) CreateMember(ctx context.Context, m *rbac.OrganizationMember) (*rbac.OrganizationMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := m.UserID.String() + ":" + m.OrganizationID.String()
	if _, exists := r.userOrgKey[key]; exists {
		return nil, apperror.AlreadyExists("user membership in organization")
	}

	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.IsActive = true

	cp := *m
	r.members[m.ID] = &cp
	r.userOrgKey[key] = m.ID
	return &cp, nil
}

func (r *MemberRepository) GetMemberByID(ctx context.Context, id uuid.UUID) (*rbac.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.members[id]
	if !exists {
		return nil, apperror.NotFound("organization member")
	}
	cp := *m
	return &cp, nil
}

func (r *MemberRepository) GetMemberByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*rbac.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := userID.String() + ":" + orgID.String()
	id, exists := r.userOrgKey[key]
	if !exists {
		return nil, apperror.NotFound("organization member")
	}
	m := r.members[id]
	cp := *m
	return &cp, nil
}

func (r *MemberRepository) ListMembersByOrg(ctx context.Context, orgID uuid.UUID) ([]*rbac.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*rbac.OrganizationMember
	for _, m := range r.members {
		if m.OrganizationID == orgID && m.IsActive {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *MemberRepository) UpdateMemberRole(ctx context.Context, id uuid.UUID, role rbac.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.members[id]
	if !exists {
		return apperror.NotFound("organization member")
	}
	m.Role = role
	m.UpdatedAt = time.Now()
	return nil
}

func (r *MemberRepository) DeactivateMember(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.members[id]
	if !exists {
		return apperror.NotFound("organization member")
	}
	m.IsActive = false
	m.UpdatedAt = time.Now()
	return nil
}

func (r *MemberRepository) CountActiveMembers(ctx context.Context, orgID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, m := range r.members {
		if m.OrganizationID == orgID && m.IsActive {
			count++
		}
	}
	return count, nil
}

func (r *MemberRepository) CreateInvitation(ctx context.Context, inv *rbac.MemberInvitation) (*rbac.MemberInvitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	inv.CreatedAt = time.Now()
	cp := *inv
	r.invitations[inv.ID] = &cp
	return &cp, nil
}

func (r *MemberRepository) GetInvitationByToken(ctx context.Context, tokenHash string) (*rbac.MemberInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, inv := range r.invitations {
		if inv.TokenHash == tokenHash {
			cp := *inv
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("invitation")
}

func (r *MemberRepository) GetPendingInvitationByEmail(ctx context.Context, orgID uuid.UUID, email string) (*rbac.MemberInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emailKey := strings.ToLower(strings.TrimSpace(email))
	for _, inv := range r.invitations {
		if inv.OrganizationID == orgID && strings.ToLower(inv.Email) == emailKey && inv.AcceptedAt == nil {
			cp := *inv
			return &cp, nil
		}
	}
	return nil, apperror.NotFound("pending invitation")
}

func (r *MemberRepository) AcceptInvitation(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inv, exists := r.invitations[id]
	if !exists {
		return apperror.NotFound("invitation")
	}
	now := time.Now()
	inv.AcceptedAt = &now
	return nil
}

func (r *MemberRepository) ListPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]*rbac.MemberInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*rbac.MemberInvitation
	for _, inv := range r.invitations {
		if inv.OrganizationID == orgID && inv.AcceptedAt == nil {
			cp := *inv
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *MemberRepository) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.invitations[id]; !exists {
		return apperror.NotFound("invitation")
	}
	delete(r.invitations, id)
	return nil
}
