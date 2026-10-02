package customer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/repository/memory"
	customeruc "github.com/appointly/appointly/backend/internal/usecase/customer"
)

func setupTestCustomerService() (*customeruc.Service, *memory.CustomerRepository, uuid.UUID, uuid.UUID) {
	repo := memory.NewCustomerRepository()
	service := customeruc.NewService(repo)
	orgID := uuid.New()
	userID := uuid.New()

	return service, repo, orgID, userID
}

func contextWithAuth(orgID, userID uuid.UUID, role rbac.Role) context.Context {
	return rbac.NewContext(context.Background(), &rbac.AuthContext{
		UserID:  userID,
		OrgID:   orgID,
		Role:    role,
		IsOwner: role == rbac.RoleOwner,
	})
}

func TestCustomer_CreateAndGet(t *testing.T) {
	svc, _, orgID, userID := setupTestCustomerService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	cmd := customer.CreateCustomerCmd{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		Phone:     "+15551234567",
		Notes:     "Prefers morning appointments",
	}

	created, err := svc.CreateCustomer(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error creating customer, got %v", err)
	}

	if created.FirstName != "Alice" || created.LastName != "Smith" {
		t.Errorf("expected Alice Smith, got %s", created.FullName())
	}
	if created.Email != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %s", created.Email)
	}

	retrieved, err := svc.GetCustomer(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error retrieving customer, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected customer ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestCustomer_DuplicateHandling(t *testing.T) {
	svc, _, orgID, userID := setupTestCustomerService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	cmd1 := customer.CreateCustomerCmd{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+15559876543",
	}
	_, err := svc.CreateCustomer(ctx, cmd1)
	if err != nil {
		t.Fatalf("failed to create first customer: %v", err)
	}

	// Attempt duplicate email in same org
	cmdDuplicateEmail := customer.CreateCustomerCmd{
		FirstName: "Johnny",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}
	_, err = svc.CreateCustomer(ctx, cmdDuplicateEmail)
	if err == nil {
		t.Error("expected error creating customer with duplicate email, got nil")
	}

	// Attempt duplicate phone in same org
	cmdDuplicatePhone := customer.CreateCustomerCmd{
		FirstName: "Jack",
		LastName:  "Doe",
		Phone:     "+15559876543",
	}
	_, err = svc.CreateCustomer(ctx, cmdDuplicatePhone)
	if err == nil {
		t.Error("expected error creating customer with duplicate phone, got nil")
	}
}

func TestCustomer_SearchAndPagination(t *testing.T) {
	svc, _, orgID, userID := setupTestCustomerService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	names := []string{"Bob Marley", "Bob Dylan", "Charlie Puth", "David Bowie"}
	for _, n := range names {
		parts := []string{n, ""}
		if idx := len(n) - len("Marley"); n[idx:] == "Marley" {
			parts = []string{"Bob", "Marley"}
		}
		_, err := svc.CreateCustomer(ctx, customer.CreateCustomerCmd{
			FirstName: parts[0],
			LastName:  parts[1],
		})
		if err != nil {
			t.Fatalf("failed to create customer: %v", err)
		}
	}

	// Search "Bob"
	list, total, err := svc.ListCustomers(ctx, customer.ListCustomersFilter{
		Search: "Bob",
	}, 1, 10)

	if err != nil {
		t.Fatalf("failed to search customers: %v", err)
	}
	if total < 2 {
		t.Errorf("expected at least 2 customers matching 'Bob', got %d", total)
	}
	if len(list) < 2 {
		t.Errorf("expected 2 items in page, got %d", len(list))
	}
}

func TestCustomer_TenantIsolation(t *testing.T) {
	svc, _, orgA, userA := setupTestCustomerService()
	orgB := uuid.New()
	userB := uuid.New()

	ctxA := contextWithAuth(orgA, userA, rbac.RoleOwner)
	ctxB := contextWithAuth(orgB, userB, rbac.RoleOwner)

	created, err := svc.CreateCustomer(ctxA, customer.CreateCustomerCmd{
		FirstName: "Secret",
		LastName:  "Customer",
		Email:     "secret@orga.com",
	})
	if err != nil {
		t.Fatalf("failed to create customer in Org A: %v", err)
	}

	// Org B attempting to fetch Org A's customer
	_, err = svc.GetCustomer(ctxB, created.ID)
	if err == nil {
		t.Error("expected error when Org B accesses Org A customer, got nil")
	}

	// Org B list should return 0 customers
	list, total, err := svc.ListCustomers(ctxB, customer.ListCustomersFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("failed to list customers for Org B: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Errorf("expected 0 customers for Org B, got total %d", total)
	}
}

func TestCustomer_IdentityLinkingAndNotes(t *testing.T) {
	svc, _, orgID, userID := setupTestCustomerService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	created, err := svc.CreateCustomer(ctx, customer.CreateCustomerCmd{
		FirstName: "Emma",
		LastName:  "Watson",
		Email:     "emma@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// Explicit Identity Linking
	appUserID := uuid.New()
	linked, err := svc.LinkUser(ctx, created.ID, appUserID)
	if err != nil {
		t.Fatalf("failed to link user account to customer: %v", err)
	}
	if linked.UserID == nil || *linked.UserID != appUserID {
		t.Errorf("expected UserID to be %s, got %v", appUserID, linked.UserID)
	}

	// Add Timeline Note
	note, err := svc.AddNote(ctx, created.ID, "Customer requested sensitive skin products only.")
	if err != nil {
		t.Fatalf("failed to add customer note: %v", err)
	}
	if note.Content != "Customer requested sensitive skin products only." {
		t.Errorf("unexpected note content: %s", note.Content)
	}

	notes, err := svc.ListNotes(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to list notes: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func TestCustomer_SoftDelete(t *testing.T) {
	svc, _, orgID, userID := setupTestCustomerService()
	ctx := contextWithAuth(orgID, userID, rbac.RoleOwner)

	created, err := svc.CreateCustomer(ctx, customer.CreateCustomerCmd{
		FirstName: "Temp",
		LastName:  "User",
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	err = svc.DeleteCustomer(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to delete customer: %v", err)
	}

	// Should not be retrieved after soft delete
	_, err = svc.GetCustomer(ctx, created.ID)
	if err == nil {
		t.Error("expected error retrieving soft-deleted customer, got nil")
	}
}
