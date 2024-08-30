package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App           `yaml:"app"`
		HTTP          `yaml:"http"`
		GRPC          `yaml:"grpc"`
		PG            `yaml:"postgres"`
		PubSub        `yaml:"pubsub"`
		Notifications `yaml:"notifications"`
	}

	App struct {
		Name     string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version  string `env-required:"true" yaml:"version" env:"APP_VERSION"`
		LogLevel string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	HTTP struct {
		Port int32 `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	GRPC struct {
		Port int32 `env-required:"true" yaml:"port" env:"GRPC_PORT"`
	}

	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		DSN     string `env-required:"true" yaml:"dsn" env:"PG_DSN"`
	}

	Notifications struct {
		MaxBatchSize int32 `env-required:"true" yaml:"batch_size_max" env:"NOTIFICATIONS_BATCH_SIZE_MAX"`
		Interval     int   `env-required:"true" yaml:"interval" env:"NOTIFICATIONS_INTERVAL"`
	}

	PubSub struct {
		Enabled    bool   `env-required:"true" yaml:"enabled" env:"PUBSUB_ENABLED"`
		ProjectID  string `env-required:"true" yaml:"project_id" env:"PUBSUB_PROJECT_ID"`
		UsersTopic string `env-required:"true" yaml:"users_topic" env:"PUBSUB_USERS_TOPIC"`
	}
)

// NewConfig returns app config.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	if path == "" {
		fmt.Println("Config path not provided")
		return nil, fmt.Errorf("config path not provided")
	}

	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
