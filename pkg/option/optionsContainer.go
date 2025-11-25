package option

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a validation failure with details about which options failed.
// This type is designed to be expanded in the future with more specific error information.
type ValidationError struct {
	Container      string
	InvalidOptions []string
}

func (e *ValidationError) Error() string {
	if len(e.InvalidOptions) == 0 {
		return fmt.Sprintf("validation failed for container %q", e.Container)
	}
	return fmt.Sprintf("validation failed for container %q: invalid options: [%s]",
		e.Container, strings.Join(e.InvalidOptions, ", "))
}

// Container holds a collection of options and nested containers.
// It provides parsing and JSON Schema generation capabilities.
// The type parameter T represents the source struct type.
type Container[T any] struct {
	name           string
	Options        map[string]OptionInterface
	jsonProperties map[string]any
	source         *T
	checks         []func(*T) bool
}

// NewContainer creates a new Container with the specified name and collects options from the provided struct.
func NewContainer[T any](name string, source *T) *Container[T] {
	oc := Container[T]{
		name:           name,
		source:         source,
		jsonProperties: map[string]any{},
		checks:         make([]func(*T) bool, 0),
	}
	oc.Collect(source)
	return &oc
}

// GetName returns the container name, implementing ContainerInterface.
func (oc *Container[T]) Name() string {
	return oc.name
}

// Check adds a validation check function that receives typed access to the source struct.
// The check function should return true if validation passes, false otherwise.
// The prop parameter should contain JSON Schema properties that reflect the same validation
// logic as the check function, allowing external tools to validate the same constraints.
func (oc *Container[T]) Check(check func(*T) bool, prop map[string]any) *Container[T] {
	maps.Copy(oc.jsonProperties, prop)
	oc.checks = append(oc.checks, check)
	return oc
}

// IsValid checks if all options, checks, and subcontainers are valid.
// It returns true only if:
// - All container-level checks return true
// - All options in the Options map have IsValid() == true
// - All subcontainers have IsValid() == true
func (oc *Container[T]) IsValid() bool {
	// Run checks on this container's source
	for _, check := range oc.checks {
		if !check(oc.source) {
			return false
		}
	}

	// Check all options are valid
	for _, opt := range oc.Options {
		if !opt.IsValid() {
			return false
		}
	}

	return true
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
// Returns a ValidationError if validation fails, allowing inspection of which options failed.
func (oc *Container[T]) ParseAndValidate(data []byte) error {
	if err := oc.Parse(data); err != nil {
		return err
	}
	if !oc.IsValid() {
		return &ValidationError{
			Container:      oc.name,
			InvalidOptions: []string{},
		}
	}
	return nil
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
	for key, value := range data {
		if opt, ok := oc.Options[key]; ok {
			if err := opt.SetAny(value); err != nil {
				return fmt.Errorf("failed to set option %s: %w", key, err)
			}
		}
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

		resultOptions[opt.Name()] = opt
	}

	return resultOptions
}

// JSONSchema generates a JSON Schema representation of the container and its contents.
// It recursively includes schemas for all options and nested containers.
func (oc *Container[T]) JSONSchema() map[string]any {
	properties := make(map[string]any)

	for name, opt := range oc.Options {
		properties[name] = opt.JSONSchema()
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
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
// This enables auto-recursion when containers are nested within other containers or slices/maps.
func (oc *Container[T]) GetAny() any {
	return oc.ToMap()
}
