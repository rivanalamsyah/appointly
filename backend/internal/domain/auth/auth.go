// Package auth defines the authentication and identity domain.
// This package manages users (identity accounts), refresh tokens,
// email verification, and password reset flows.
// NOTE: Authorization (RBAC) is handled by the rbac domain package.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated identity in the system.
// A user can be a member of multiple organizations.
// Customers (who book appointments) may or may not have a user account.
type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`          // never serialized to JSON
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Phone           string     `json:"phone,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	IsEmailVerified bool       `json:"is_email_verified"`
	IsSuperAdmin    bool       `json:"is_super_admin"` // platform-level admin
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// RefreshToken represents an opaque refresh token stored in the database.
// Refresh tokens are rotated on each use (one-time use).
type RefreshToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"` // SHA-256 hash of the actual token
	DeviceInfo string     `json:"device_info,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsValid returns true if the refresh token has not expired or been revoked.
func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && time.Now().Before(rt.ExpiresAt)
}

// EmailVerification represents a pending email verification token.
type EmailVerification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// PasswordReset represents a pending password reset token.
type PasswordReset struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsValid returns true if the reset token has not expired or been used.
func (pr *PasswordReset) IsValid() bool {
	return pr.UsedAt == nil && time.Now().Before(pr.ExpiresAt)
}

// --- Commands ----------------------------------------------------------------

// RegisterCmd holds data for new user registration.
type RegisterCmd struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

// LoginCmd holds data for user login.
type LoginCmd struct {
	Email      string
	Password   string
	DeviceInfo string
	IPAddress  string
}

// ChangePasswordCmd holds data for changing password.
type ChangePasswordCmd struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
}

// ResetPasswordCmd holds data for resetting password via token.
type ResetPasswordCmd struct {
	Token       string
	NewPassword string
}

// UpdateProfileCmd holds data for updating user profile.
type UpdateProfileCmd struct {
	UserID    uuid.UUID
	FirstName *string
	LastName  *string
	Phone     *string
	AvatarURL *string
}

// --- Results -----------------------------------------------------------------

// TokenPair is returned after successful login or token refresh.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"` // "Bearer"
}

// Claims holds the JWT claims for an access token.
type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	Email     string    `json:"email"`
	OrgID     uuid.UUID `json:"org_id,omitempty"` // active organization
	Role      string    `json:"role,omitempty"`   // role in active org
	IsAdmin   bool      `json:"is_admin,omitempty"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// --- Repository Interface ----------------------------------------------------

// Repository defines the persistence contract for the auth domain.
type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)

	// Refresh Tokens
	CreateRefreshToken(ctx context.Context, rt *RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	CleanExpiredRefreshTokens(ctx context.Context) error

	// Email Verification
	CreateEmailVerification(ctx context.Context, ev *EmailVerification) error
	GetEmailVerification(ctx context.Context, tokenHash string) (*EmailVerification, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	DeleteEmailVerification(ctx context.Context, id uuid.UUID) error

	// Password Reset
	CreatePasswordReset(ctx context.Context, pr *PasswordReset) error
	GetPasswordReset(ctx context.Context, tokenHash string) (*PasswordReset, error)
	MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredPasswordResets(ctx context.Context) error
}

// TokenService defines the contract for JWT operations.
// Implementation lives in the infrastructure layer.
type TokenService interface {
	// GenerateAccessToken creates a signed JWT access token from claims.
	GenerateAccessToken(claims Claims) (string, time.Time, error)

	// ValidateAccessToken parses and validates a JWT, returning claims.
	ValidateAccessToken(tokenString string) (*Claims, error)
}
