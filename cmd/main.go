package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Gilf4/effective-mobile-task/docs"
	"github.com/Gilf4/effective-mobile-task/internal/config"
	"github.com/Gilf4/effective-mobile-task/internal/http/handler"
	"github.com/Gilf4/effective-mobile-task/internal/http/middleware"
	"github.com/Gilf4/effective-mobile-task/internal/repository/db"
	"github.com/Gilf4/effective-mobile-task/internal/service"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// @title Subscriptions API
// @version 1.0
// @description REST сервис для управления подписками пользователя.
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting application",
		slog.String("env", cfg.Env),
		slog.Any("Server config", cfg.Server),
	)

	ctx := context.Background()

	repo, err := db.NewSubscriptionRepository(ctx, &cfg.DB)
	if err != nil {
		log.Error("failed to init repo", "err", err)
		os.Exit(1)
	}

	service := service.NewSubscriptionService(repo)

	h := handler.NewHandler(service, log)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	handler := middleware.Logging(log)(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen error", "err", err)
		}
		log.Info("server started", slog.Int("port", cfg.Server.Port))
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", "err", err)
	}

	log.Info("server exited properly")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
