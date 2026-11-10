package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cs7sc/server/internal/auth"
	"cs7sc/server/internal/config"
	"cs7sc/server/internal/httpapi"
	"cs7sc/server/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if len(os.Args) > 1 {
		os.Exit(runCommand(os.Args[1], os.Args[2:], logger))
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Error("konfigurasi tidak valid", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startupCtx, cancelStartup := context.WithTimeout(ctx, 15*time.Second)
	defer cancelStartup()

	db, err := store.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("membuka basis data", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(startupCtx); err != nil {
		logger.Error("basis data tidak dapat dihubungi", "error", err)
		os.Exit(1)
	}

	logger.Info("audit", "action", "server_start", "ip", "internal")

	hasher := auth.NewHasher(cfg.ArgonMemory, cfg.ArgonIterations, cfg.ArgonParallelism)
	api, err := httpapi.New(cfg, db, hasher, logger)
	if err != nil {
		logger.Error("menyiapkan server", "error", err)
		os.Exit(1)
	}
	api.StartLimiterSweeper(ctx)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go cleanupSessions(ctx, db, logger)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	logger.Info("server berjalan", "addr", cfg.Addr)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server berhenti", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("menghentikan server", "error", err)
			os.Exit(1)
		}
		logger.Info("audit", "action", "server_stop", "ip", "internal")
		logger.Info("server berhenti dengan rapi")
	}
}

func audit(logger *slog.Logger, action, subject, ip string) {
	logger.Info("audit", "action", action, "subject", subject, "ip", ip)
}

func cleanupSessions(ctx context.Context, db *store.Store, logger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			deleted, err := db.DeleteExpiredSessions(jobCtx, time.Now())
			cancel()
			if err != nil {
				logger.Error("membersihkan sesi kedaluwarsa", "error", err)
				continue
			}
			if deleted > 0 {
				logger.Info("sesi kedaluwarsa dibersihkan", "jumlah", deleted)
			}
		}
	}
}
