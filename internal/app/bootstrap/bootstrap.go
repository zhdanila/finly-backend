package bootstrap

import (
	"context"
	"finly-backend/internal/config"
	"finly-backend/pkg/logger"
	"finly-backend/pkg/server"
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

	srv := new(server.Server)

	go func() {
		if err := srv.Run(cfg.HTTPPort, nil); err != nil {
			log.Error("error with running server", "error", err)
		}
	}()

	log.Info("Finly backend started")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info("Finly backend Shutting Down")

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("error with shutting down server", "error", err)
	}
}
