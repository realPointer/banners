package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App   `yaml:"app"`
		HTTP  `yaml:"http"`
		Log   `yaml:"logger"`
		PG    `yaml:"postgres"`
		Redis `yaml:"redis"`
		JWT   `yaml:"jwt"`
	}

	App struct {
		Name    string `env:"APP_NAME"    env-required:"true" yaml:"name"`
		Version string `env:"APP_VERSION" env-required:"true" yaml:"version"`
	}

	HTTP struct {
		Port string `env:"HTTP_PORT" env-required:"true" yaml:"port"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" env-required:"true" yaml:"log_level"`
	}

	PG struct {
		PoolMax int    `env:"PG_POOL_MAX" env-required:"true" yaml:"pool_max"`
		URL     string `env:"PG_URL"      env-required:"true"`
	}

	Redis struct {
		URL string `env:"REDIS_URL" env-required:"true"`
	}

	JWT struct {
		SignKey  string        `env-required:"true"                  env:"JWT_SIGN_KEY"`
		TokenTTL time.Duration `env-required:"true" yaml:"token_ttl" env:"JWT_TOKEN_TTL"`
		Salt     string        `env-required:"true" env:"JWT_SALT"`
	}
)

func NewConfig() (*Config, error) {
	//nolint:exhaustruct // Fields are initialized by cleanenv.ReadConfig
	cfg := &Config{}

	configPath := os.Getenv("CONFIG_PATH")

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return cfg, nil
}
