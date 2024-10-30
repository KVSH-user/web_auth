package main

import (
	"context"
	"fmt"
	stlog "log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"web_auth/internal/adapters/db/postgres"
	"web_auth/internal/api"
	"web_auth/internal/config"
	"web_auth/internal/modules/auth"
	"web_auth/internal/modules/messages"
	"web_auth/internal/utils/mockDB"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	storage, err := postgres.New(ctx, cfg, log)
	if err != nil {
		stlog.Fatal("failed to connect to db")
	}
	defer storage.Close(ctx)

	authService := auth.New(log, storage, storage)
	messageService := messages.New(log, storage)

	if err = mockDB.SeedDatabase(ctx, storage, cfg.MockDB.UserCount, cfg.MockDB.MsgCount); err != nil {
		log.Error("can`t create mock for DB")
	}

	router := api.NewRouter(authService, messageService)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.REST.Port),
		Handler:      router,
		ReadTimeout:  cfg.REST.ReadTimeout,
		WriteTimeout: cfg.REST.WriteTimeout,
		IdleTimeout:  cfg.REST.IdleTimeout,
	}

	Print(cfg)

	if err := srv.ListenAndServe(); err != nil {
		stlog.Fatal("server failed:", err)
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func Print(cfg *config.Config) {
	time.Sleep(time.Second * 3)

	stlog.Println("==========SSO APP STARTED===========")
	stlog.Printf("|gRPC PORT................%d\n", cfg.REST.Port)
	stlog.Printf("|POSTGRESQL HOST..........%s\n", cfg.Postgres.Host)
	stlog.Printf("|POSTGRESQL PORT..........%d\n", cfg.Postgres.Port)
	stlog.Printf("|ENV CONFIG...............%s\n", cfg.Env)
	stlog.Println("====================================")
}
