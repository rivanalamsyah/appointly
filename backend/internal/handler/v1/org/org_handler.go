package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/org"
)

type OrgHandler struct {
	orgService *usecase.Service
}

func NewOrgHandler(orgService *usecase.Service) *OrgHandler {
	return &OrgHandler{
		orgService: orgService,
	}
}

type CreateOrgRequest struct {
	Name         string                    `json:"name"`
	Slug         string                    `json:"slug,omitempty"`
	BusinessType organization.BusinessType `json:"business_type,omitempty"`
	Email        string                    `json:"email"`
	Phone        string                    `json:"phone,omitempty"`
	Timezone     string                    `json:"timezone,omitempty"`
	Currency     string                    `json:"currency,omitempty"`
	Country      string                    `json:"country,omitempty"`
}

type CreateOrgResponse struct {
	Organization *organization.Organization `json:"organization"`
	Member       *rbac.OrganizationMember   `json:"member"`
}

// CreateOrganization handles POST /api/v1/organizations
func (h *OrgHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		v1.RespondError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	var req CreateOrgRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := organization.CreateOrganizationCmd{
		Name:         req.Name,
		Slug:         req.Slug,
		BusinessType: req.BusinessType,
		Email:        req.Email,
		Phone:        req.Phone,
		Timezone:     req.Timezone,
		Currency:     req.Currency,
		Country:      req.Country,
		OwnerID:      claims.UserID,
	}

	org, member, err := h.orgService.CreateOrganization(r.Context(), claims.UserID, cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, CreateOrgResponse{
		Organization: org,
		Member:       member,
	})
}

// ListUserOrganizations handles GET /api/v1/organizations
func (h *OrgHandler) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		v1.RespondError(w, r, nil)
		return
	}

	orgs, err := h.orgService.ListUserOrganizations(r.Context(), claims.UserID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, orgs)
}

// GetOrganization handles GET /api/v1/organizations/{orgID}
func (h *OrgHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	org, err := h.orgService.GetOrganization(r.Context(), authCtx, orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, org)
}

// ListMembers handles GET /api/v1/organizations/{orgID}/members
func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	members, err := h.orgService.ListMembers(r.Context(), authCtx, orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, members)
}

type AddMemberRequest struct {
	UserID string    `json:"user_id"`
	Role   rbac.Role `json:"role"`
}

// AddMember handles POST /api/v1/organizations/{orgID}/members
func (h *OrgHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var req AddMemberRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	member, err := h.orgService.AddMember(r.Context(), authCtx, orgID, targetUserID, req.Role)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, member)
}

type UpdateRoleRequest struct {
	Role rbac.Role `json:"role"`
}

// UpdateMemberRole handles PATCH /api/v1/organizations/{orgID}/members/{memberID}
func (h *OrgHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	memberIDStr := chi.URLParam(r, "memberID")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	var req UpdateRoleRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	if err := h.orgService.UpdateMemberRole(r.Context(), authCtx, orgID, memberID, req.Role); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}

// RemoveMember handles DELETE /api/v1/organizations/{orgID}/members/{memberID}
func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	authCtx := rbac.FromContext(r.Context())
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	memberIDStr := chi.URLParam(r, "memberID")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	if err := h.orgService.RemoveMember(r.Context(), authCtx, orgID, memberID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondNoContent(w)
}
