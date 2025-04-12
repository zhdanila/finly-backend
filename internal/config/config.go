package config

import (
	"fmt"
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
	viper.SetConfigFile(`.env`)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}
	if err := viper.Unmarshal(cnf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return nil
}
