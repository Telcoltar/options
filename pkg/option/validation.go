package option

import (
	"fmt"
	"strings"
	"unicode"
)

// FieldError represents a single validation error with its JSONPath location.
type FieldError struct {
	Path    string // JSONPath like "$.users[0].email" or "$.config.timeout"
	Message string // Human-readable error message
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// ValidationErrors collects multiple validation errors with their paths.
type ValidationErrors struct {
	Errors []FieldError
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation passed"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("validation failed: %s", e.Errors[0].Error())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("validation failed with %d errors:\n", len(e.Errors)))
	for _, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// Add appends a new field error to the collection.
func (e *ValidationErrors) Add(path, message string) {
	e.Errors = append(e.Errors, FieldError{Path: path, Message: message})
}

// AddError appends an existing FieldError to the collection.
func (e *ValidationErrors) AddError(err FieldError) {
	e.Errors = append(e.Errors, err)
}

// Merge combines errors from another ValidationErrors, optionally prepending a path prefix.
// The prefix is joined with existing paths using appropriate JSONPath syntax.
func (e *ValidationErrors) Merge(other *ValidationErrors, pathPrefix string) {
	if other == nil {
		return
	}
	for _, err := range other.Errors {
		newPath := JoinPath(pathPrefix, err.Path)
		e.Errors = append(e.Errors, FieldError{Path: newPath, Message: err.Message})
	}
}

// HasErrors returns true if there are any validation errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors, otherwise returns itself.
// This is useful for returning from Validate() methods.
func (e *ValidationErrors) ErrorOrNil() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}

// NewValidationErrors creates a new empty ValidationErrors collection.
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]FieldError, 0),
	}
}

// JoinPath combines two JSONPath segments appropriately.
// Examples:
//   - JoinPath("$", "foo") => "$.foo"
//   - JoinPath("$.foo", "bar") => "$.foo.bar"
//   - JoinPath("$.foo", "[0]") => "$.foo[0]"
//   - JoinPath("$.foo", "$.bar") => "$.foo.bar" (strips leading $ from second)
func JoinPath(base, suffix string) string {
	if base == "" {
		return suffix
	}
	if suffix == "" {
		return base
	}

	// Strip leading "$" or "$." from suffix if present
	cleanSuffix := suffix
	if strings.HasPrefix(cleanSuffix, "$.") {
		cleanSuffix = cleanSuffix[2:]
	} else if strings.HasPrefix(cleanSuffix, "$") {
		cleanSuffix = cleanSuffix[1:]
	}

	if cleanSuffix == "" {
		return base
	}

	// If suffix starts with "[", no dot needed
	if strings.HasPrefix(cleanSuffix, "[") {
		return base + cleanSuffix
	}

	// Otherwise join with a dot
	return base + "." + cleanSuffix
}

// FieldPath creates a field path for a named field.
func FieldPath(name string) string {
	return "$." + name
}

// IndexPath creates an array index path segment.
func IndexPath(index int) string {
	return fmt.Sprintf("[%d]", index)
}

// KeyPath creates a map key path segment.
func KeyPath(key string) string {
	// For keys that can use dot notation, use dot notation; otherwise, use bracket notation
	if canUseDotNotation(key) {
		return key
	}
	return fmt.Sprintf("['%s']", key)
}

// canUseDotNotation returns true if the key can be used with dot notation in a JSONPath.
func canUseDotNotation(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		if i == 0 && unicode.IsDigit(r) {
			return false // Can't start with digit
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}
