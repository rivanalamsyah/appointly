package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/auth"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// AuthRepository is a thread-safe in-memory repository for authentication domain entities.
type AuthRepository struct {
	mu            sync.RWMutex
	users         map[uuid.UUID]*auth.User
	emailIndex    map[string]uuid.UUID
	tokens        map[string]*auth.RefreshToken
	verifications map[string]*auth.EmailVerification
	resets        map[string]*auth.PasswordReset
}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{
		users:         make(map[uuid.UUID]*auth.User),
		emailIndex:    make(map[string]uuid.UUID),
		tokens:        make(map[string]*auth.RefreshToken),
		verifications: make(map[string]*auth.EmailVerification),
		resets:        make(map[string]*auth.PasswordReset),
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, u *auth.User) (*auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailKey := strings.ToLower(strings.TrimSpace(u.Email))
	if _, exists := r.emailIndex[emailKey]; exists {
		return nil, apperror.AlreadyExists("user with email " + u.Email)
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	// Deep copy
	cp := *u
	r.users[u.ID] = &cp
	r.emailIndex[emailKey] = u.ID

	return &cp, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	if !exists {
		return nil, apperror.NotFound("user")
	}
	cp := *u
	return &cp, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emailKey := strings.ToLower(strings.TrimSpace(email))
	id, exists := r.emailIndex[emailKey]
	if !exists {
		return nil, apperror.NotFound("user")
	}
	u := r.users[id]
	cp := *u
	return &cp, nil
}

func (r *AuthRepository) UpdateUser(ctx context.Context, u *auth.User) (*auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.users[u.ID]
	if !exists {
		return nil, apperror.NotFound("user")
	}

	u.UpdatedAt = time.Now()
	u.CreatedAt = existing.CreatedAt
	cp := *u
	r.users[u.ID] = &cp
	return &cp, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[userID]
	if !exists {
		return apperror.NotFound("user")
	}
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
	return nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[userID]
	if !exists {
		return apperror.NotFound("user")
	}
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
	return nil
}

func (r *AuthRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emailKey := strings.ToLower(strings.TrimSpace(email))
	_, exists := r.emailIndex[emailKey]
	return exists, nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, rt *auth.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	rt.CreatedAt = time.Now()
	cp := *rt
	r.tokens[rt.TokenHash] = &cp
	return nil
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rt, exists := r.tokens[tokenHash]
	if !exists {
		return nil, apperror.NotFound("refresh token")
	}
	cp := *rt
	return &cp, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rt, exists := r.tokens[tokenHash]
	if !exists {
		return apperror.NotFound("refresh token")
	}
	now := time.Now()
	rt.RevokedAt = &now
	return nil
}

func (r *AuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, rt := range r.tokens {
		if rt.UserID == userID && rt.RevokedAt == nil {
			rt.RevokedAt = &now
		}
	}
	return nil
}

func (r *AuthRepository) CleanExpiredRefreshTokens(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for k, rt := range r.tokens {
		if now.After(rt.ExpiresAt) || rt.RevokedAt != nil {
			delete(r.tokens, k)
		}
	}
	return nil
}

func (r *AuthRepository) CreateEmailVerification(ctx context.Context, ev *auth.EmailVerification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	ev.CreatedAt = time.Now()
	cp := *ev
	r.verifications[ev.TokenHash] = &cp
	return nil
}

func (r *AuthRepository) GetEmailVerification(ctx context.Context, tokenHash string) (*auth.EmailVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ev, exists := r.verifications[tokenHash]
	if !exists {
		return nil, apperror.NotFound("email verification token")
	}
	cp := *ev
	return &cp, nil
}

func (r *AuthRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, exists := r.users[userID]
	if !exists {
		return apperror.NotFound("user")
	}
	u.IsEmailVerified = true
	u.UpdatedAt = time.Now()
	return nil
}

func (r *AuthRepository) DeleteEmailVerification(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, ev := range r.verifications {
		if ev.ID == id {
			delete(r.verifications, k)
			break
		}
	}
	return nil
}

func (r *AuthRepository) CreatePasswordReset(ctx context.Context, pr *auth.PasswordReset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	pr.CreatedAt = time.Now()
	cp := *pr
	r.resets[pr.TokenHash] = &cp
	return nil
}

func (r *AuthRepository) GetPasswordReset(ctx context.Context, tokenHash string) (*auth.PasswordReset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pr, exists := r.resets[tokenHash]
	if !exists {
		return nil, apperror.NotFound("password reset token")
	}
	cp := *pr
	return &cp, nil
}

func (r *AuthRepository) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for _, pr := range r.resets {
		if pr.ID == id {
			pr.UsedAt = &now
			break
		}
	}
	return nil
}

func (r *AuthRepository) DeleteExpiredPasswordResets(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, pr := range r.resets {
		if now.After(pr.ExpiresAt) || pr.UsedAt != nil {
			delete(r.resets, k)
		}
	}
	return nil
}
