package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListObjects(t *testing.T) {
	t.Parallel()

	// Setup the fake Kubernetes client
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	for name, tc := range map[string]struct {
		namespace      string
		existingPods   []unstructured.Unstructured
		expectedResult []types.NamespacedName
		expectedError  error
	}{
		"list pods in a namespace": {
			namespace: "test-namespace",
			existingPods: []unstructured.Unstructured{
				newUnstructuredPod("test-namespace", "pod1"),
				newUnstructuredPod("test-namespace", "pod2"),
			},
			expectedResult: []types.NamespacedName{
				{Namespace: "test-namespace", Name: "pod1"},
				{Namespace: "test-namespace", Name: "pod2"},
			},
		},
		"no pods in namespace": {
			namespace:      "test-namespace",
			existingPods:   []unstructured.Unstructured{},
			expectedResult: []types.NamespacedName{},
		},
		"pods in different namespaces": {
			namespace: "test-namespace",
			existingPods: []unstructured.Unstructured{
				newUnstructuredPod("test-namespace", "pod1"),
				newUnstructuredPod("other-namespace", "pod2"),
			},
			expectedResult: []types.NamespacedName{
				{Namespace: "test-namespace", Name: "pod1"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Add existing pods to the fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(convertUnstructuredToObjects(tc.existingPods)...).
				Build()

			// Create the UnstructuredReconciler
			reconciler := &UnstructuredReconciler{
				Client: fakeClient,
				gvk:    corev1.SchemeGroupVersion.WithKind("Pod"),
			}

			// Call the listObjects method
			result, err := reconciler.listObjects(context.Background(), tc.namespace)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.ElementsMatch(t, tc.expectedResult, result)
		})
	}
}

// Helper function to create an Unstructured object of type Pod
func newUnstructuredPod(namespace, name string) unstructured.Unstructured {
	pod := unstructured.Unstructured{}
	pod.SetAPIVersion("v1")
	pod.SetKind("Pod")
	pod.SetNamespace(namespace)
	pod.SetName(name)
	return pod
}

// Helper function to convert a slice of unstructured.Unstructured to runtime.Object slice
func convertUnstructuredToObjects(items []unstructured.Unstructured) []client.Object {
	objects := make([]client.Object, len(items))
	for i, u := range items {
		copy := u.DeepCopy()
		objects[i] = copy
	}
	return objects
}
