package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config stores configuration options for Shortlink.
type Config struct {
	Hostname string `yaml:"hostname"`
	Listen   string `yaml:"listen"`

	DatabasePath string `yaml:"database_path"`

	OAuthClientID     string `yaml:"oauth_client_id"`
	OAuthClientSecret string `yaml:"oauth_client_secret"`

	JWTSecret string        `yaml:"jwt_secret"`
	JWTTTL    time.Duration `yaml:"jwt_ttl"`

	DefaultRedirectURL string `yaml:"default_redirect_url"`

	Admins []string `yaml:"admin"`
}

// Read unmarshals a file into Config.
func ReadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", path, err)
	}
	return cfg, nil
}
