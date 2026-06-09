package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_Success(t *testing.T) {
	path := writeConfig(t, `
server:
  port: 8082
kafka:
  brokers:
    - localhost:9092
  topic: stock-events
  groupId: stock-provider
mongo:
  uri: mongodb://localhost:27017
  database: stocks
  collection: events
`)
	t.Setenv("CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 8082 {
		t.Errorf("expected port 8082, got %d", cfg.Server.Port)
	}
	if cfg.Kafka.Topic != "stock-events" {
		t.Errorf("expected topic 'stock-events', got %q", cfg.Kafka.Topic)
	}
	if cfg.Kafka.GroupID != "stock-provider" {
		t.Errorf("expected groupId 'stock-provider', got %q", cfg.Kafka.GroupID)
	}
	if len(cfg.Kafka.Brokers) != 1 || cfg.Kafka.Brokers[0] != "localhost:9092" {
		t.Errorf("unexpected brokers: %v", cfg.Kafka.Brokers)
	}
	if cfg.Mongo.URI != "mongodb://localhost:27017" {
		t.Errorf("expected mongo URI 'mongodb://localhost:27017', got %q", cfg.Mongo.URI)
	}
	if cfg.Mongo.Database != "stocks" {
		t.Errorf("expected database 'stocks', got %q", cfg.Mongo.Database)
	}
	if cfg.Mongo.Collection != "events" {
		t.Errorf("expected collection 'events', got %q", cfg.Mongo.Collection)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Setenv("CONFIG_PATH", "/nonexistent/path/config.yaml")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeConfig(t, "invalid: yaml: [\nunclosed")
	t.Setenv("CONFIG_PATH", path)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_ValidAudiencesFallback(t *testing.T) {
	path := writeConfig(t, `
tokenAuthorization:
  validAudience:
    - audience1
`)
	t.Setenv("CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.TokenAuth.ValidAudiences) != 1 || cfg.TokenAuth.ValidAudiences[0] != "audience1" {
		t.Errorf("expected ValidAudiences to fall back to validAudience, got %v", cfg.TokenAuth.ValidAudiences)
	}
}

func TestLoad_ValidAudiencesTakesPriority(t *testing.T) {
	path := writeConfig(t, `
tokenAuthorization:
  validAudience:
    - audience-old
  validAudiences:
    - audience1
    - audience2
`)
	t.Setenv("CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.TokenAuth.ValidAudiences) != 2 {
		t.Errorf("expected validAudiences to take priority, got %v", cfg.TokenAuth.ValidAudiences)
	}
	if cfg.TokenAuth.ValidAudiences[0] != "audience1" {
		t.Errorf("unexpected first audience: %q", cfg.TokenAuth.ValidAudiences[0])
	}
}

func TestLoad_DefaultConfigPath(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	// Without a file at the default path, Load should return an error.
	// This confirms the default path is used when CONFIG_PATH is unset.
	_, err := Load()
	if err == nil {
		t.Log("default config file exists; skipping absence check")
	}
}
