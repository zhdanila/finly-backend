package db

import (
	"context"
	"finly-backend/internal/config"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPostgresDB_Success(t *testing.T) {
	cnf := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUsername: "postgres",
		DBPassword: "qwerty",
		DBName:     "finly",
		DBSSLMode:  "disable",
	}

	db, err := NewPostgresDB(cnf)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()
}

func TestNewPostgresDB_Failure(t *testing.T) {
	cnf := &config.Config{
		DBHost:     "wronghost",
		DBPort:     "5432",
		DBUsername: "postgres",
		DBPassword: "password",
		DBName:     "testdb",
		DBSSLMode:  "disable",
	}

	db, err := NewPostgresDB(cnf)
	require.Error(t, err)
	require.Nil(t, db)
}

func TestWithTransaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = WithTransaction(context.Background(), sqlxDB, func(tx *sqlx.Tx) error {
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTransaction_FailRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = WithTransaction(context.Background(), sqlxDB, func(tx *sqlx.Tx) error {
		return fmt.Errorf("some error")
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
