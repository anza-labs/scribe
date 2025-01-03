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
	"errors"
	"fmt"
	"maps"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

const AnnotationValidationFailure = "AnnotationValidationFailure"

// ValidationErrors represents a collection of validation errors.
type ValidationErrors struct {
	Items []*ValidationError
}

// Error formats all validation errors into a single string.
func (ve *ValidationErrors) Error() string {
	if len(ve.Items) == 0 {
		return "no validation errors"
	}
	parts := make([]string, 0, len(ve.Items))
	for _, item := range ve.Items {
		parts = append(parts, item.Message())
	}
	return fmt.Sprintf("errors: %s", strings.Join(parts, "; "))
}

func (ve *ValidationErrors) Message() string {
	if len(ve.Items) == 0 {
		return "No validation errors"
	}
	parts := make([]string, 0, len(ve.Items))
	for _, item := range ve.Items {
		parts = append(parts, item.Message())
	}
	return strings.Join(parts, "; ")
}

// ValidationError represents an error associated with a specific key.
type ValidationError struct {
	Key  string
	Errs []error
}

// NewValidationError creates a new ValidationError or appends errors to an existing one.
func NewValidationError(existing *ValidationError, key string, errs ...error) *ValidationError {
	if existing == nil {
		return &ValidationError{Key: key, Errs: errs}
	}
	existing.Key = key
	existing.Errs = append(existing.Errs, errs...)
	return existing
}

// AppendError adds an error to the ValidationError.
func (ve *ValidationError) AppendError(err error) {
	ve.Errs = append(ve.Errs, err)
}

// Message formats the ValidationError into a readable string.
func (ve *ValidationError) Message() string {
	if len(ve.Errs) == 0 {
		return fmt.Sprintf("validation error at key %q with no specific error details", ve.Key)
	}
	errMessages := make([]string, 0, len(ve.Errs))
	for _, err := range ve.Errs {
		errMessages = append(errMessages, err.Error())
	}
	return fmt.Sprintf("Validation error at key %q: [%s]", ve.Key, strings.Join(errMessages, ", "))
}

func errsFromStrs(strs []string) []error {
	errs := make([]error, 0, len(strs))
	for _, s := range strs {
		errs = append(errs, errors.New(s))
	}

	return errs
}

func ValidateAnnotations(annotations map[string]string) (map[string]string, *ValidationErrors) {
	result := maps.Clone(annotations)

	var validationErrs []*ValidationError
	for k := range annotations {
		var verr *ValidationError

		errStrs := validation.IsQualifiedName(k)
		if errStrs != nil {
			verr = NewValidationError(verr, k, errsFromStrs(errStrs)...)
		}

		if verr != nil {
			delete(result, k)
			validationErrs = append(validationErrs, verr)
		}
	}

	if len(validationErrs) > 0 {
		return result, &ValidationErrors{Items: validationErrs}
	}

	return result, nil
}
