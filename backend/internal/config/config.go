// Package config loads and validates application configuration from environment
// variables. All configuration is loaded once at startup; no global singletons
// are used — config is passed explicitly via dependency injection.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the root configuration struct.
type Config struct {
	App      AppConfig
	API      APIConfig
	DB       DBConfig
	Redis    RedisConfig
	Storage  StorageConfig
	Auth     AuthConfig
	Email    EmailConfig
	WhatsApp WhatsAppConfig
	Payment  PaymentConfig
	Worker   WorkerConfig
	Log      LogConfig
	CORS     CORSConfig
	Feature  FeatureConfig
	Migrate  MigrateConfig
}

type AppConfig struct {
	Env         string
	Name        string
	Version     string
	BaseURL     string
	FrontendURL string
}

type APIConfig struct {
	Port             int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ShutdownTimeout  time.Duration
}

type DBConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	URL             string // overrides individual params if set
}

// DSN returns the PostgreSQL connection string.
func (d DBConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
}

type RedisConfig struct {
	URL          string
	Password     string
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type StorageConfig struct {
	Type          string // minio | s3
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	Region        string
	UseSSL        bool
	PublicBaseURL string
}

type AuthConfig struct {
	JWTSecret            string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	CookieDomain         string
	CookieSecure         bool
	CookieSameSite       string
	CSRFSecret           string
	CSRFSecure           bool
	Argon2Memory         uint32
	Argon2Iterations     uint32
	Argon2Parallelism    uint8
}

type EmailConfig struct {
	Provider    string // smtp | sendgrid | ses | resend
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	FromName    string
	FromEmail   string
	SMTPTLS     bool
	SendGridKey string
	ResendKey   string
}

type WhatsAppConfig struct {
	Provider   string
	APIKey     string
	FromNumber string
}

type PaymentConfig struct {
	Provider        string // stripe | midtrans | xendit
	StripeSecretKey string
	StripeWebhookSecret string
	MidtransServerKey   string
	MidtransEnv         string
	XenditSecretKey     string
}

type WorkerConfig struct {
	Concurrency  int
	Queue        string
	PollInterval time.Duration
}

type LogConfig struct {
	Level     slog.Level
	Format    string // text | json
	AddSource bool
}

type CORSConfig struct {
	AllowedOrigins     []string
	AllowCredentials   bool
}

type FeatureConfig struct {
	OnlinePayment         bool
	WhatsAppNotification  bool
	MultiLocation         bool
	ResourceBooking       bool
}

type MigrateConfig struct {
	Path string
}

// Load reads all configuration from environment variables.
// Returns an error if any required variable is missing or invalid.
func Load() (*Config, error) {
	cfg := &Config{}

	// App
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Name = getEnv("APP_NAME", "appointly")
	cfg.App.Version = getEnv("APP_VERSION", "0.0.0")
	cfg.App.BaseURL = getEnv("APP_BASE_URL", "http://localhost:8080")
	cfg.App.FrontendURL = getEnv("APP_FRONTEND_URL", "http://localhost:4321")

	// API
	cfg.API.Port = getEnvInt("API_PORT", 8080)
	cfg.API.ReadTimeout = getEnvDuration("API_READ_TIMEOUT", 30*time.Second)
	cfg.API.WriteTimeout = getEnvDuration("API_WRITE_TIMEOUT", 30*time.Second)
	cfg.API.IdleTimeout = getEnvDuration("API_IDLE_TIMEOUT", 120*time.Second)
	cfg.API.ShutdownTimeout = getEnvDuration("API_SHUTDOWN_TIMEOUT", 30*time.Second)

	// Database
	cfg.DB.URL = getEnv("DATABASE_URL", "")
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnvInt("DB_PORT", 5432)
	cfg.DB.Name = getEnv("DB_NAME", "appointly_dev")
	cfg.DB.User = getEnv("DB_USER", "appointly")
	cfg.DB.Password = getEnv("DB_PASSWORD", "")
	cfg.DB.SSLMode = getEnv("DB_SSL_MODE", "disable")
	cfg.DB.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	cfg.DB.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 10)
	cfg.DB.ConnMaxLifetime = getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	cfg.DB.ConnMaxIdleTime = getEnvDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute)

	// Redis
	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379/0")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.MaxRetries = getEnvInt("REDIS_MAX_RETRIES", 3)
	cfg.Redis.DialTimeout = getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	cfg.Redis.ReadTimeout = getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	cfg.Redis.WriteTimeout = getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)

	// Storage
	cfg.Storage.Type = getEnv("STORAGE_TYPE", "minio")
	cfg.Storage.Endpoint = getEnv("STORAGE_ENDPOINT", "localhost:9000")
	cfg.Storage.AccessKey = getEnv("STORAGE_ACCESS_KEY", "minioadmin")
	cfg.Storage.SecretKey = getEnv("STORAGE_SECRET_KEY", "minioadmin")
	cfg.Storage.Bucket = getEnv("STORAGE_BUCKET", "appointly")
	cfg.Storage.Region = getEnv("STORAGE_REGION", "us-east-1")
	cfg.Storage.UseSSL = getEnvBool("STORAGE_USE_SSL", false)
	cfg.Storage.PublicBaseURL = getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:9000/appointly")

	// Auth
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	cfg.Auth.JWTSecret = jwtSecret
	cfg.Auth.AccessTokenTTL = getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute)
	cfg.Auth.RefreshTokenTTL = getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	cfg.Auth.CookieDomain = getEnv("COOKIE_DOMAIN", "localhost")
	cfg.Auth.CookieSecure = getEnvBool("COOKIE_SECURE", false)
	cfg.Auth.CookieSameSite = getEnv("COOKIE_SAME_SITE", "Lax")
	cfg.Auth.CSRFSecret = getEnv("CSRF_SECRET", "")
	cfg.Auth.CSRFSecure = getEnvBool("CSRF_SECURE", false)
	cfg.Auth.Argon2Memory = uint32(getEnvInt("ARGON2_MEMORY", 65536))
	cfg.Auth.Argon2Iterations = uint32(getEnvInt("ARGON2_ITERATIONS", 3))
	cfg.Auth.Argon2Parallelism = uint8(getEnvInt("ARGON2_PARALLELISM", 4))

	// Email
	cfg.Email.Provider = getEnv("EMAIL_PROVIDER", "smtp")
	cfg.Email.SMTPHost = getEnv("SMTP_HOST", "localhost")
	cfg.Email.SMTPPort = getEnvInt("SMTP_PORT", 1025)
	cfg.Email.SMTPUser = getEnv("SMTP_USER", "")
	cfg.Email.SMTPPass = getEnv("SMTP_PASSWORD", "")
	cfg.Email.FromName = getEnv("SMTP_FROM_NAME", "Appointly")
	cfg.Email.FromEmail = getEnv("SMTP_FROM_EMAIL", "noreply@appointly.dev")
	cfg.Email.SMTPTLS = getEnvBool("SMTP_TLS", false)
	cfg.Email.SendGridKey = getEnv("SENDGRID_API_KEY", "")
	cfg.Email.ResendKey = getEnv("RESEND_API_KEY", "")

	// WhatsApp
	cfg.WhatsApp.Provider = getEnv("WHATSAPP_PROVIDER", "")
	cfg.WhatsApp.APIKey = getEnv("WHATSAPP_API_KEY", "")
	cfg.WhatsApp.FromNumber = getEnv("WHATSAPP_FROM_NUMBER", "")

	// Payment
	cfg.Payment.Provider = getEnv("PAYMENT_PROVIDER", "stripe")
	cfg.Payment.StripeSecretKey = getEnv("STRIPE_SECRET_KEY", "")
	cfg.Payment.StripeWebhookSecret = getEnv("STRIPE_WEBHOOK_SECRET", "")
	cfg.Payment.MidtransServerKey = getEnv("MIDTRANS_SERVER_KEY", "")
	cfg.Payment.MidtransEnv = getEnv("MIDTRANS_ENV", "sandbox")
	cfg.Payment.XenditSecretKey = getEnv("XENDIT_SECRET_KEY", "")

	// Worker
	cfg.Worker.Concurrency = getEnvInt("WORKER_CONCURRENCY", 10)
	cfg.Worker.Queue = getEnv("WORKER_QUEUE", "default")
	cfg.Worker.PollInterval = getEnvDuration("WORKER_POLL_INTERVAL", time.Second)

	// Log
	cfg.Log.Level = parseLogLevel(getEnv("LOG_LEVEL", "debug"))
	cfg.Log.Format = getEnv("LOG_FORMAT", "text")
	cfg.Log.AddSource = getEnvBool("LOG_ADD_SOURCE", false)

	// CORS
	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:4321,http://localhost:3000")
	cfg.CORS.AllowedOrigins = splitAndTrim(originsStr, ",")
	cfg.CORS.AllowCredentials = getEnvBool("CORS_ALLOW_CREDENTIALS", true)

	// Features
	cfg.Feature.OnlinePayment = getEnvBool("FEATURE_ONLINE_PAYMENT", false)
	cfg.Feature.WhatsAppNotification = getEnvBool("FEATURE_WHATSAPP_NOTIFICATION", false)
	cfg.Feature.MultiLocation = getEnvBool("FEATURE_MULTI_LOCATION", true)
	cfg.Feature.ResourceBooking = getEnvBool("FEATURE_RESOURCE_BOOKING", true)

	// Migrations
	cfg.Migrate.Path = getEnv("MIGRATE_PATH", "db/migrations")

	return cfg, nil
}

// IsProduction returns true if APP_ENV is "production".
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsDevelopment returns true if APP_ENV is "development".
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// --- Helpers -----------------------------------------------------------------

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return i
}

func getEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultValue
	}
	return b
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
