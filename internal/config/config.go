package config

import (
	"fmt"
	"os"

	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	commonmiddleware "github.com/pocwithmehul/common-go-lib/pkg/middleware"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config/dev/config.yaml"

type Config struct {
	Server    ServerConfig                              `yaml:"server"`
	Kafka     KafkaConfig                               `yaml:"kafka"`
	Mongo     MongoConfig                               `yaml:"mongo"`
	Datadog   commonlogger.DatadogConfig                `yaml:"datadog"`
	TokenAuth commonmiddleware.TokenAuthorizationConfig `yaml:"-"`
}

type rawConfig struct {
	Server             ServerConfig               `yaml:"server"`
	Kafka              KafkaConfig                `yaml:"kafka"`
	Mongo              MongoConfig                `yaml:"mongo"`
	Datadog            commonlogger.DatadogConfig `yaml:"datadog"`
	TokenAuthorization tokenAuthConfig            `yaml:"tokenAuthorization"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"groupId"`
}

type MongoConfig struct {
	URI        string `yaml:"uri"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type tokenAuthConfig struct {
	JWKURI         string              `yaml:"jwk_uri"`
	ValidIssuers   []string            `yaml:"validIssuers"`
	ValidAudience  []string            `yaml:"validAudience"`
	ValidAudiences []string            `yaml:"validAudiences"`
	ExpectedScopes map[string][]string `yaml:"expectedScopes"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", configPath, err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", configPath, err)
	}

	validAudiences := raw.TokenAuthorization.ValidAudiences
	if len(validAudiences) == 0 {
		validAudiences = raw.TokenAuthorization.ValidAudience
	}

	return &Config{
		Server:  raw.Server,
		Kafka:   raw.Kafka,
		Mongo:   raw.Mongo,
		Datadog: raw.Datadog,
		TokenAuth: commonmiddleware.TokenAuthorizationConfig{
			JWKURI:         raw.TokenAuthorization.JWKURI,
			ValidIssuers:   raw.TokenAuthorization.ValidIssuers,
			ValidAudiences: validAudiences,
			ExpectedScopes: raw.TokenAuthorization.ExpectedScopes,
		},
	}, nil
}
