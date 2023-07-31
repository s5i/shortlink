package config

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

//go:embed example.yaml
var ExampleConfig string

// Config stores configuration options for Shortlink.
type Config struct {
	Hostname string `yaml:"hostname"`
	Listen   string `yaml:"listen"`

	OAuthClientID     string `yaml:"oauth_client_id"`
	OAuthClientSecret string `yaml:"oauth_client_secret"`

	JWTSecret string        `yaml:"jwt_secret"`
	JWTTTL    time.Duration `yaml:"jwt_ttl"`

	DefaultRedirectURL string `yaml:"default_redirect_url"`

	Admins []string `yaml:"admin"`
}

// Read unmarshals a file into Config.
func Read(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", path, err)
	}
	return cfg, nil
}

func CreateExample(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("file %q already exists", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create %q directory: %v", dir, err)
	}

	if err := ioutil.WriteFile(path, []byte(ExampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to create %q file: %v", path, err)
	}

	return nil
}
