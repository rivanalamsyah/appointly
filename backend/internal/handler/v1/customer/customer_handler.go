package customer

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/customer"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	"github.com/appointly/appointly/backend/internal/pkg/pagination"
	usecase "github.com/appointly/appointly/backend/internal/usecase/customer"
)

type CustomerHandler struct {
	customerSvc *usecase.Service
}

func NewCustomerHandler(customerSvc *usecase.Service) *CustomerHandler {
	return &CustomerHandler{
		customerSvc: customerSvc,
	}
}

type CreateCustomerRequest struct {
	FirstName string                  `json:"first_name"`
	LastName  string                  `json:"last_name"`
	Email     string                  `json:"email,omitempty"`
	Phone     string                  `json:"phone,omitempty"`
	Notes     string                  `json:"notes,omitempty"`
	Status    customer.CustomerStatus `json:"status,omitempty"`
	Source    customer.CustomerSource `json:"source,omitempty"`
	Tags      []string                `json:"tags,omitempty"`
}

type UpdateCustomerRequest struct {
	FirstName *string                  `json:"first_name,omitempty"`
	LastName  *string                  `json:"last_name,omitempty"`
	Email     *string                  `json:"email,omitempty"`
	Phone     *string                  `json:"phone,omitempty"`
	Notes     *string                  `json:"notes,omitempty"`
	Status    *customer.CustomerStatus `json:"status,omitempty"`
	Source    *customer.CustomerSource `json:"source,omitempty"`
	Tags      []string                 `json:"tags,omitempty"`
}

type LinkUserRequest struct {
	UserID string `json:"user_id"`
}

type AddNoteRequest struct {
	Content string `json:"content"`
}

func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	filter := customer.ListCustomersFilter{
		OrganizationID: orgID,
		Search:         r.URL.Query().Get("search"),
		SortBy:         r.URL.Query().Get("sort_by"),
		SortOrder:      r.URL.Query().Get("sort_order"),
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		st := customer.CustomerStatus(statusStr)
		filter.Status = &st
	}

	if sourceStr := r.URL.Query().Get("source"); sourceStr != "" {
		src := customer.CustomerSource(sourceStr)
		filter.Source = &src
	}

	customers, total, err := h.customerSvc.ListCustomers(r.Context(), filter, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	paginationMeta := pagination.NewOffsetMeta(pagination.OffsetParams{Page: page, PerPage: perPage}, total)
	v1.RespondList(w, r, http.StatusOK, customers, &paginationMeta)
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid orgID"))
		return
	}

	var req CreateCustomerRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := customer.CreateCustomerCmd{
		OrganizationID: orgID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		Notes:          req.Notes,
		Status:         req.Status,
		Source:         req.Source,
		Tags:           req.Tags,
	}

	created, err := h.customerSvc.CreateCustomer(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, created)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	cust, err := h.customerSvc.GetCustomer(r.Context(), custID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, cust)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	var req UpdateCustomerRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	cmd := customer.UpdateCustomerCmd{
		ID:        custID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Notes:     req.Notes,
		Status:    req.Status,
		Source:    req.Source,
		Tags:      req.Tags,
	}

	updated, err := h.customerSvc.UpdateCustomer(r.Context(), cmd)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, updated)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	if err := h.customerSvc.DeleteCustomer(r.Context(), custID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"message": "customer deleted successfully"})
}

func (h *CustomerHandler) LinkUser(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	var req LinkUserRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid user_id"))
		return
	}

	linked, err := h.customerSvc.LinkUser(r.Context(), custID, targetUserID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, linked)
}

func (h *CustomerHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	var req AddNoteRequest
	if err := v1.DecodeJSON(r, &req); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	note, err := h.customerSvc.AddNote(r.Context(), custID, req.Content)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondCreated(w, r, note)
}

func (h *CustomerHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	custIDStr := chi.URLParam(r, "customerID")
	custID, err := uuid.Parse(custIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid customerID"))
		return
	}

	notes, err := h.customerSvc.ListNotes(r.Context(), custID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, notes)
}

// Helper registration for Chi router
func (h *CustomerHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/organizations/{orgID}/customers", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.ListCustomers)
		r.Post("/", h.CreateCustomer)
		r.Get("/{customerID}", h.GetCustomer)
		r.Put("/{customerID}", h.UpdateCustomer)
		r.Delete("/{customerID}", h.DeleteCustomer)
		r.Post("/{customerID}/link-user", h.LinkUser)
		r.Get("/{customerID}/notes", h.ListNotes)
		r.Post("/{customerID}/notes", h.AddNote)
	})
}
