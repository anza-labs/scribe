package config

import (
	"strings"
	"testing"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

func TestUnmarshall(t *testing.T) {
	t.Parallel()

	configFile := strings.NewReader(`---
types:
- apiVersion: v1
  kind: Namespace
- apiVersion: apps/v1
  kind: Deployment
`)

	cfg := Config{}

	err := yaml.NewDecoder(configFile).Decode(&cfg)
	if err != nil {
		t.Errorf("Unexpected error while decoding config: %v", err)
		return
	}

	if len(cfg.Types) != 2 {
		t.Errorf("Unexpected length of Types: expected %v, got %v", 2, len(cfg.Types))
		return
	}

	if cfg.Types[0].APIVersion != "v1" {
		t.Errorf("Unexpected apiVersion: expected %v, got %v", "v1", cfg.Types[0].APIVersion)
		return
	}

	if cfg.Types[0].Kind != "Namespace" {
		t.Errorf("Unexpected kind: expected %v, got %v", "Namespace", cfg.Types[0].Kind)
		return
	}

	if cfg.Types[1].APIVersion != "apps/v1" {
		t.Errorf("Unexpected apiVersion: expected %v, got %v", "apps/v1", cfg.Types[1].APIVersion)
		return
	}

	if cfg.Types[1].Kind != "Deployment" {
		t.Errorf("Unexpected Kind: expected %v, got %v", "Deployment", cfg.Types[1].Kind)
		return
	}
}
