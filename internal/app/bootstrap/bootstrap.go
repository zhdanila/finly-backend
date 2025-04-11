package bootstrap

import (
	"context"
	"finly-backend/internal/config"
	"finly-backend/internal/repository"
	"finly-backend/internal/repository/service"
	http2 "finly-backend/internal/transport/http"
	"finly-backend/pkg/db"
	"finly-backend/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Website() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	log := logger.SetupLogger(cfg.Env)

	db, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("error with connecting to database: %s", err.Error()))
	}
	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	srv := http2.NewServer(cfg.HTTPPort, services)

	fmt.Println("Starting server on port", cfg.HTTPPort)

	go func() {
		if err = srv.Start(); err != nil {
			log.Error("error with starting server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info("Finly backend Shutting Down")

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("error with shutting down server", "error", err)
	}
}
