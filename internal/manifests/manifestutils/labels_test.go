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
// Copyright The registry Authors

package manifestutils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	objectName      = "my-instance"
	objectNamespace = "my-ns"
)

func TestLabelsCommonSet(t *testing.T) {
	// prepare
	meta := metav1.ObjectMeta{
		Name:      objectName,
		Namespace: objectNamespace,
	}

	// test
	labels := Labels(meta, objectName, "podmonitor", []string{})
	assert.Equal(t, "scribe", labels["app.kubernetes.io/managed-by"])
	assert.Equal(t, "my-ns.my-instance", labels["app.kubernetes.io/instance"])
	assert.Equal(t, "monitoring", labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "registry", labels["app.kubernetes.io/component"])
}

func TestLabelsPropagateDown(t *testing.T) {
	// prepare
	meta := metav1.ObjectMeta{
		Labels: map[string]string{
			"myapp":                  "mycomponent",
			"app.kubernetes.io/name": "test",
		},
	}

	// test
	labels := Labels(meta, objectName, "podmonitor", []string{})

	// verify
	assert.Len(t, labels, 7)
	assert.Equal(t, "mycomponent", labels["myapp"])
	assert.Equal(t, "test", labels["app.kubernetes.io/name"])
}

func TestLabelsFilter(t *testing.T) {
	meta := metav1.ObjectMeta{
		Labels: map[string]string{"test.bar.io": "foo", "test.foo.io": "bar"},
	}

	// This requires the filter to be in regex match form and not the other simpler wildcard one.
	labels := Labels(meta, objectName, "registry", []string{".*.bar.io"})

	// verify
	assert.Len(t, labels, 7)
	assert.NotContains(t, labels, "test.bar.io")
	assert.Equal(t, "bar", labels["test.foo.io"])
}
