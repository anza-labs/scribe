/*
Copyright 2024-2025 anza-labs contributors.

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
	"errors"
	"fmt"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// UnstructuredReconciler reconciles a Unstructured object
type UnstructuredReconciler struct {
	client.Client
	gvk      schema.GroupVersionKind
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *UnstructuredReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	u := r.empty(req)

	log := log.FromContext(ctx,
		"group_version_kind", r.gvk,
		"namespaced_name", klog.KObj(u),
	)

	log.V(2).Info("Reconciling")

	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		if apierrors.IsNotFound(err) {
			// If the resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.V(2).Info("Not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get the resource: %w", err)
	}

	isMarkedToBeDeleted := u.GetDeletionTimestamp() != nil
	if isMarkedToBeDeleted {
		log.V(2).Info("Ignoring object with deletion timestamp")
		return ctrl.Result{}, nil
	}

	nss := NewNamespaceScope(r.Client, req.Namespace)

	ann, err := nss.UpdateAnnotations(ctx, u.GetAnnotations(), u.Object)
	if err != nil {
		if errors.Is(err, ErrSkipReconciliation) {
			log.V(2).Info("Ignoring unmanaged object")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to update the annotation map: %w", err)
	}

	ann, validationErrors := ValidateAnnotations(ann)
	if validationErrors != nil {
		validationErrorsCounter.With(prometheus.Labels{"source_namespace": req.Namespace}).Inc()

		log.V(1).Error(validationErrors, "Validation error")
		r.Recorder.Event(nss.namespace, corev1.EventTypeWarning, AnnotationValidationFailure, validationErrors.Message())
		for _, err := range validationErrors.Items {
			r.Recorder.Event(u, corev1.EventTypeWarning, AnnotationValidationFailure, err.Message())
		}
	}

	original := u.DeepCopy()
	u.SetAnnotations(ann)

	if reflect.DeepEqual(original.GetAnnotations(), u.GetAnnotations()) {
		log.V(2).Info("Nothing to do, skipping")
		return ctrl.Result{}, nil
	}

	if err := r.Update(ctx, u); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update annotations on object: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UnstructuredReconciler) SetupWithManager(mgr ctrl.Manager, gvk schema.GroupVersionKind) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	r.gvk = gvk

	return ctrl.NewControllerManagedBy(mgr).
		For(u).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(mapFunc(r)),
		).
		Complete(r)
}

func (r *UnstructuredReconciler) listObjects(ctx context.Context, namespace string) ([]types.NamespacedName, error) {
	log := log.FromContext(ctx,
		"group_version_kind", r.gvk,
		"namespace", namespace,
	)

	ul := r.emptyList()
	if err := r.List(ctx, ul, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list objects: %w", err)
	}

	nn := []types.NamespacedName{}

	for _, u := range ul.Items {
		nn = append(nn, types.NamespacedName{
			Namespace: u.GetNamespace(),
			Name:      u.GetName(),
		})
	}

	log.V(4).Info("Listed objects", "count", len(nn), "objects", nn)

	return nn, nil
}

func (r *UnstructuredReconciler) empty(req ctrl.Request) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}

	u.SetGroupVersionKind(r.gvk)
	u.SetName(req.Name)
	u.SetNamespace(req.Namespace)

	return u
}

func (r *UnstructuredReconciler) emptyList() *unstructured.UnstructuredList {
	ul := &unstructured.UnstructuredList{}

	apiVersion, kind := r.gvk.ToAPIVersionAndKind()
	ul.SetAPIVersion(apiVersion)
	ul.SetKind(kind + "List")

	return ul
}
