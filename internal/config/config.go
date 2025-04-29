package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

type Config struct {
	Env      string `mapstructure:"ENV" validate:"required"`
	HTTPPort string `mapstructure:"HTTP_PORT" validate:"required"`

	// Database configuration
	DBHost     string `mapstructure:"DB_HOST" validate:"required"`
	DBPort     string `mapstructure:"DB_PORT" validate:"required"`
	DBUsername string `mapstructure:"DB_USERNAME" validate:"required"`
	DBName     string `mapstructure:"DB_NAME" validate:"required"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE" validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD" validate:"required"`

	// Redis configuration
	RedisHost     string `mapstructure:"REDIS_HOST" validate:"required"`
	RedisPort     string `mapstructure:"REDIS_PORT" validate:"required"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) Load() error {
	cfg.Env = os.Getenv("ENV")

	switch cfg.Env {
	case "dev":
		if err := loadDevConfig(cfg); err != nil {
			return err
		}
	case "staging":
		if err := loadStagingConfig(cfg); err != nil {
			return err
		}
	case "test":
		if err := loadTestConfig(cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid ENV value: %s", cfg.Env)
	}

	if err := validateConfig(cfg); err != nil {
		return err
	}

	return nil
}

func loadDevConfig(cfg *Config) error {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err)
	}
	return nil
}

func loadStagingConfig(cfg *Config) error {
	cfg.HTTPPort = getEnv("HTTP_PORT", "")
	cfg.DBHost = getEnv("DB_HOST", "")
	cfg.DBPort = getEnv("DB_PORT", "")
	cfg.DBUsername = getEnv("DB_USERNAME", "")
	cfg.DBName = getEnv("DB_NAME", "")
	cfg.DBSSLMode = getEnv("DB_SSLMODE", "")
	cfg.DBPassword = getEnv("DB_PASSWORD", "")
	cfg.RedisHost = getEnv("REDIS_HOST", "")
	cfg.RedisPort = getEnv("REDIS_PORT", "")
	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")

	redisDB := os.Getenv("REDIS_DB")
	if redisDB != "" {
		redisDBInt, err := strconv.Atoi(redisDB)
		if err != nil {
			return fmt.Errorf("failed to convert REDIS_DB to int: %v", err)
		}
		cfg.RedisDB = redisDBInt
	}
	return nil
}

func loadTestConfig(cfg *Config) error {
	viper.SetConfigFile("test.env")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err)
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	return nil
}
