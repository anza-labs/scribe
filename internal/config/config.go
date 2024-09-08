package config

import "k8s.io/apimachinery/pkg/runtime/schema"

type Config struct {
	Types []Type `json:"types"`
}

type Type struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

func (t *Type) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(t.APIVersion, t.Kind)
}
