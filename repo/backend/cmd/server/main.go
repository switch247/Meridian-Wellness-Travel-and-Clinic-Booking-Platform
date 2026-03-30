package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"meridian/backend/internal/api"
	"meridian/backend/internal/api/handlers"
	"meridian/backend/internal/config"
	"meridian/backend/internal/logger"
	"meridian/backend/internal/platform/db"
	"meridian/backend/internal/platform/migrate"
	"meridian/backend/internal/repository"
	"meridian/backend/internal/security"
	"meridian/backend/internal/service"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateSecurityKeys(); err != nil {
		log.Fatalf("%v", err)
	}
	var logOut io.Writer = os.Stdout
	var rotator *logger.RotatingWriter
	if cfg.LogFilePath != "" {
		rw, err := logger.NewRotatingWriter(cfg.LogFilePath, cfg.LogMaxBytes, cfg.LogMaxBackups)
		if err == nil {
			rotator = rw
			logOut = rw
			defer rotator.Close()
		}
	}
	appLogger := logger.NewWithWriter(logOut)
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer pool.Close()

	migrationsPath := filepath.Join("migrations")
	if err := migrate.Run(ctx, pool, migrationsPath); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := migrate.Seed(ctx, pool, filepath.Join("seed", "seed.sql")); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	encryptor, err := security.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("encryption setup failed: %v", err)
	}

	repo := repository.New(pool)
	authSvc := service.NewAuthService(repo, cfg, encryptor, appLogger.Logger)
	profileSvc := service.NewProfileService(repo, cfg, encryptor)
	bookingSvc := service.NewBookingService(repo, cfg)

	authH := handlers.NewAuthHandler(authSvc)
	domainH := handlers.NewDomainHandler(profileSvc, bookingSvc, repo, encryptor, cfg.SlotGranularity)

	go func() {
		ticker := time.NewTicker(cfg.ReportWorkerInterval)
		defer ticker.Stop()
		exportDir := os.Getenv("EXPORT_DIR")
		if exportDir == "" {
			exportDir = "/tmp/exports"
		}
		for range ticker.C {
			_ = repo.ProcessDueReportJobs(context.Background(), exportDir)
		}
	}()

	e := api.NewRouter(cfg, appLogger.Logger, authH, domainH)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}
	e.Server = srv

	if cfg.TLSEnabled {
		if _, err := os.Stat(cfg.TLSCertFile); err != nil {
			log.Fatalf("tls cert missing: %v", err)
		}
		if _, err := os.Stat(cfg.TLSKeyFile); err != nil {
			log.Fatalf("tls key missing: %v", err)
		}
		if err := e.StartTLS(srv.Addr, cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
		return
	}

	if err := e.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
