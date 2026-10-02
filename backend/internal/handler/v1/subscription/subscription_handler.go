package subscription

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/rbac"
	"github.com/appointly/appointly/backend/internal/domain/subscription"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	subuc "github.com/appointly/appointly/backend/internal/usecase/subscription"
)

// Handler handles SaaS subscription and plan HTTP endpoints.
type Handler struct {
	subService *subuc.Service
}

// NewHandler constructs a new Subscription Handler.
func NewHandler(subService *subuc.Service) *Handler {
	return &Handler{subService: subService}
}

// ListPlans handles GET /api/v1/plans
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subService.GetPlans(r.Context())
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}
	v1.RespondJSON(w, r, http.StatusOK, plans)
}

// GetCurrentSubscription handles GET /api/v1/subscriptions/current
func (h *Handler) GetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authCtx := rbac.FromContext(ctx)
	orgID := uuid.Nil
	if authCtx != nil {
		orgID = authCtx.OrgID
	}

	sub, metrics, err := h.subService.GetSubscription(ctx, orgID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"subscription": sub,
		"usage":        metrics,
	})
}

// CheckoutRequest defines plan change payload.
type CheckoutRequest struct {
	PlanSlug      string                     `json:"plan_slug"`
	BillingPeriod subscription.BillingPeriod `json:"billing_period,omitempty"`
}

// CheckoutSubscription handles POST /api/v1/subscriptions/checkout
func (h *Handler) CheckoutSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil && !authCtx.IsOwner {
		v1.RespondError(w, r, apperror.Unauthorized("only organization owner can change subscription plan"))
		return
	}

	orgID := uuid.Nil
	if authCtx != nil {
		orgID = authCtx.OrgID
	}

	var req CheckoutRequest
	if appErr := v1.DecodeJSON(r, &req); appErr != nil {
		v1.RespondError(w, r, appErr)
		return
	}

	if req.PlanSlug == "" {
		v1.RespondError(w, r, apperror.BadRequest("plan_slug is required"))
		return
	}

	if req.BillingPeriod == "" {
		req.BillingPeriod = subscription.BillingPeriodMonthly
	}

	updatedSub, err := h.subService.UpgradePlan(ctx, orgID, req.PlanSlug, req.BillingPeriod)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, updatedSub)
}

// CancelSubscription handles POST /api/v1/subscriptions/cancel
func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authCtx := rbac.FromContext(ctx)
	if authCtx != nil && !authCtx.IsOwner {
		v1.RespondError(w, r, apperror.Unauthorized("only organization owner can cancel subscription"))
		return
	}

	orgID := uuid.Nil
	if authCtx != nil {
		orgID = authCtx.OrgID
	}

	if err := h.subService.CancelSubscription(ctx, orgID); err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, map[string]string{"status": "cancelled"})
}

// HandleSaaSHook handles POST /api/v1/webhooks/saas-subscriptions/{provider}
func (h *Handler) HandleSaaSHook(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Webhook-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Signature")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("failed to read webhook body"))
		return
	}

	if hookErr := h.subService.HandleSaaSHook(r.Context(), bodyBytes, sig); hookErr != nil {
		v1.RespondError(w, r, hookErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// RegisterRoutes mounts subscription routes under Chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/plans", func(r chi.Router) {
		r.Get("/", h.ListPlans)
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/current", h.GetCurrentSubscription)
		r.Post("/checkout", h.CheckoutSubscription)
		r.Post("/cancel", h.CancelSubscription)
	})

	r.Route("/webhooks/saas-subscriptions/{provider}", func(r chi.Router) {
		r.Post("/", h.HandleSaaSHook)
	})
}
