package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/auth"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

const (
	claimsCtxKey contextKey = "user_claims"
)

// AuthenticateJWT parses and validates Bearer tokens.
func AuthenticateJWT(tokenService auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims, err := tokenService.ValidateAccessToken(tokenString)
			if err != nil {
				// Don't abort here — let RequireAuth decide if token is required
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext extracts user claims from HTTP request context.
func ClaimsFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(claimsCtxKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// RequireAuth enforces that a valid authenticated user claims exists.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			RespondError(w, r, apperror.Unauthorized("authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireTenantMember resolves tenant membership for the user and constructs rbac.AuthContext.
func RequireTenantMember(orgRepo organization.Repository, memberRepo rbac.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				RespondError(w, r, apperror.Unauthorized("authentication required"))
				return
			}

			// Extract org ID from path param "orgID" or header "X-Tenant-Id"
			orgIDStr := chi.URLParam(r, "orgID")
			if orgIDStr == "" {
				orgIDStr = r.Header.Get("X-Tenant-Id")
			}

			var orgID uuid.UUID
			var err error
			if orgIDStr != "" {
				orgID, err = uuid.Parse(orgIDStr)
				if err != nil {
					RespondError(w, r, apperror.BadRequest("invalid tenant organization ID"))
					return
				}
			} else if claims.OrgID != uuid.Nil {
				orgID = claims.OrgID
			} else {
				// Resolve from slug header if available
				slug := r.Header.Get("X-Tenant-Slug")
				if slug != "" {
					orgObj, err := orgRepo.GetBySlug(r.Context(), slug)
					if err == nil {
						orgID = orgObj.ID
					}
				}
			}

			if orgID == uuid.Nil {
				RespondError(w, r, apperror.BadRequest("organization tenant context is required"))
				return
			}

			// Retrieve membership
			member, err := memberRepo.GetMemberByUserAndOrg(r.Context(), claims.UserID, orgID)
			if err != nil || !member.IsActive {
				RespondError(w, r, apperror.Forbidden("you are not an active member of this organization"))
				return
			}

			org, err := orgRepo.GetByID(r.Context(), orgID)
			if err != nil || org.Status != organization.StatusActive {
				RespondError(w, r, apperror.Forbidden("organization is inactive or suspended"))
				return
			}

			authCtx := &rbac.AuthContext{
				UserID:       claims.UserID,
				OrgID:        orgID,
				Role:         member.Role,
				IsOwner:      org.OwnerID == claims.UserID,
				IsSuperAdmin: claims.IsAdmin,
			}

			ctx := rbac.NewContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission ensures the authenticated user has a specific permission in the active tenant context.
func RequirePermission(perm rbac.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := rbac.FromContext(r.Context())
			if authCtx == nil {
				RespondError(w, r, apperror.Unauthorized("authentication and organization context required"))
				return
			}

			if !authCtx.Can(perm) {
				RespondError(w, r, apperror.Forbidden("permission denied: requires "+string(perm)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
