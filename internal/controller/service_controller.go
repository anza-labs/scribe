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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	PrometheusScope *PrometheusScope
}

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch

// Reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx,
		"service", klog.KRef(req.Namespace, req.Name),
	)

	ok, err := r.PrometheusScope.autoDetect.PrometheusCRsAvailability()
	if err != nil {
		return ctrl.Result{}, err
	} else if !ok {
		return ctrl.Result{}, ErrPrometheusCRsNotAvailable
	}

	log.V(2).Info("Reconciling")

	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to get pod: %w", err)
	}

	// TODO: check if has monitoring enabled
	// TODO: get all annotation

	isMarkedToBeDeleted := svc.GetDeletionTimestamp() != nil
	if isMarkedToBeDeleted {
		// TODO: handle deletion of the adjacent resources
	}

	// TODO: handle creation/update of the resources

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
