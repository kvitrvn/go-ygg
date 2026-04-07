package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host string `env:"GO_YGG_SERVER_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"GO_YGG_SERVER_PORT" envDefault:"8080"`
}

type DatabaseConfig struct {
	DSN string `env:"GO_YGG_DATABASE_DSN"`
}

type LogConfig struct {
	Level  string `env:"GO_YGG_LOG_LEVEL" envDefault:"info"`
	Format string `env:"GO_YGG_LOG_FORMAT" envDefault:"json"`
}

// Load reads the application config from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		var parseErr env.ParseError
		if errors.As(err, &parseErr) {
			if key, ok := envKeyForField(reflect.TypeOf(cfg), parseErr.Name); ok {
				return nil, fmt.Errorf("parse environment config %s: %w", key, parseErr.Err)
			}
		}

		return nil, fmt.Errorf("parse environment config: %w", err)
	}
	return &cfg, nil
}

func envKeyForField(t reflect.Type, fieldName string) (string, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			if key, ok := envKeyForField(field.Type, fieldName); ok {
				return key, true
			}
		}

		if field.Name != fieldName {
			continue
		}

		key := field.Tag.Get("env")
		if key == "" {
			return "", false
		}

		return strings.Split(key, ",")[0], true
	}

	return "", false
}
