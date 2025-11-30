package option

import (
	"fmt"
	"maps"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// ContainerCheckFunc is a validation check for containers that returns an error message if validation fails.
// Return empty string if validation passes.
type ContainerCheckFunc[T any] func(*T) string

// Container holds a collection of options and nested containers.
// It provides parsing and JSON Schema generation capabilities.
// The type parameter T represents the source struct type.
type Container[T any] struct {
	name           string
	Options        map[string]OptionInterface
	jsonProperties map[string]any
	source         *T
	checks         []ContainerCheckFunc[T]
	transform      func(*T)
	required       bool
	isSet          bool
}

// NewContainer creates a new Container with the specified name and collects options from the provided struct.
func NewContainer[T any](name string, source *T) *Container[T] {
	oc := Container[T]{
		name:           name,
		source:         source,
		jsonProperties: map[string]any{},
		checks:         make([]ContainerCheckFunc[T], 0),
	}
	oc.Collect(source)
	return &oc
}

// GetName returns the container name, implementing ContainerInterface.
func (oc *Container[T]) GetName() string {
	return oc.name
}

// Required marks this container as required by its parent container.
// Required containers must have at least one value set for IsValid() to pass.
func (oc *Container[T]) Required() *Container[T] {
	oc.required = true
	return oc
}

// IsRequired returns true if this container is marked as required.
func (oc *Container[T]) IsRequired() bool {
	return oc.required
}

// Description sets a description for the container, used in JSON Schema generation.
func (oc *Container[T]) Description(desc string) *Container[T] {
	oc.jsonProperties["description"] = desc
	return oc
}

// Check adds a validation check function that receives typed access to the source struct.
// The check function should return an empty string if validation passes, or an error message if it fails.
// The prop parameter should contain JSON Schema properties that reflect the same validation
// logic as the check function, allowing external tools to validate the same constraints.
func (oc *Container[T]) Check(check ContainerCheckFunc[T], prop map[string]any) *Container[T] {
	maps.Copy(oc.jsonProperties, prop)
	oc.checks = append(oc.checks, check)
	return oc
}

// Transform sets a transformation function that will be applied to the source struct
// after parsing. The transformation receives a pointer to the source struct and can
// modify it in place. This is useful for normalizing values or computing derived fields.
func (oc *Container[T]) Transform(transform func(*T)) *Container[T] {
	oc.transform = transform
	return oc
}

// IsValid checks if the container is valid according to JSON Schema semantics.
// It returns true if:
// - Container is not set, not required, and no child options have values
// - Container is set and all checks pass and all options are valid
// It returns false if:
// - Container is not set but is required
// - Container is set (or has child values) and any check or option fails validation
func (oc *Container[T]) IsValid() bool {
	return !oc.Validate("$").HasErrors()
}

// Validate performs validation and returns all errors with their JSONPath locations.
// The path parameter is the JSONPath prefix for this container (e.g., "$.config" or "$").
func (oc *Container[T]) Validate(path string) *ValidationErrors {
	// set default path to root $
	if path == "" {
		path = "$"
	}
	errs := NewValidationErrors()

	// Check if any child option has a value
	hasChildValues := false
	for _, opt := range oc.Options {
		if opt.HasValue() {
			hasChildValues = true
			break
		}
	}

	// If container was never set and no children have values, only fail if required
	if !oc.isSet && !hasChildValues {
		if oc.required {
			errs.Add(path, "required container is missing")
		}
		return errs
	}

	// Container is set or has child values - run checks on this container's source
	for _, check := range oc.checks {
		if msg := check(oc.source); msg != "" {
			errs.Add(path, msg)
		}
	}

	// Check all options are valid
	for name, opt := range oc.Options {
		optPath := JoinPath(path, name)
		optErrs := opt.Validate(optPath)
		errs.Merge(optErrs, "")
	}

	return errs
}

// Collect gathers options and nested containers from the provided struct.
func (oc *Container[T]) Collect(s any) {
	oc.Options = Collect(s)
}

// Parse parses YAML data and populates the container's options and nested containers.
func (oc *Container[T]) Parse(data []byte) error {
	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return oc.parseMap(yamlData)
}

// ParseAndValidate parses YAML data and validates the container in one step.
// Returns a ValidationErrors if validation fails, allowing inspection of which options failed.
func (oc *Container[T]) ParseAndValidate(data []byte) error {
	if err := oc.Parse(data); err != nil {
		return err
	}
	return oc.Validate("$").ErrorOrNil()
}

// ParseAndValidateFile reads data from a file, parses it, and validates the container.
func (oc *Container[T]) ParseAndValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return oc.ParseAndValidate(data)
}

