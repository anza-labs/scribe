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
	"context"
	"fmt"
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

const (
	annotationsAnnotation = "scribe.anza-labs.dev/annotations"
)

// lister is an interface that defines the listObjects method which returns a list of namespaced names.
type lister interface {
	listObjects(context.Context, string) ([]types.NamespacedName, error)
}

// mapFunc returns a function that triggers a reconcile request based on the provided lister.
// It logs the namespace details and returns reconcile requests for each object in the namespace.
func mapFunc(l lister) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		log := log.FromContext(ctx,
			"group_version_kind", (&corev1.Namespace{}).GroupVersionKind(),
			"namespaced_name", klog.KObj(obj),
		)

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
// It merges the new annotations with existing ones.
func (ss *NamespaceScope) UpdateAnnotations(
	ctx context.Context,
	annotations map[string]string,
) (map[string]string, error) {
	ns := &corev1.Namespace{}

	results := map[string]string{}
	maps.Copy(results, annotations)

	if err := ss.Get(ctx, ss.namespace, ns); err != nil {
		return nil, fmt.Errorf("unable to get namespace: %w", err)
	}

	s, ok := ns.Annotations[annotationsAnnotation]
	if !ok {
		return results, nil
	}

	for k, v := range parseAnnotations(s) {
		results[k] = v
	}

	return results, nil
}

// parseAnnotations parses a string containing key-value pairs into a map.
// The input string should be formatted as comma-separated key=value pairs.
// Newline characters are treated as commas for parsing.
func parseAnnotations(input string) map[string]string {
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
