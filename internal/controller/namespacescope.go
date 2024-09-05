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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

const (
	enabledForAnnotation   = "scribe.anza-labs.dev/enabled-for"
	enabledForDaemonSets   = "daemonsets.apps"
	enabledForDeployments  = "deployments.apps"
	enabledForStatefulSets = "statefulsets.apps"

	annotationsAnnotation = "scribe.anza-labs.dev/annotations"
)

type lister interface {
	listObjects(context.Context, string) ([]types.NamespacedName, error)
}

func mapFunc(l lister) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		log := log.FromContext(ctx, "namespace", klog.KObj(obj))

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

// NamespaceScope
type NamespaceScope struct {
	client.Client
	namespace types.NamespacedName
}

// NewNamespaceScope
func NewNamespaceScope(c client.Client, ns string) *NamespaceScope {
	return &NamespaceScope{
		Client: c,
		namespace: types.NamespacedName{
			Namespace: ns,
			Name:      ns,
		},
	}
}

// UpdateAnnotations
func (ss *NamespaceScope) UpdateAnnotations(ctx context.Context, annotations map[string]string) (map[string]string, error) {
	ns := &corev1.Namespace{}

	if err := ss.Get(ctx, ss.namespace, ns); err != nil {
		return nil, err
	}

	// TODO: parse annotations

	return annotations, nil
}
