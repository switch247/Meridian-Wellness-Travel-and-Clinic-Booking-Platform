package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	EncryptionKey        string
	LockoutThreshold     int
	LockoutDuration      time.Duration
	TokenTTL             time.Duration
	ReservationHold      time.Duration
	SlotGranularity      int
	AllowedPostalCode    []string
	TLSEnabled           bool
	TLSCertFile          string
	TLSKeyFile           string
	AllowedIPs           []string
	TrustProxyHeaders    bool
	CORSAllowedOrigins   []string
	LogFilePath          string
	LogMaxBytes          int64
	LogMaxBackups        int
	ReportWorkerInterval time.Duration
}

func Load() Config {
	return Config{
		Port:                 getEnv("PORT", "8443"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/meridian?sslmode=disable"),
		JWTSecret:            strings.TrimSpace(os.Getenv("JWT_SECRET")),
		EncryptionKey:        strings.TrimSpace(os.Getenv("ENCRYPTION_KEY")),
		LockoutThreshold:     getEnvInt("LOCKOUT_THRESHOLD", 5),
		LockoutDuration:      getEnvDuration("LOCKOUT_DURATION", 15*time.Minute),
		TokenTTL:             getEnvDuration("TOKEN_TTL", 8*time.Hour),
		ReservationHold:      getEnvDuration("RESERVATION_HOLD", 10*time.Minute),
		SlotGranularity:      getEnvInt("SLOT_GRANULARITY_MINUTES", 15),
		AllowedPostalCode:    getEnvCSV("ALLOWED_POSTAL_CODES", []string{"10001", "10002", "10003", "60601", "90001"}),
		TLSEnabled:           getEnvBool("TLS_ENABLED", true),
		TLSCertFile:          getEnv("TLS_CERT_FILE", "/certs/server.crt"),
		TLSKeyFile:           getEnv("TLS_KEY_FILE", "/certs/server.key"),
		AllowedIPs:           getEnvCSV("ALLOWED_IPS", []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}),
		TrustProxyHeaders:    getEnvBool("TRUST_PROXY_HEADERS", false),
		CORSAllowedOrigins:   getEnvCSV("CORS_ALLOWED_ORIGINS", []string{"https://localhost:5173", "https://127.0.0.1:5173", "http://localhost:5173"}),
		LogFilePath:          getEnv("LOG_FILE", ""),
		LogMaxBytes:          int64(getEnvInt("LOG_MAX_BYTES", 10485760)),
		LogMaxBackups:        getEnvInt("LOG_MAX_BACKUPS", 5),
		ReportWorkerInterval: getEnvDuration("REPORT_WORKER_INTERVAL", 30*time.Second),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func getEnvCSV(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func ParseCIDRorIP(v string) (*net.IPNet, error) {
	if ip := net.ParseIP(v); ip != nil {
		if ip.To4() != nil {
			_, net4, _ := net.ParseCIDR(ip.String() + "/32")
			return net4, nil
		}
		_, net6, _ := net.ParseCIDR(ip.String() + "/128")
		return net6, nil
	}
	_, n, err := net.ParseCIDR(v)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c Config) ValidateSecurityKeys() error {
	if c.JWTSecret == "" || c.EncryptionKey == "" {
		return fmt.Errorf("security keys must be provided via environment variables")
	}
	return nil
}
