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
	baseCnf := &Config{}

	if err := baseCnf.Load(); err != nil {
		return nil, err
	}

	return baseCnf, nil
}

func (cnf *Config) Load() error {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	cnf.Env = os.Getenv("ENV")
	cnf.HTTPPort = os.Getenv("HTTP_PORT")

	cnf.DBHost = os.Getenv("DB_HOST")
	cnf.DBPort = os.Getenv("DB_PORT")
	cnf.DBUsername = os.Getenv("DB_USERNAME")
	cnf.DBName = os.Getenv("DB_NAME")
	cnf.DBSSLMode = os.Getenv("DB_SSLMODE")
	cnf.DBPassword = os.Getenv("DB_PASSWORD")

	cnf.RedisHost = os.Getenv("REDIS_HOST")
	cnf.RedisPort = os.Getenv("REDIS_PORT")
	cnf.RedisPassword = os.Getenv("REDIS_PASSWORD")

	redisDB := os.Getenv("REDIS_DB")
	if redisDB != "" {
		redisDBInt, err := strconv.Atoi(redisDB)
		if err != nil {
			return fmt.Errorf("failed to convert REDIS_DB to int: %v", err)
		}
		cnf.RedisDB = redisDBInt
	}

	if err := viper.Unmarshal(cnf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if err := validateConfig(cnf); err != nil {
		return err
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