func (oc *Container[T]) SetAny(data any) error {
	if dataMap, ok := data.(map[string]any); !ok {
		return fmt.Errorf("error setting %s, data is not of type map[string]any, but %T", oc.name, data)
	} else {
		return oc.parseMap(dataMap)
	}
}

func (oc *Container[T]) parseMap(data map[string]any) error {
	oc.isSet = true
	for key, value := range data {
		if opt, ok := oc.Options[key]; ok {
			if err := opt.SetAny(value); err != nil {
				return fmt.Errorf("failed to set option %s: %w", key, err)
			}
		}
	}
	if oc.transform != nil {
		oc.transform(oc.source)
	}
	return nil
}

// Collect extracts options and nested containers from a struct using reflection.
// It returns maps of options and containers found in the struct fields.
func Collect(s any) map[string]OptionInterface {
	resultOptions := make(map[string]OptionInterface)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return resultOptions
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldValue := field.Interface()
		opt, ok := fieldValue.(OptionInterface)
		if !ok {
			continue
		}

		if field.Kind() == reflect.Pointer && field.IsNil() {
			continue
		}

		resultOptions[opt.GetName()] = opt
	}

	return resultOptions
}

// JSONSchema generates a JSON Schema representation of the container and its contents.
// It recursively includes schemas for all options and nested containers.
func (oc *Container[T]) JSONSchema() map[string]any {
	properties := make(map[string]any)
	requiredFields := make([]string, 0)

	for name, opt := range oc.Options {
		properties[name] = opt.JSONSchema()
		if opt.IsRequired() {
			requiredFields = append(requiredFields, name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(requiredFields) > 0 {
		schema["required"] = requiredFields
	}

	if oc.name != "" {
		schema["title"] = oc.name
	}

	// Add any custom JSON properties from Check() calls
	maps.Copy(schema, oc.jsonProperties)

	return schema
}

// JSONSchemaWithMetadata generates a JSON Schema with $schema metadata field.
// This produces a complete, standalone JSON Schema document.
func (oc *Container[T]) JSONSchemaWithMetadata() map[string]any {
	schema := oc.JSONSchema()
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	return schema
}

// ToMap exports the container's options as a map[string]any for unstructured serialization.
// It calls GetAny() on each option and skips options that return nil (no value set).
// Nested containers are automatically recursed since they also implement GetAny().
func (oc *Container[T]) ToMap() map[string]any {
	result := make(map[string]any)
	for name, opt := range oc.Options {
		if value := opt.GetAny(); value != nil {
			result[name] = value
		}
	}
	return result
}

// GetAny returns the container's options as map[string]any, implementing OptionInterface.
// Returns nil if the container was never set (no input received).
// This enables auto-recursion when containers are nested within other containers or slices/maps.
func (oc *Container[T]) GetAny() any {
	if !oc.isSet {
		return nil
	}
	return oc.ToMap()
}

// HasValue returns true if the container has been set (received input via SetAny/Parse).
// This distinguishes between "never received input" and "received empty object {}".
func (oc *Container[T]) HasValue() bool {
	return oc.isSet
}

// NotZero returns true if the container has been set and contains at least one non-nil option value.
func (oc *Container[T]) NotZero() bool {
	if !oc.isSet {
		return false
	}
	return len(oc.ToMap()) > 0
}
