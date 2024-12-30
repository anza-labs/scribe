/*
Copyright 2024 anza-labs contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

const (
	annotations            = "scribe.anza-labs.dev/annotations"
	lastAppliedAnnotations = "scribe.anza-labs.dev/last-applied-annotations"
)

var ErrSkipReconciliation = errors.New("skip reconciliation")

// lister is an interface that defines the listObjects method which returns a list of namespaced names.
type getLister interface {
	Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error
	listObjects(context.Context, string) ([]types.NamespacedName, error)
}

// mapFunc returns a function that triggers a reconcile request based on the provided lister.
// It logs the namespace details and returns reconcile requests for each object in the namespace.
func mapFunc(l getLister) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		ns := &corev1.Namespace{}

		log := log.FromContext(ctx,
			"group_version_kind", ns.GroupVersionKind(),
			"namespaced_name", klog.KObj(obj),
		)

		err := l.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, ns)
		if err != nil {
			log.V(0).Error(err, "Unable to get namespace to trigger reconcile")
			return nil
		}

		if _, ok := ns.Annotations[annotations]; !ok {
			log.V(3).Info("Skipping unmanaged namespace")
			return nil
		}

		namespace := obj.GetName()

		nns, err := l.listObjects(ctx, namespace)
		if err != nil {
			log.V(0).Error(err, "Unable to trigger reconcile")
			return nil
		}

		req := []reconcile.Request{}

		for _, nn := range nns {
			req = append(req, reconcile.Request{NamespacedName: nn})
		}

		return req
	}
}

// NamespaceScope defines the scope of operations for a specific namespace.
// It contains the client to interact with the Kubernetes API and the namespace name.
type NamespaceScope struct {
	client.Client
	namespace types.NamespacedName
}

// NewNamespaceScope creates a new instance of NamespaceScope for the given namespace name.
func NewNamespaceScope(c client.Client, ns string) *NamespaceScope {
	return &NamespaceScope{
		Client: c,
		namespace: types.NamespacedName{
			Namespace: ns,
			Name:      ns,
		},
	}
}

// UpdateAnnotations updates the annotations of a namespace.
// It synchronizes annotations with the new ones, removes missing ones, and tracks last-applied annotations.
func (ss *NamespaceScope) UpdateAnnotations(
	ctx context.Context,
	objAnnotations map[string]string,
	object map[string]any,
) (map[string]string, error) {
	ns := &corev1.Namespace{}

	if err := ss.Get(ctx, ss.namespace, ns); err != nil {
		return nil, fmt.Errorf("unable to get namespace: %w", err)
	}

	tpl, err := template.New("").Parse(ns.Annotations[annotations])
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, object)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Retrieve expected and last-applied annotations
	expected := unmarshalAnnotations(buf.String())
	lastApplied := unmarshalAnnotations(objAnnotations[lastAppliedAnnotations])
	if len(expected) == 0 && len(lastApplied) == 0 {
		return nil, ErrSkipReconciliation
	}

	// Calculate the resulting annotations
	results := make(map[string]string)
	maps.Copy(results, objAnnotations) // Start with current annotations

	// Add/Update new annotations
	for k, v := range expected {
		results[k] = v
	}

	// Remove annotations that were in last-applied but are missing in newAnnotations
	for k := range lastApplied {
		if _, exists := expected[k]; !exists {
			delete(results, k)
		}
	}

	final := make(map[string]string)
	maps.Copy(final, results)
	delete(results, lastAppliedAnnotations)

	final[lastAppliedAnnotations] = marshalAnnotations(results)

	return final, nil
}

// unmarshalAnnotations parses a string containing key-value pairs into a map.
// The input string should be formatted as comma-separated key=value pairs.
// Newline characters are treated as commas for parsing.
func unmarshalAnnotations(input string) map[string]string {
	result := make(map[string]string)

	// Normalize the string by replacing newlines and whitespace followed by commas
	input = strings.ReplaceAll(input, ",\n", ",")
	input = strings.ReplaceAll(input, "\n", ",")
	input = strings.TrimSpace(input)

	// Split by comma to get individual key=value pairs
	pairs := strings.Split(input, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		// Split by equal sign to get key and value
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			result[key] = value
		}
	}

	return result
}

// marshalAnnotations converts a map into a formatted string of key-value pairs.
// The output string will be formatted as comma-separated key=value pairs,
// with each pair appearing on a new line for readability. The keys are sorted.
func marshalAnnotations(annotations map[string]string) string {
	var builder strings.Builder

	// Collect keys into a slice
	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, key)
	}

	// Sort the keys
	slices.Sort(keys)

	// Iterate over the sorted keys and build the result string
	for i, key := range keys {
		if i > 0 {
			builder.WriteString(",\n")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(annotations[key])
	}

	return builder.String()
}
