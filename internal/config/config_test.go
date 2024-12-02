package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/runtime/schema"
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
	require.NoError(t, err)

	require.Len(t, cfg.Types, 2)

	assert.Equal(t, "v1", cfg.Types[0].APIVersion)
	assert.Equal(t, "Namespace", cfg.Types[0].Kind)
	assert.Equal(t,
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		cfg.Types[0].GroupVersionKind(),
	)

	assert.Equal(t, "apps/v1", cfg.Types[1].APIVersion)
	assert.Equal(t, "Deployment", cfg.Types[1].Kind)
	assert.Equal(t,
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		cfg.Types[1].GroupVersionKind(),
	)
}
