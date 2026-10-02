package auth

import (
	"net/http"

	"github.com/appointly/appointly/backend/internal/domain/auth"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/auth"
)

type AuthHandler struct {
	authService *usecase.Service
}

func NewAuthHandler(authService *usecase.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone,omitempty"`
}

type AuthResponse struct {
	User       *auth.User      `json:"user"`
	Tokens     *auth.TokenPair `json:"tokens"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := auth.RegisterCmd{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	user, tokens, err := h.authService.Register(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, AuthResponse{
		User:   user,
		Tokens: tokens,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := auth.LoginCmd{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: r.UserAgent(),
		IPAddress:  r.RemoteAddr,
	}

	user, tokens, err := h.authService.Login(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, AuthResponse{
		User:   user,
		Tokens: tokens,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims != nil {
		var req RefreshRequest
		_ = v1.DecodeJSON(r, &req)
		_ = h.authService.Logout(r.Context(), claims.UserID, req.RefreshToken)
	}

	v1.RespondNoContent(w)
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		v1.RespondError(w, r, apperror.TokenInvalid())
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, user)
}
