package payment

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/organization"
	"github.com/appointly/appointly/backend/internal/domain/payment"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	paymentuc "github.com/appointly/appointly/backend/internal/usecase/payment"
)

// Handler handles payment-related HTTP requests.
type Handler struct {
	paymentService *paymentuc.Service
	orgRepo        organization.Repository
}

// NewHandler constructs a new Handler.
func NewHandler(svc *paymentuc.Service, orgRepo organization.Repository) *Handler {
	return &Handler{
		paymentService: svc,
		orgRepo:        orgRepo,
	}
}

// CreateIntentRequest defines payload for initiating payment.
type CreateIntentRequest struct {
	AppointmentID uuid.UUID        `json:"appointment_id"`
	Provider      payment.Provider `json:"provider,omitempty"`
	SuccessURL    string           `json:"success_url,omitempty"`
	CancelURL     string           `json:"cancel_url,omitempty"`
}

// CreatePublicPaymentIntent handles POST /api/v1/public/orgs/{slug}/payments/intent
func (h *Handler) CreatePublicPaymentIntent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	org, err := h.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		v1.RespondError(w, r, apperror.NotFound("organization not found"))
		return
	}

	var req CreateIntentRequest
	if appErr := v1.DecodeJSON(r, &req); appErr != nil {
		v1.RespondError(w, r, appErr)
		return
	}

	if req.AppointmentID == uuid.Nil {
		v1.RespondError(w, r, apperror.BadRequest("appointment_id is required"))
		return
	}

	cmd := paymentuc.CreateIntentCmd{
		OrganizationID: org.ID,
		AppointmentID:  req.AppointmentID,
		Provider:       req.Provider,
		SuccessURL:     req.SuccessURL,
		CancelURL:      req.CancelURL,
	}

	payRecord, createErr := h.paymentService.CreatePaymentIntent(ctx, cmd)
	if createErr != nil {
		v1.RespondError(w, r, createErr)
		return
	}

	v1.RespondCreated(w, r, payRecord)
}

// HandleWebhook handles POST /api/v1/webhooks/payments/{provider}
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	providerStr := chi.URLParam(r, "provider")
	if providerStr == "" {
		providerStr = "manual"
	}
	prov := payment.Provider(providerStr)

	// Read signature from standard headers
	sig := r.Header.Get("X-Webhook-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Signature")
	}

	// Read raw body for HMAC verification
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("failed to read webhook body"))
		return
	}

	if handleErr := h.paymentService.HandleWebhook(r.Context(), prov, bodyBytes, sig); handleErr != nil {
		v1.RespondError(w, r, handleErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// ListPayments handles GET /api/v1/payments
func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, perPage := getPagination(r)

	payments, total, err := h.paymentService.ListPayments(ctx, uuid.Nil, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetPaymentByID handles GET /api/v1/payments/{id}
func (h *Handler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	payID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid payment ID"))
		return
	}

	payRecord, getErr := h.paymentService.GetPaymentByID(ctx, uuid.Nil, payID)
	if getErr != nil {
		v1.RespondError(w, r, getErr)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, payRecord)
}

// RefundRequest defines refund creation body.
type RefundRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Reason      string `json:"reason"`
}

// ProcessRefund handles POST /api/v1/payments/{id}/refund
func (h *Handler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	payID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid payment ID"))
		return
	}

	var req RefundRequest
	if appErr := v1.DecodeJSON(r, &req); appErr != nil {
		v1.RespondError(w, r, appErr)
		return
	}

	cmd := payment.CreateRefundCmd{
		PaymentID:   payID,
		AmountCents: req.AmountCents,
		Reason:      req.Reason,
	}

	ref, refErr := h.paymentService.ProcessRefund(ctx, cmd)
	if refErr != nil {
		v1.RespondError(w, r, refErr)
		return
	}

	v1.RespondCreated(w, r, ref)
}

// RegisterRoutes registers payment routes under Chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/payments", func(r chi.Router) {
		r.Get("/", h.ListPayments)
		r.Get("/{id}", h.GetPaymentByID)
		r.Post("/{id}/refund", h.ProcessRefund)
	})

	r.Route("/webhooks/payments/{provider}", func(r chi.Router) {
		r.Post("/", h.HandleWebhook)
	})

	r.Route("/public/orgs/{slug}/payments", func(r chi.Router) {
		r.Post("/intent", h.CreatePublicPaymentIntent)
	})
}

func getPagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
