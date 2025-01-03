/*
Copyright 2025 anza-labs contributors.

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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationErrors(t *testing.T) {
	verr1 := &ValidationError{
		Key:  "test",
		Errs: []error{ErrSkipReconciliation},
	}
	assert.Equal(t, "Validation error at key \"test\": [skip reconciliation]", verr1.Message())

	verr2 := &ValidationError{
		Key:  "other",
		Errs: []error{io.EOF},
	}

	verrs := ValidationErrors{Items: []*ValidationError{verr1, verr2}}
	assert.Equal(t, "errors: Validation error at key \"test\": [skip reconciliation]; Validation error at key \"other\": [EOF]", verrs.Error())
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		verr        *ValidationError
		key         string
		inputErrors []error
	}{
		"success": {
			verr:        &ValidationError{Key: "test"},
			key:         "test",
			inputErrors: []error{ErrSkipReconciliation},
		},
		"success - joined error": {
			verr:        &ValidationError{Key: "test", Errs: []error{io.EOF}},
			key:         "test",
			inputErrors: []error{ErrSkipReconciliation},
		},
		"success - without input errors": {
			key: "test",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var expectedErrors []error
			if tc.verr != nil {
				expectedErrors = append(tc.inputErrors, tc.verr.Errs...)
			}

			valErr := NewValidationError(tc.verr, tc.key, tc.inputErrors...)
			assert.Equal(t, valErr.Key, tc.key)
			assert.ElementsMatch(t, expectedErrors, valErr.Errs)
		})
	}
}
