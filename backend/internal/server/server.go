// Package server sets up the HTTP server, all middleware, and routes.
// This is the transport layer entry point — it wires together config,
// middleware, and handlers without containing any business logic.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/appointly/appointly/backend/internal/config"
	"github.com/appointly/appointly/backend/internal/handler/v1/public"
	"github.com/appointly/appointly/backend/internal/handler/v1/subscription"
	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/logger"
)

// Server wraps the HTTP server with its dependencies.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	log        *logger.Logger
}

// Dependencies holds all the injected dependencies for the server.
// This struct grows as we add more handlers/use cases.
type Dependencies struct {
	Config              *config.Config
	Logger              *logger.Logger
	PublicHandler       *public.PublicHandler
	SubscriptionHandler *subscription.Handler
}

// New creates a new Server with the given dependencies.
func New(deps Dependencies) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg: deps.Config,
		log: deps.Logger,
	}

	s.setupMiddleware(r, deps)
	s.setupRoutes(r, deps)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", deps.Config.API.Port),
		Handler:      r,
		ReadTimeout:  deps.Config.API.ReadTimeout,
		WriteTimeout: deps.Config.API.WriteTimeout,
		IdleTimeout:  deps.Config.API.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(deps.Logger.Handler(), slog.LevelError),
	}

	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	s.log.Info("starting HTTP server", "addr", s.httpServer.Addr, "env", s.cfg.App.Env)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware configures the global middleware stack.
func (s *Server) setupMiddleware(r *chi.Mux, deps Dependencies) {
	// Request ID — must be first so all subsequent middleware can use it
	r.Use(middleware.RequestID)

	// Structured request logging
	r.Use(middleware.RequestLogger(deps.Logger))

	// Recovery — catch panics and return 500
	r.Use(chimiddleware.Recoverer)

	// Real IP extraction (trusts X-Forwarded-For behind known proxies)
	r.Use(chimiddleware.RealIP)

	// CORS
	r.Use(middleware.CORS(deps.Config.CORS))

	// Content type negotiation
	r.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

	// Timeout per request
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// Compression
	r.Use(chimiddleware.Compress(5, "application/json"))
}

// setupRoutes wires all HTTP routes.
func (s *Server) setupRoutes(r *chi.Mux, deps Dependencies) {
	// Health and readiness endpoints (no auth, no rate-limiting)
	r.Get("/health", s.handleHealth)
	r.Get("/ready", s.handleReady)

	// API v1 — authenticated routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply rate limiting for API routes
		r.Use(middleware.RateLimit(deps.Config))

		// Auth endpoints (stricter rate limiting applied per-handler)
		r.Route("/auth", func(r chi.Router) {
			// TODO: mount AuthHandler routes
			// r.Post("/register", deps.AuthHandler.Register)
			// r.Post("/login", deps.AuthHandler.Login)
			// r.Post("/logout", deps.AuthHandler.Logout)
			// r.Post("/refresh", deps.AuthHandler.Refresh)
			// r.Post("/forgot-password", deps.AuthHandler.ForgotPassword)
			// r.Post("/reset-password", deps.AuthHandler.ResetPassword)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			// TODO: mount auth middleware
			// r.Use(middleware.Authenticate(deps.TokenService))

			// Organization
			r.Route("/organizations", func(r chi.Router) {
				// TODO: org routes
			})

			// Appointments
			r.Route("/appointments", func(r chi.Router) {
				// TODO: appointment routes
			})

			// Staff
			r.Route("/staff", func(r chi.Router) {
				// TODO: staff routes
			})

			// Customers
			r.Route("/customers", func(r chi.Router) {
				// TODO: customer routes
			})

			// Services
			r.Route("/services", func(r chi.Router) {
				// TODO: service routes
			})

			// Locations
			r.Route("/locations", func(r chi.Router) {
				// TODO: location routes
			})
		})

		// Public booking API & SaaS Subscriptions
		r.Group(func(r chi.Router) {
			r.Use(middleware.PublicRateLimit(deps.Config))
			if deps.PublicHandler != nil {
				deps.PublicHandler.RegisterRoutes(r)
			}
			if deps.SubscriptionHandler != nil {
				deps.SubscriptionHandler.RegisterRoutes(r)
			}
		})
	})

	// API documentation (development only)
	if deps.Config.IsDevelopment() {
		r.Get("/docs", s.handleDocs)
		r.Get("/openapi.yaml", s.handleOpenAPISpec)
	}
}

// --- Health Handlers ---------------------------------------------------------

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","version":%q,"time":%q}`,
		s.cfg.App.Version,
		time.Now().UTC().Format(time.RFC3339),
	)
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// TODO: check DB and Redis connectivity
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ready"}`)
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Appointly API Docs</title>
<meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1">
<link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
<style>body { margin: 0; padding: 0; }</style>
</head><body>
<redoc spec-url="/openapi.yaml"></redoc>
<script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"></script>
</body></html>`)
}

func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "api/openapi.yaml")
}
