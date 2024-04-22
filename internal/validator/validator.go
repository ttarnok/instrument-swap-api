// Package Validator helps to validate incoming json data.
package validator

import (
	"regexp"
	"slices"
)

// Regexp to validate email addresses.
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Validator stores errors that come up during the validation of imcoming json data.
type Validator struct {
	Errors map[string]string
}

// New creates a new validator instance.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true is we dont have any validation errors in the Validator.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds a new error to the Validator.
// If the Validator already has an error with the provided key,
// the function will ommit the new message.
func (v *Validator) AddError(key, message string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = message
	}
}

// Check checks the first expression, if it is false, adds the error to the validator.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// -----------------------------------------------------------------------------

// PermittedValue returns true if the given value is in the permitted values slice.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches return true if the given string matches the provided regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique returns true if all vales az unique in the provided slice.
func Unique[T comparable](values []T) bool {
	if len(values) == 0 {
		return true
	}

	uniqueValues := make(map[T]struct{}, len(values))

	for _, v := range values {
		uniqueValues[v] = struct{}{}
	}

	return len(values) == len(uniqueValues)
}
