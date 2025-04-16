package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
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
	baseCnf := &Config{}

	if err := baseCnf.Load(); err != nil {
		return nil, err
	}

	return baseCnf, nil
}

func (cnf *Config) Load() error {
	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Try to load .env file if it exists
	if err := loadEnvFile(); err != nil {
		return err
	}

	envVars := []string{
		"ENV", "HTTP_PORT", "DB_HOST", "DB_PORT", "DB_USERNAME",
		"DB_NAME", "DB_SSLMODE", "DB_PASSWORD", "REDIS_HOST",
		"REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	}
	if err := bindEnvVars(envVars); err != nil {
		return err
	}

	// Unmarshal configuration into the Config struct
	if err := viper.Unmarshal(cnf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate the config struct
	if err := validateConfig(cnf); err != nil {
		return err
	}

	return nil
}

func loadEnvFile() error {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("error reading config file: %v", err)
		}
		// .env file not found, rely on environment variables
	}
	return nil
}

func bindEnvVars(vars []string) error {
	for _, v := range vars {
		if err := viper.BindEnv(v); err != nil {
			return fmt.Errorf("failed to bind %s: %v", v, err)
		}
	}
	return nil
}

func validateConfig(cnf *Config) error {
	validate := validator.New()
	if err := validate.Struct(cnf); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	return nil
}
