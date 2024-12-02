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
	"errors"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete

var ErrPrometheusCRsNotAvailable = errors.New("prometheus custom resources cannot be found on cluster")

const (
	prometheusScrapeAnnotation = "prometheus.io/scrape"
	prometheusPortAnnotation   = "prometheus.io/port"
	prometheusPathAnnotation   = "prometheus.io/path"

	monitorsAnnotation         = "scribe.anza-labs.dev/monitors"
	monitorsAnnotationDisabled = "disabled"
)

type PrometheusScope struct {
	client.Client
	autoDetect autoDetect
}

func NewPrometheusScope(c client.Client, cfg *rest.Config) (*PrometheusScope, error) {
	dcl, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &PrometheusScope{
		Client: c,
		autoDetect: autoDetect{
			DiscoveryInterface: dcl,
		},
	}, nil
}

type autoDetect struct {
	discovery.DiscoveryInterface
}

func (a *autoDetect) PrometheusCRsAvailability() (bool, error) {
	apiList, err := a.ServerGroups()
	if err != nil {
		return false, err
	}

	foundServiceMonitor := false
	foundPodMonitor := false
	apiGroups := apiList.Groups
	for i := 0; i < len(apiGroups); i++ {
		if apiGroups[i].Name == "monitoring.coreos.com" {
			for _, version := range apiGroups[i].Versions {
				resources, err := a.ServerResourcesForGroupVersion(version.GroupVersion)
				if err != nil {
					return false, err
				}

				for _, resource := range resources.APIResources {
					if resource.Kind == "ServiceMonitor" {
						foundServiceMonitor = true
					} else if resource.Kind == "PodMonitor" {
						foundPodMonitor = true
					}
				}
			}
		}
	}

	if foundServiceMonitor && foundPodMonitor {
		return true, nil
	}

	return false, nil
}
