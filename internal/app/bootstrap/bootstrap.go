package bootstrap

import (
	"context"
	"finly-backend/internal/config"
	"finly-backend/internal/repository"
	"finly-backend/internal/service"
	"finly-backend/internal/transport/http/router"
	"finly-backend/pkg/db"
	"finly-backend/pkg/logger"
	"finly-backend/pkg/server"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func Website() {
	logger.InitLogger()

	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	db, err := db.NewPostgresDB(cfg)
	if err != nil {
		zap.L().Sugar().Errorf("error with connecting to database: %s", err.Error())
	}

	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	srv := server.NewServer(cfg.HTTPPort)
	router.RegisterRoutes(srv, services)

	zap.L().Sugar().Infof("Finly backend started on port %s", cfg.HTTPPort)
	go func() {
		if err = srv.Start(); err != nil {
			zap.L().Sugar().Fatalf("error with starting server: %s", err.Error())
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	zap.L().Sugar().Info("Finly backend Shutting Down")

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Sugar().Fatalf("error with shutting down server: %s", err.Error())
	}
}
