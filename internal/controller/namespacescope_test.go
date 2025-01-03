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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestUpdateAnnotations(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	for name, tc := range map[string]struct {
		// Input parameters
		object               *corev1.Pod
		namespaceAnnotations map[string]string
		// Expected output
		expectedResult map[string]string
		expectedError  error
	}{
		"empty": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			namespaceAnnotations: map[string]string{},
			expectedError:        ErrSkipReconciliation,
		},
		"add annotations": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			namespaceAnnotations: map[string]string{
				annotations: marshalAnnotations(map[string]string{
					"key1": "value1",
				}),
			},
			expectedResult: map[string]string{
				"key1": "value1",
				lastAppliedAnnotations: marshalAnnotations(map[string]string{
					"key1": "value1",
				}),
			},
		},
		"add annotations with template": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Annotations: map[string]string{},
				},
			},
			namespaceAnnotations: map[string]string{
				annotations: marshalAnnotations(map[string]string{
					"key1": "{{ .metadata.name }}",
				}),
			},
			expectedResult: map[string]string{
				"key1": "test-pod",
				lastAppliedAnnotations: marshalAnnotations(map[string]string{
					"key1": "test-pod",
				}),
			},
		},
		"append annotations": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"key1": "value1",
						lastAppliedAnnotations: marshalAnnotations(map[string]string{
							"key1": "value1",
						}),
					},
				},
			},
			namespaceAnnotations: map[string]string{
				annotations: marshalAnnotations(map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
			},
			expectedResult: map[string]string{
				"key1": "value1",
				"key2": "value2",
				lastAppliedAnnotations: marshalAnnotations(map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
			},
		},
		"remove annotations": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"key1": "value1",
						"key2": "value2",
						lastAppliedAnnotations: marshalAnnotations(map[string]string{
							"key1": "value1",
							"key2": "value2",
						}),
					},
				},
			},
			namespaceAnnotations: map[string]string{
				annotations: marshalAnnotations(map[string]string{
					"key1": "value1",
				}),
			},
			expectedResult: map[string]string{
				"key1": "value1",
				lastAppliedAnnotations: marshalAnnotations(map[string]string{
					"key1": "value1",
				}),
			},
		},
		"update annotations": {
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"key1": "value1",
						"key2": "old-value",
						lastAppliedAnnotations: marshalAnnotations(map[string]string{
							"key1": "value1",
							"key2": "old-value",
						}),
					},
				},
			},
			namespaceAnnotations: map[string]string{
				annotations: marshalAnnotations(map[string]string{
					"key1": "value1",
					"key2": "new-value", // Updated value for key2
				}),
			},
			expectedResult: map[string]string{
				"key1": "value1",
				"key2": "new-value", // key2 updated to new value
				lastAppliedAnnotations: marshalAnnotations(map[string]string{
					"key1": "value1",
					"key2": "new-value", // Updated value in lastAppliedAnnotations
				}),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-namespace",
						Namespace:   "test-namespace",
						Annotations: tc.namespaceAnnotations,
					},
				}).
				Build()

			nss := NewNamespaceScope(fakeClient, "test-namespace")

			unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tc.object)
			require.NoError(t, err)

			result, err := nss.UpdateAnnotations(context.Background(), tc.object.ObjectMeta.Annotations, unstructuredObj)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestUnmarshalAnnotations(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input    string
		expected map[string]string
	}{
		"simple_case": {
			input:    "key1=value1,key2=value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		"with_newlines": {
			input:    "key1=value1,\nkey2=value2\nkey3=value3",
			expected: map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
		},
		"with_whitespace": {
			input:    "  key1  =  value1 ,  key2=value2  ",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		"empty_string": {
			input:    "",
			expected: map[string]string{},
		},
		"invalid_pairs": {
			input:    "key1=value1,key2,key3=value3",
			expected: map[string]string{"key1": "value1", "key3": "value3"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := unmarshalAnnotations(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMarshalAnnotations(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input    map[string]string
		expected string
	}{
		"simple_case": {
			input:    map[string]string{"key1": "value1", "key2": "value2"},
			expected: "key1=value1,\nkey2=value2",
		},
		"single_pair": {
			input:    map[string]string{"key1": "value1"},
			expected: "key1=value1",
		},
		"empty_map": {
			input:    map[string]string{},
			expected: "",
		},
		"with_special_chars": {
			input:    map[string]string{"key1": "value1", "key_2": "value=2"},
			expected: "key1=value1,\nkey_2=value=2",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := marshalAnnotations(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMapFunc(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	for name, tc := range map[string]struct {
		namespaceAnnotations map[string]string
		expectedRequests     []reconcile.Request
	}{
		"managed namespace with objects": {
			namespaceAnnotations: map[string]string{
				annotations: "key=value",
			},
			expectedRequests: []reconcile.Request{
				{NamespacedName: types.NamespacedName{Namespace: "test-namespace", Name: "pod1"}},
				{NamespacedName: types.NamespacedName{Namespace: "test-namespace", Name: "pod2"}},
			},
		},
		"unmanaged namespace": {
			namespaceAnnotations: map[string]string{},
			expectedRequests:     nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Setup the fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(
					&corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "test-namespace",
							Namespace:   "test-namespace",
							Annotations: tc.namespaceAnnotations,
						},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "test-namespace",
						},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod2",
							Namespace: "test-namespace",
						},
					},
				).
				Build()

			// Create a mock lister
			lister := &UnstructuredReconciler{
				Client: fakeClient,
				Scheme: scheme,
				gvk:    corev1.SchemeGroupVersion.WithKind("Pod"),
			}

			// Create the map function
			mapFn := mapFunc(lister)

			// Create a test object to pass to the map function
			testObject := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-namespace",
					Namespace: "test-namespace",
				},
			}

			// Call the map function
			requests := mapFn(context.Background(), testObject)

			// Verify the results
			assert.ElementsMatch(t, tc.expectedRequests, requests)
		})
	}
}
