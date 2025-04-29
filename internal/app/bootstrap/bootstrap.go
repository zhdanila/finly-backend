package bootstrap

import (
	"context"
	"errors"
	"finly-backend/internal/config"
	"finly-backend/internal/repository"
	"finly-backend/internal/service"
	"finly-backend/internal/transport/http/router"
	"finly-backend/pkg/db"
	"finly-backend/pkg/logger"
	"finly-backend/pkg/server"
	"fmt"
	"go.uber.org/zap"
	"net/http"
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

	postgres, err := db.NewPostgresDB(cfg)
	if err != nil {
		panic(err)
	}

	redis, err := db.NewRedisDB(ctx, cfg)
	if err != nil {
		panic(err)
	}

	repo := repository.NewRepository(postgres, redis)
	services := service.NewService(repo)
	srv := server.NewServer(cfg.HTTPPort)
	router.RegisterRoutes(srv, services)

	zap.L().Sugar().Infof("Finly backend started on port %s", cfg.HTTPPort)
	go func() {
		zap.L().Sugar().Info("Starting server...")
		if err = srv.Start(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			zap.L().Sugar().Fatalf("error with starting server: %s", err.Error())
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	zap.L().Sugar().Info("Finly backend shutting down")

	if err = srv.Shutdown(ctx); err != nil {
		zap.L().Sugar().Fatalf("error with shutting down server: %s", err.Error())
	}

	if err = postgres.Close(); err != nil {
		zap.L().Fatal(fmt.Sprintf("error with closing db: %s", err.Error()))
	}

	if err = redis.Close(); err != nil {
		zap.L().Fatal(fmt.Sprintf("error with closing db: %s", err.Error()))
	}
}
