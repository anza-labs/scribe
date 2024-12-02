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

package manifestutils

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Annotations return the annotations for the resources.
func Annotations(instance metav1.ObjectMeta, filterAnnotations []string) (map[string]string, error) {
	// new map every time, so that we don't touch the instance's annotations
	annotations := map[string]string{}

	if nil != instance.Annotations {
		for k, v := range instance.Annotations {
			if !IsFilteredSet(k, filterAnnotations) {
				annotations[k] = v
			}
		}
	}

	return annotations, nil
}
