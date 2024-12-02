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

// Additional copyrights:
// Copyright The OpenTelemetry Authors

package manifests

import (
	"errors"
	"fmt"
	"reflect"

	"dario.cat/mergo"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	ErrImmutableChange = errors.New("immutable field change attempted")
)

// MutateFuncFor returns a mutate function based on the
// existing resource's concrete type. It supports currently
// only the following types or else panics:
// - PodMonitor
// - ServiceMonitor
// In order for the operator to reconcile other types, they must be added here.
// The function returned takes no arguments but instead uses the existing and desired inputs here. Existing is expected
// to be set by the controller-runtime package through a client get call.
func MutateFuncFor(existing, desired client.Object) controllerutil.MutateFn {
	return func() error {
		// Get the existing annotations and override any conflicts with the desired annotations
		// This will preserve any annotations on the existing set.
		existingAnnotations := existing.GetAnnotations()
		if err := mergeWithOverride(&existingAnnotations, desired.GetAnnotations()); err != nil {
			return err
		}
		existing.SetAnnotations(existingAnnotations)

		// Get the existing labels and override any conflicts with the desired labels
		// This will preserve any labels on the existing set.
		existingLabels := existing.GetLabels()
		if err := mergeWithOverride(&existingLabels, desired.GetLabels()); err != nil {
			return err
		}
		existing.SetLabels(existingLabels)

		if ownerRefs := desired.GetOwnerReferences(); len(ownerRefs) > 0 {
			existing.SetOwnerReferences(ownerRefs)
		}

		switch existing.(type) {
		case *monitoringv1.PodMonitor:
			pm := existing.(*monitoringv1.PodMonitor)
			wantPm := desired.(*monitoringv1.PodMonitor)
			mutatePodMonitor(pm, wantPm)

		case *monitoringv1.ServiceMonitor:
			sm := existing.(*monitoringv1.ServiceMonitor)
			wantSm := desired.(*monitoringv1.ServiceMonitor)
			mutateServiceMonitor(sm, wantSm)

		default:
			t := reflect.TypeOf(existing).String()
			return fmt.Errorf("missing mutate implementation for resource type: %s", t)
		}
		return nil
	}
}

func mergeWithOverride(dst, src interface{}) error {
	return mergo.Merge(dst, src, mergo.WithOverride)
}

func mutatePodMonitor(existing, desired *monitoringv1.PodMonitor) {
	existing.Spec.PodMetricsEndpoints = desired.Spec.PodMetricsEndpoints
	existing.Spec.Selector = desired.Spec.Selector
}

func mutateServiceMonitor(existing, desired *monitoringv1.ServiceMonitor) {
	// TODO: write mutation
	existing.Spec.Endpoints = desired.Spec.Endpoints
	existing.Spec.Selector = desired.Spec.Selector
}
