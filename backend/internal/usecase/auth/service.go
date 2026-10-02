package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/audit"
	"github.com/appointly/appointly/backend/internal/domain/auth"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	"github.com/appointly/appointly/backend/internal/pkg/crypto"
)

type Service struct {
	authRepo     auth.Repository
	auditRepo    audit.Repository
	tokenService auth.TokenService
}

func NewService(authRepo auth.Repository, auditRepo audit.Repository, tokenService auth.TokenService) *Service {
	return &Service{
		authRepo:     authRepo,
		auditRepo:    auditRepo,
		tokenService: tokenService,
	}
}

func (s *Service) Register(ctx context.Context, cmd auth.RegisterCmd) (*auth.User, *auth.TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, nil, apperror.ValidationFailed("valid email is required")
	}
	if len(cmd.Password) < 8 {
		return nil, nil, apperror.ValidationFailed("password must be at least 8 characters long")
	}

	exists, err := s.authRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, apperror.AlreadyExists("user with email " + email)
	}

	hashedPassword, err := crypto.HashPasswordDefault(cmd.Password)
	if err != nil {
		return nil, nil, apperror.Internal("failed to hash password")
	}

	newUser := &auth.User{
		ID:              uuid.New(),
		Email:           email,
		PasswordHash:    hashedPassword,
		FirstName:       strings.TrimSpace(cmd.FirstName),
		LastName:        strings.TrimSpace(cmd.LastName),
		Phone:           strings.TrimSpace(cmd.Phone),
		IsEmailVerified: false,
		IsSuperAdmin:    false,
	}

	createdUser, err := s.authRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, nil, err
	}

	// Generate initial token pair for registration
	tokenPair, err := s.issueTokenPair(ctx, createdUser, "registration")
	if err != nil {
		return nil, nil, err
	}

	// Log audit event
	if s.auditRepo != nil {
		actorID := createdUser.ID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			ActorID:    &actorID,
			ActorEmail: createdUser.Email,
			ActorType:  "user",
			Action:     "user.registered",
			CreatedAt:  time.Now(),
		})
	}

	return createdUser, tokenPair, nil
}

func (s *Service) Login(ctx context.Context, cmd auth.LoginCmd) (*auth.User, *auth.TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	if email == "" || cmd.Password == "" {
		return nil, nil, apperror.ValidationFailed("email and password are required")
	}

	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, apperror.InvalidCredentials()
	}

	valid, err := crypto.VerifyPassword(cmd.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, nil, apperror.InvalidCredentials()
	}

	tokenPair, err := s.issueTokenPair(ctx, user, cmd.DeviceInfo)
	if err != nil {
		return nil, nil, err
	}

	_ = s.authRepo.UpdateLastLogin(ctx, user.ID)

	// Log login audit event
	if s.auditRepo != nil {
		actorID := user.ID
		_ = s.auditRepo.Create(ctx, audit.AuditLog{
			ActorID:    &actorID,
			ActorEmail: user.Email,
			ActorType:  "user",
			Action:     "user.login",
			IPAddress:  cmd.IPAddress,
			UserAgent:  cmd.DeviceInfo,
			CreatedAt:  time.Now(),
		})
	}

	return user, tokenPair, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (*auth.TokenPair, error) {
	if refreshTokenStr == "" {
		return nil, apperror.TokenInvalid()
	}

	tokenHash := crypto.HashSHA256(refreshTokenStr)
	rt, err := s.authRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil || !rt.IsValid() {
		return nil, apperror.TokenExpired()
	}

	// Revoke old refresh token (Token rotation)
	_ = s.authRepo.RevokeRefreshToken(ctx, tokenHash)

	user, err := s.authRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, apperror.Unauthorized("user no longer exists")
	}

	return s.issueTokenPair(ctx, user, rt.DeviceInfo)
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, refreshTokenStr string) error {
	if refreshTokenStr != "" {
		tokenHash := crypto.HashSHA256(refreshTokenStr)
		_ = s.authRepo.RevokeRefreshToken(ctx, tokenHash)
	}
	return s.authRepo.RevokeAllUserRefreshTokens(ctx, userID)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	return s.authRepo.GetUserByID(ctx, id)
}

func (s *Service) issueTokenPair(ctx context.Context, user *auth.User, deviceInfo string) (*auth.TokenPair, error) {
	claims := auth.Claims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsSuperAdmin,
	}

	accessToken, expiresAt, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, apperror.Internal("failed to generate access token")
	}

	rawRefreshToken, err := crypto.GenerateSecureTokenHex(32)
	if err != nil {
		return nil, apperror.Internal("failed to generate refresh token")
	}

	rtHash := crypto.HashSHA256(rawRefreshToken)
	rt := &auth.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  rtHash,
		DeviceInfo: deviceInfo,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.authRepo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, apperror.Internal("failed to store refresh token")
	}

	return &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}
