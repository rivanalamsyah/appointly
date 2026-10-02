package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

type CustomerRepository struct {
	mu        sync.RWMutex
	customers map[uuid.UUID]*customer.Customer
	notes     map[uuid.UUID][]*customer.CustomerNote // customerID -> []note
}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		customers: make(map[uuid.UUID]*customer.Customer),
		notes:     make(map[uuid.UUID][]*customer.CustomerNote),
	}
}

func (r *CustomerRepository) Create(ctx context.Context, cmd customer.CreateCustomerCmd) (*customer.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	firstName := strings.TrimSpace(cmd.FirstName)
	if firstName == "" {
		return nil, apperror.ValidationFailed("First name is required")
	}

	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	phone := strings.TrimSpace(cmd.Phone)

	// Check duplicate email in same organization
	if email != "" {
		for _, c := range r.customers {
			if c.OrganizationID == cmd.OrganizationID && !c.IsDeleted() && strings.ToLower(c.Email) == email {
				return nil, apperror.Conflict("Customer with this email already exists in organization")
			}
		}
	}

	// Check duplicate phone in same organization
	if phone != "" {
		for _, c := range r.customers {
			if c.OrganizationID == cmd.OrganizationID && !c.IsDeleted() && c.Phone == phone {
				return nil, apperror.Conflict("Customer with this phone number already exists in organization")
			}
		}
	}

	status := cmd.Status
	if status == "" {
		status = customer.StatusActive
	}

	source := cmd.Source
	if source == "" {
		source = customer.SourceManual
	}

	now := time.Now()
	custID := uuid.New()

	cust := &customer.Customer{
		ID:             custID,
		OrganizationID: cmd.OrganizationID,
		UserID:         cmd.UserID,
		FirstName:      firstName,
		LastName:       strings.TrimSpace(cmd.LastName),
		Email:          email,
		Phone:          phone,
		Notes:          strings.TrimSpace(cmd.Notes),
		Status:         status,
		Source:         source,
		Tags:           cmd.Tags,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.customers[custID] = cust
	cp := *cust
	return &cp, nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, orgID, customerID uuid.UUID) (*customer.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cust, exists := r.customers[customerID]
	if !exists || cust.OrganizationID != orgID || cust.IsDeleted() {
		return nil, apperror.NotFound("customer")
	}

	cp := *cust
	return &cp, nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (*customer.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	targetEmail := strings.ToLower(strings.TrimSpace(email))
	if targetEmail == "" {
		return nil, apperror.NotFound("customer")
	}

	for _, c := range r.customers {
		if c.OrganizationID == orgID && !c.IsDeleted() && strings.ToLower(c.Email) == targetEmail {
			cp := *c
			return &cp, nil
		}
	}

	return nil, apperror.NotFound("customer")
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, orgID uuid.UUID, phone string) (*customer.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	targetPhone := strings.TrimSpace(phone)
	if targetPhone == "" {
		return nil, apperror.NotFound("customer")
	}

	for _, c := range r.customers {
		if c.OrganizationID == orgID && !c.IsDeleted() && c.Phone == targetPhone {
			cp := *c
			return &cp, nil
		}
	}

	return nil, apperror.NotFound("customer")
}

func (r *CustomerRepository) GetByUserID(ctx context.Context, orgID, userID uuid.UUID) (*customer.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.customers {
		if c.OrganizationID == orgID && !c.IsDeleted() && c.UserID != nil && *c.UserID == userID {
			cp := *c
			return &cp, nil
		}
	}

	return nil, apperror.NotFound("customer")
}

func (r *CustomerRepository) List(ctx context.Context, filter customer.ListCustomersFilter, page, perPage int) ([]*customer.Customer, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var matched []*customer.Customer
	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, c := range r.customers {
		if c.OrganizationID != filter.OrganizationID {
			continue
		}
		if !filter.IncludeDeleted && c.IsDeleted() {
			continue
		}
		if filter.Status != nil && c.Status != *filter.Status {
			continue
		}
		if filter.Source != nil && c.Source != *filter.Source {
			continue
		}
		if searchLower != "" {
			fullName := strings.ToLower(c.FullName())
			matchName := strings.Contains(fullName, searchLower)
			matchEmail := strings.Contains(strings.ToLower(c.Email), searchLower)
			matchPhone := strings.Contains(c.Phone, searchLower)
			if !matchName && !matchEmail && !matchPhone {
				continue
			}
		}

		cp := *c
		matched = append(matched, &cp)
	}

	total := len(matched)

	// Sorting
	sort.Slice(matched, func(i, j int) bool {
		if filter.SortBy == "name" {
			if filter.SortOrder == "asc" {
				return matched[i].FullName() < matched[j].FullName()
			}
			return matched[i].FullName() > matched[j].FullName()
		}
		if filter.SortBy == "total_appointments" {
			if filter.SortOrder == "asc" {
				return matched[i].TotalAppointments < matched[j].TotalAppointments
			}
			return matched[i].TotalAppointments > matched[j].TotalAppointments
		}

		// Default sort by created_at desc
		if filter.SortOrder == "asc" {
			return matched[i].CreatedAt.Before(matched[j].CreatedAt)
		}
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	// Pagination bounds
	start := (page - 1) * perPage
	if start >= total {
		return []*customer.Customer{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *CustomerRepository) Update(ctx context.Context, cmd customer.UpdateCustomerCmd) (*customer.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, exists := r.customers[cmd.ID]
	if !exists || c.OrganizationID != cmd.OrganizationID || c.IsDeleted() {
		return nil, apperror.NotFound("customer")
	}

	if cmd.FirstName != nil {
		fname := strings.TrimSpace(*cmd.FirstName)
		if fname == "" {
			return nil, apperror.ValidationFailed("First name cannot be empty")
		}
		c.FirstName = fname
	}
	if cmd.LastName != nil {
		c.LastName = strings.TrimSpace(*cmd.LastName)
	}
	if cmd.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*cmd.Email))
		if email != "" && email != c.Email {
			// check conflict
			for _, existing := range r.customers {
				if existing.ID != c.ID && existing.OrganizationID == c.OrganizationID && !existing.IsDeleted() && strings.ToLower(existing.Email) == email {
					return nil, apperror.Conflict("Email already used by another customer")
				}
			}
		}
		c.Email = email
	}
	if cmd.Phone != nil {
		phone := strings.TrimSpace(*cmd.Phone)
		if phone != "" && phone != c.Phone {
			// check conflict
			for _, existing := range r.customers {
				if existing.ID != c.ID && existing.OrganizationID == c.OrganizationID && !existing.IsDeleted() && existing.Phone == phone {
					return nil, apperror.Conflict("Phone number already used by another customer")
				}
			}
		}
		c.Phone = phone
	}
	if cmd.Notes != nil {
		c.Notes = strings.TrimSpace(*cmd.Notes)
	}
	if cmd.Status != nil {
		c.Status = *cmd.Status
	}
	if cmd.Source != nil {
		c.Source = *cmd.Source
	}
	if cmd.Tags != nil {
		c.Tags = cmd.Tags
	}

	c.UpdatedAt = time.Now()
	cp := *c
	return &cp, nil
}

func (r *CustomerRepository) SoftDelete(ctx context.Context, orgID, customerID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, exists := r.customers[customerID]
	if !exists || c.OrganizationID != orgID || c.IsDeleted() {
		return apperror.NotFound("customer")
	}

	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
	return nil
}

func (r *CustomerRepository) LinkUser(ctx context.Context, cmd customer.LinkUserCmd) (*customer.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, exists := r.customers[cmd.CustomerID]
	if !exists || c.OrganizationID != cmd.OrganizationID || c.IsDeleted() {
		return nil, apperror.NotFound("customer")
	}

	userID := cmd.UserID
	c.UserID = &userID
	c.UpdatedAt = time.Now()

	cp := *c
	return &cp, nil
}

func (r *CustomerRepository) AddNote(ctx context.Context, cmd customer.AddNoteCmd) (*customer.CustomerNote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, exists := r.customers[cmd.CustomerID]
	if !exists || c.OrganizationID != cmd.OrganizationID || c.IsDeleted() {
		return nil, apperror.NotFound("customer")
	}

	note := &customer.CustomerNote{
		ID:             uuid.New(),
		OrganizationID: cmd.OrganizationID,
		CustomerID:     cmd.CustomerID,
		AuthorID:       cmd.AuthorID,
		AuthorName:     strings.TrimSpace(cmd.AuthorName),
		Content:        strings.TrimSpace(cmd.Content),
		CreatedAt:      time.Now(),
	}

	r.notes[cmd.CustomerID] = append(r.notes[cmd.CustomerID], note)
	return note, nil
}

func (r *CustomerRepository) ListNotes(ctx context.Context, orgID, customerID uuid.UUID) ([]*customer.CustomerNote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, exists := r.customers[customerID]
	if !exists || c.OrganizationID != orgID || c.IsDeleted() {
		return nil, apperror.NotFound("customer")
	}

	notes, exists := r.notes[customerID]
	if !exists {
		return []*customer.CustomerNote{}, nil
	}

	var result []*customer.CustomerNote
	for _, n := range notes {
		cp := *n
		result = append(result, &cp)
	}

	return result, nil
}

func (r *CustomerRepository) FindOrCreate(ctx context.Context, cmd customer.CreateCustomerCmd) (*customer.Customer, bool, error) {
	// Try email first
	if cmd.Email != "" {
		cust, err := r.GetByEmail(ctx, cmd.OrganizationID, cmd.Email)
		if err == nil {
			return cust, false, nil
		}
	}

	// Try phone next
	if cmd.Phone != "" {
		cust, err := r.GetByPhone(ctx, cmd.OrganizationID, cmd.Phone)
		if err == nil {
			return cust, false, nil
		}
	}

	// Create new customer
	created, err := r.Create(ctx, cmd)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}
