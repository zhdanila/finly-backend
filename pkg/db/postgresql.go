package db

import (
	"context"
	"finly-backend/internal/config"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"go.uber.org/zap"
	"time"
)

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cnf *config.Config) (*sqlx.DB, error) {
	dbConfig := DBConfig{
		Host:     cnf.DBHost,
		Port:     cnf.DBPort,
		Username: cnf.DBUsername,
		DBName:   cnf.DBName,
		SSLMode:  cnf.DBSSLMode,
		Password: cnf.DBPassword,
	}

	var db *sqlx.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.DBName, dbConfig.Password, dbConfig.SSLMode))
		if err == nil {
			err = db.Ping()
			if err == nil {
				zap.L().Info("DB connected")
				return db, nil
			}
		}

		zap.L().Error("Database not ready, retrying in 1 seconds...", zap.Error(err))
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("unable to connect to database after 10 attempts: %w", err)
}

func WithTransaction(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v; original error: %w", rbErr, err)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
