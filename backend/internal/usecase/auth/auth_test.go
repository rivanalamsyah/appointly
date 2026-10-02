package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/appointly/appointly/backend/internal/domain/auth"
	jwtpkg "github.com/appointly/appointly/backend/internal/pkg/jwt"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	usecase "github.com/appointly/appointly/backend/internal/usecase/auth"
)

func setupAuthService() (*usecase.Service, *memory.AuthRepository, *memory.AuditRepository) {
	authRepo := memory.NewAuthRepository()
	auditRepo := memory.NewAuditRepository()
	tokenSvc := jwtpkg.NewJWTService("test-secret-key-at-least-32-bytes-long!", "appointly-test", 15*time.Minute)
	svc := usecase.NewService(authRepo, auditRepo, tokenSvc)
	return svc, authRepo, auditRepo
}

func TestRegister_Success(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	cmd := auth.RegisterCmd{
		Email:     "user@example.com",
		Password:  "SecurePassword123!",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+15551234567",
	}

	user, tokens, err := svc.Register(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error on register, got %v", err)
	}

	if user.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", user.Email)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected token pair to be populated")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	cmd := auth.RegisterCmd{
		Email:     "duplicate@example.com",
		Password:  "SecurePassword123!",
		FirstName: "John",
		LastName:  "Doe",
	}

	_, _, err := svc.Register(ctx, cmd)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, _, err = svc.Register(ctx, cmd)
	if err == nil {
		t.Fatalf("expected error on duplicate email registration, got nil")
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	cmd := auth.RegisterCmd{
		Email:     "short@example.com",
		Password:  "12345", // < 8 chars
		FirstName: "John",
		LastName:  "Doe",
	}

	_, _, err := svc.Register(ctx, cmd)
	if err == nil {
		t.Fatalf("expected validation error for short password, got nil")
	}
}

func TestLogin_Success(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	regCmd := auth.RegisterCmd{
		Email:     "login@example.com",
		Password:  "MySecretPass123",
		FirstName: "Jane",
		LastName:  "Smith",
	}
	_, _, err := svc.Register(ctx, regCmd)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	loginCmd := auth.LoginCmd{
		Email:      "login@example.com",
		Password:   "MySecretPass123",
		DeviceInfo: "TestAgent",
		IPAddress:  "127.0.0.1",
	}

	user, tokens, err := svc.Login(ctx, loginCmd)
	if err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}

	if user.Email != "login@example.com" {
		t.Errorf("expected email login@example.com, got %s", user.Email)
	}
	if tokens.AccessToken == "" {
		t.Errorf("expected access token to be generated")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	regCmd := auth.RegisterCmd{
		Email:     "wrongpass@example.com",
		Password:  "CorrectPassword123",
		FirstName: "Alice",
		LastName:  "Wonder",
	}
	_, _, _ = svc.Register(ctx, regCmd)

	loginCmd := auth.LoginCmd{
		Email:    "wrongpass@example.com",
		Password: "IncorrectPassword",
	}

	_, _, err := svc.Login(ctx, loginCmd)
	if err == nil {
		t.Fatalf("expected error on invalid password, got nil")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	svc, _, _ := setupAuthService()
	ctx := context.Background()

	regCmd := auth.RegisterCmd{
		Email:     "refresh@example.com",
		Password:  "ValidPass123!",
		FirstName: "Bob",
		LastName:  "Builder",
	}
	_, initialTokens, err := svc.Register(ctx, regCmd)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	newTokens, err := svc.RefreshToken(ctx, initialTokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected successful token refresh, got %v", err)
	}

	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Errorf("expected new token pair after refresh")
	}

	// Used refresh token should be revoked (token rotation)
	_, err = svc.RefreshToken(ctx, initialTokens.RefreshToken)
	if err == nil {
		t.Fatalf("expected error when re-using rotated refresh token, got nil")
	}
}
