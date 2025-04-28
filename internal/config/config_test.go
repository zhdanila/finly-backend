package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewConfig_Success(t *testing.T) {
	os.Setenv("ENV", "development")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USERNAME", "user")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "redispassword")
	os.Setenv("REDIS_DB", "0")

	config, err := NewConfig()

	require.NoError(t, err)
	assert.Equal(t, "development", config.Env)
	assert.Equal(t, "8080", config.HTTPPort)
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "5432", config.DBPort)
	assert.Equal(t, "user", config.DBUsername)
	assert.Equal(t, "testdb", config.DBName)
	assert.Equal(t, "disable", config.DBSSLMode)
	assert.Equal(t, "password", config.DBPassword)
	assert.Equal(t, "localhost", config.RedisHost)
	assert.Equal(t, "6379", config.RedisPort)
	assert.Equal(t, "redispassword", config.RedisPassword)
	assert.Equal(t, 0, config.RedisDB)
}

func TestNewConfig_MissingEnvVars(t *testing.T) {
	os.Unsetenv("HTTP_PORT")

	config, err := NewConfig()

	require.Error(t, err)
	assert.Nil(t, config)
}

func TestNewConfig_InvalidRedisDB(t *testing.T) {
	os.Setenv("REDIS_DB", "invalid")

	config, err := NewConfig()

	require.Error(t, err)
	assert.Nil(t, config)
}

func TestLoad_Success(t *testing.T) {
	os.Setenv("ENV", "production")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USERNAME", "produser")
	os.Setenv("DB_NAME", "proddb")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("DB_PASSWORD", "prodpassword")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "prodredispassword")
	os.Setenv("REDIS_DB", "1")

	config := &Config{}
	err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "production", config.Env)
	assert.Equal(t, "8080", config.HTTPPort)
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "5432", config.DBPort)
	assert.Equal(t, "produser", config.DBUsername)
	assert.Equal(t, "proddb", config.DBName)
	assert.Equal(t, "disable", config.DBSSLMode)
	assert.Equal(t, "prodpassword", config.DBPassword)
	assert.Equal(t, "localhost", config.RedisHost)
	assert.Equal(t, "6379", config.RedisPort)
	assert.Equal(t, "prodredispassword", config.RedisPassword)
	assert.Equal(t, 1, config.RedisDB)
}

func TestValidateConfig_Success(t *testing.T) {
	config := &Config{
		Env:        "production",
		HTTPPort:   "8080",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUsername: "user",
		DBName:     "testdb",
		DBSSLMode:  "disable",
		DBPassword: "password",
		RedisHost:  "localhost",
		RedisPort:  "6379",
	}

	err := validateConfig(config)

	require.NoError(t, err)
}

func TestValidateConfig_Failure(t *testing.T) {
	config := &Config{
		Env:        "production",
		HTTPPort:   "",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUsername: "user",
		DBName:     "testdb",
		DBSSLMode:  "disable",
		DBPassword: "password",
		RedisHost:  "localhost",
		RedisPort:  "6379",
	}

	err := validateConfig(config)

	require.Error(t, err)
}
