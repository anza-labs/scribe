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

package prometheus

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/anza-labs/scribe/internal/manifests/manifestutils"
	"github.com/anza-labs/scribe/internal/naming"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceMonitor(svc *corev1.Service) (*monitoringv1.ServiceMonitor, error) {
	name := naming.ServiceMonitor(svc.Name)
	labels := manifestutils.Labels(svc.ObjectMeta, name, ComponentServiceMonitor, nil)
	annotations, err := manifestutils.Annotations(svc.ObjectMeta, nil)
	if err != nil {
		return nil, err
	}

	if len(svc.Labels) == 0 {
		return nil, ErrMissingLabels
	}
	selector := metav1.SetAsLabelSelector(svc.Labels)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   svc.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: *selector,
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					svc.Namespace,
				},
			},
			Endpoints: []monitoringv1.Endpoint{}, // TODO: extract this to function
		},
	}, nil
}
