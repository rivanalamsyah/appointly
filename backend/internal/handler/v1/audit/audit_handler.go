package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/appointly/appointly/backend/internal/domain/audit"
	v1 "github.com/appointly/appointly/backend/internal/handler/v1"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	usecase "github.com/appointly/appointly/backend/internal/usecase/audit"
)

type Handler struct {
	svc *usecase.Service
}

func NewHandler(svc *usecase.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/audit-logs", func(r chi.Router) {
		r.Get("/", h.ListLogs)
		r.Get("/{id}", h.GetLogByID)
	})
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	filter := audit.ListFilter{}
	if actionStr := r.URL.Query().Get("action"); actionStr != "" {
		act := audit.Action(actionStr)
		filter.Action = &act
	}

	logs, total, err := h.svc.ListLogs(r.Context(), orgID, filter, page, perPage)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	meta := v1.Meta{
		RequestID: middleware.RequestIDFromContext(r.Context()),
	}
	v1.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"items": logs,
		"total": total,
		"meta":  meta,
	})
}

func (h *Handler) GetLogByID(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.OrgIDFromContext(r.Context())
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		v1.RespondError(w, r, apperror.Unauthorized("organization context required"))
		return
	}

	idStr := chi.URLParam(r, "id")
	logID, err := uuid.Parse(idStr)
	if err != nil {
		v1.RespondError(w, r, apperror.BadRequest("invalid audit log ID"))
		return
	}

	entry, err := h.svc.GetLogByID(r.Context(), orgID, logID)
	if err != nil {
		v1.RespondError(w, r, err)
		return
	}

	v1.RespondJSON(w, r, http.StatusOK, entry)
}
