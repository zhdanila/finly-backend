package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_Load(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USERNAME", "user")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_PASSWORD", "redis_password")
	t.Setenv("REDIS_DB", "0")
	t.Setenv("HTTP_PORT", "8080")

	// Mock the loadConfig function
	viper.SetConfigFile(".env.test")
	viper.Set("DB_HOST", "localhost")
	viper.Set("DB_PORT", "5432")
	viper.Set("DB_USERNAME", "user")
	viper.Set("DB_NAME", "testdb")
	viper.Set("DB_SSLMODE", "disable")
	viper.Set("DB_PASSWORD", "password")
	viper.Set("REDIS_HOST", "localhost")
	viper.Set("REDIS_PORT", "6379")
	viper.Set("REDIS_PASSWORD", "redis_password")
	viper.Set("REDIS_DB", 0)
	viper.Set("HTTP_PORT", "8080")

	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "user", cfg.DBUsername)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "disable", cfg.DBSSLMode)
	assert.Equal(t, "password", cfg.DBPassword)
	assert.Equal(t, "localhost", cfg.RedisHost)
	assert.Equal(t, "6379", cfg.RedisPort)
	assert.Equal(t, "redis_password", cfg.RedisPassword)
	assert.Equal(t, 0, cfg.RedisDB)
	assert.Equal(t, "8080", cfg.HTTPPort)
}

func TestNewConfig_InvalidEnv(t *testing.T) {
	// Set an invalid ENV value
	t.Setenv("ENV", "invalid")
	cfg, err := NewConfig()

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadStagingConfig(t *testing.T) {
	// Simulate setting environment variables for staging
	t.Setenv("ENV", "staging")
	t.Setenv("HTTP_PORT", "8081")
	t.Setenv("DB_HOST", "staging-db-host")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USERNAME", "staging_user")
	t.Setenv("DB_NAME", "staging_db")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_PASSWORD", "staging_password")
	t.Setenv("REDIS_HOST", "staging-redis-host")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "staging_redis_password")
	t.Setenv("REDIS_DB", "1")

	cfg := &Config{}
	err := loadStagingConfig(cfg)
	require.NoError(t, err)

	// Verify the values were loaded correctly
	assert.Equal(t, "8081", cfg.HTTPPort)
	assert.Equal(t, "staging-db-host", cfg.DBHost)
	assert.Equal(t, "5433", cfg.DBPort)
	assert.Equal(t, "staging_user", cfg.DBUsername)
	assert.Equal(t, "staging_db", cfg.DBName)
	assert.Equal(t, "require", cfg.DBSSLMode)
	assert.Equal(t, "staging_password", cfg.DBPassword)
	assert.Equal(t, "staging-redis-host", cfg.RedisHost)
	assert.Equal(t, "6380", cfg.RedisPort)
	assert.Equal(t, "staging_redis_password", cfg.RedisPassword)
	assert.Equal(t, 1, cfg.RedisDB)
}
