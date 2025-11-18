package option

import (
	"fmt"
	"maps"
	"reflect"

	"gopkg.in/yaml.v3"
)

// ContainerInterface defines the minimal interface needed for nested containers.
// This allows different Container[T] types to be stored in the same map.
type ContainerInterface interface {
	parseMap(data map[string]any) error
	JSONSchema() map[string]any
	GetName() string
	IsValid() bool
}

// Container holds a collection of options and nested containers.
// It provides parsing and JSON Schema generation capabilities.
// The type parameter T represents the source struct type.
type Container[T any] struct {
	Name           string
	Options        map[string]OptionInterface
	Containers     map[string]ContainerInterface
	jsonProperties map[string]any
	source         *T
	checks         []func(*T) bool
}

// NewContainer creates a new Container with the specified name and collects options from the provided struct.
func NewContainer[T any](name string, source *T) *Container[T] {
	oc := Container[T]{
		Name:           name,
		source:         source,
		jsonProperties: map[string]any{},
		checks:         make([]func(*T) bool, 0),
	}
	oc.Collect(source)
	return &oc
}

// GetName returns the container name, implementing ContainerInterface.
func (oc *Container[T]) GetName() string {
	return oc.Name
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

	// Recursively validate subcontainers
	for _, container := range oc.Containers {
		if !container.IsValid() {
			return false
		}
	}

	return true
}

// Collect gathers options and nested containers from the provided struct.
func (oc *Container[T]) Collect(s any) {
	oc.Options, oc.Containers = Collect(s)
}

// Parse parses YAML data and populates the container's options and nested containers.
func (oc *Container[T]) Parse(data []byte) error {
	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return oc.parseMap(yamlData)
}

func (oc *Container[T]) parseMap(data map[string]any) error {
	for key, value := range data {
		if opt, ok := oc.Options[key]; ok {
			if err := opt.SetAny(value); err != nil {
				return fmt.Errorf("failed to set option %s: %w", key, err)
			}
		} else if container, ok := oc.Containers[key]; ok {
			if mapValue, ok := value.(map[string]any); ok {
				if err := container.parseMap(mapValue); err != nil {
					return fmt.Errorf("failed to parse container %s: %w", key, err)
				}
			}
		}
	}
	return nil
}

// Collect extracts options and nested containers from a struct using reflection.
// It returns maps of options and containers found in the struct fields.
func Collect(s any) (map[string]OptionInterface, map[string]ContainerInterface) {
	resultOptions := make(map[string]OptionInterface)
	resultContainers := make(map[string]ContainerInterface)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return resultOptions, resultContainers
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldValue := field.Interface()
		if opt, ok := fieldValue.(OptionInterface); ok {
			resultOptions[opt.Name()] = opt
			continue
		}

		if field.Kind() == reflect.Pointer && !field.IsNil() {
			elem := field.Elem()
			if elem.Kind() == reflect.Struct {
				for j := 0; j < elem.NumField(); j++ {
					embeddedField := elem.Field(j)
					if embeddedField.Kind() == reflect.Pointer {
						if !embeddedField.IsNil() {
							if !embeddedField.CanInterface() {
								continue
							}
							// Check if it implements ContainerInterface
							if container, ok := embeddedField.Interface().(ContainerInterface); ok {
								resultContainers[container.GetName()] = container
								break
							}
						}
					}
				}
			}
		}
	}

	return resultOptions, resultContainers
}

// JSONSchema generates a JSON Schema representation of the container and its contents.
// It recursively includes schemas for all options and nested containers.
func (oc *Container[T]) JSONSchema() map[string]any {
	properties := make(map[string]any)

	for name, opt := range oc.Options {
		properties[name] = opt.JSONSchema()
	}

	for name, container := range oc.Containers {
		properties[name] = container.JSONSchema()
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if oc.Name != "" {
		schema["title"] = oc.Name
	}

	// Add any custom JSON properties from Check() calls
	maps.Copy(schema, oc.jsonProperties)

	return schema
}

// JSONSchemaWithMetadata generates a JSON Schema with $schema metadata field.
// This produces a complete, standalone JSON Schema document.
func (oc *Container[T]) JSONSchemaWithMetadata() map[string]any {
	schema := oc.JSONSchema()
	schema["$schema"] = "https://json-schema.org/draft-07/schema"
	return schema
}
