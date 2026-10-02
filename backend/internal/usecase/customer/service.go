package customer

import (
	"context"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type Service struct {
	repo customer.Repository
}

func NewService(repo customer.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateCustomer(ctx context.Context, cmd customer.CreateCustomerCmd) (*customer.Customer, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerCreate) {
		return nil, apperror.Unauthorized("insufficient permissions to create customer")
	}

	cmd.OrganizationID = authCtx.OrgID
	return s.repo.Create(ctx, cmd)
}

func (s *Service) GetCustomer(ctx context.Context, customerID uuid.UUID) (*customer.Customer, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerRead) {
		return nil, apperror.Unauthorized("insufficient permissions to view customer")
	}

	return s.repo.GetByID(ctx, authCtx.OrgID, customerID)
}

func (s *Service) UpdateCustomer(ctx context.Context, cmd customer.UpdateCustomerCmd) (*customer.Customer, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerUpdate) {
		return nil, apperror.Unauthorized("insufficient permissions to update customer")
	}

	cmd.OrganizationID = authCtx.OrgID
	return s.repo.Update(ctx, cmd)
}

func (s *Service) DeleteCustomer(ctx context.Context, customerID uuid.UUID) error {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerDelete) {
		return apperror.Unauthorized("insufficient permissions to delete customer")
	}

	return s.repo.SoftDelete(ctx, authCtx.OrgID, customerID)
}

func (s *Service) ListCustomers(ctx context.Context, filter customer.ListCustomersFilter, page, perPage int) ([]*customer.Customer, int, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerRead) {
		return nil, 0, apperror.Unauthorized("insufficient permissions to list customers")
	}

	filter.OrganizationID = authCtx.OrgID
	return s.repo.List(ctx, filter, page, perPage)
}

func (s *Service) LinkUser(ctx context.Context, customerID, targetUserID uuid.UUID) (*customer.Customer, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || (!authCtx.Can(rbac.PermCustomerUpdate) && !authCtx.Can(rbac.PermMemberUpdate)) {
		return nil, apperror.Unauthorized("insufficient permissions to link user to customer")
	}

	cmd := customer.LinkUserCmd{
		OrganizationID: authCtx.OrgID,
		CustomerID:     customerID,
		UserID:         targetUserID,
	}

	return s.repo.LinkUser(ctx, cmd)
}

func (s *Service) AddNote(ctx context.Context, customerID uuid.UUID, content string) (*customer.CustomerNote, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerUpdate) {
		return nil, apperror.Unauthorized("insufficient permissions to add customer note")
	}

	cmd := customer.AddNoteCmd{
		OrganizationID: authCtx.OrgID,
		CustomerID:     customerID,
		AuthorID:       authCtx.UserID,
		AuthorName:     "Staff Member", // In real flow, enriched by auth user profile
		Content:        content,
	}

	return s.repo.AddNote(ctx, cmd)
}

func (s *Service) ListNotes(ctx context.Context, customerID uuid.UUID) ([]*customer.CustomerNote, error) {
	authCtx := rbac.FromContext(ctx)
	if authCtx == nil || !authCtx.Can(rbac.PermCustomerRead) {
		return nil, apperror.Unauthorized("insufficient permissions to view customer notes")
	}

	return s.repo.ListNotes(ctx, authCtx.OrgID, customerID)
}
