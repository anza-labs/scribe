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

func PodMonitor(pod *corev1.Pod) (*monitoringv1.PodMonitor, error) {
	name := naming.PodMonitor(pod.Name)
	labels := manifestutils.Labels(pod.ObjectMeta, name, ComponentPodMonitor, nil)
	annotations, err := manifestutils.Annotations(pod.ObjectMeta, nil)
	if err != nil {
		return nil, err
	}

	if len(pod.Labels) == 0 {
		return nil, ErrMissingLabels
	}
	selector := metav1.SetAsLabelSelector(pod.Labels)

	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   pod.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: monitoringv1.PodMonitorSpec{

			Selector: *selector,
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					pod.Namespace,
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{}, // TODO: extract this to function
		},
	}, nil
}
