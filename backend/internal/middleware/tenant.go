package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

const (
	tenantCtxKey contextKey = "tenant_context"
)

// TenantContext holds resolved organization/tenant metadata for the request.
type TenantContext struct {
	ID   uuid.UUID
	Slug string
	Name string
}

// TenantResolver middleware resolves tenant from X-Tenant-Slug header or host subdomain.
func TenantResolver(orgRepo organization.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := r.Header.Get("X-Tenant-Slug")
			if slug == "" {
				// Try parsing from Host header (e.g. luxe-salon.appointly.app)
				host := r.Host
				if idx := strings.Index(host, ":"); idx != -1 {
					host = host[:idx]
				}
				parts := strings.Split(host, ".")
				if len(parts) >= 3 {
					slug = parts[0]
				}
			}

			if slug == "" {
				// Default fallback for demo/development if unspecified
				slug = "default"
			}

			// Retrieve tenant from DB repository
			org, err := orgRepo.GetBySlug(r.Context(), slug)
			if err != nil {
				// If not found in DB, return 404 tenant not found
				RespondError(w, r, apperror.NotFound("tenant entity not found for slug: "+slug))
				return
			}

			tc := TenantContext{
				ID:   org.ID,
				Slug: org.Slug,
				Name: org.Name,
			}

			ctx := context.WithValue(r.Context(), tenantCtxKey, tc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantFromContext extracts TenantContext from HTTP request context.
func TenantFromContext(ctx context.Context) (TenantContext, error) {
	if tc, ok := ctx.Value(tenantCtxKey).(TenantContext); ok {
		return tc, nil
	}
	return TenantContext{}, apperror.Unauthorized("tenant context missing from request")
}

// OrgIDFromContext extracts organization ID string from TenantContext or auth context.
func OrgIDFromContext(ctx context.Context) string {
	if tc, ok := ctx.Value(tenantCtxKey).(TenantContext); ok && tc.ID != uuid.Nil {
		return tc.ID.String()
	}
	if claims := ClaimsFromContext(ctx); claims != nil && claims.OrgID != uuid.Nil {
		return claims.OrgID.String()
	}
	return ""
}
