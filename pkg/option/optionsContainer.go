package option

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Container holds a collection of options and nested containers.
// It provides parsing and JSON Schema generation capabilities.
type Container struct {
	Name       string
	Options    map[string]OptionInterface
	Containers map[string]*Container
}

// NewContainer creates a new Container with the specified name and collects options from the provided struct.
func NewContainer(name string, options any) *Container {
	oc := Container{
		Name: name,
	}
	oc.Collect(options)
	return &oc
}

// Collect gathers options and nested containers from the provided struct.
func (oc *Container) Collect(s any) {
	oc.Options, oc.Containers = Collect(s)
}

// Parse parses YAML data and populates the container's options and nested containers.
func (oc *Container) Parse(data []byte) error {
	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return oc.parseMap(yamlData)
}

func (oc *Container) parseMap(data map[string]any) error {
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
func Collect(s any) (map[string]OptionInterface, map[string]*Container) {
	resultOptions := make(map[string]OptionInterface)
	resultContainers := make(map[string]*Container)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
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

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			elem := field.Elem()
			if elem.Kind() == reflect.Struct {
				for j := 0; j < elem.NumField(); j++ {
					embeddedField := elem.Field(j)
					if embeddedField.Kind() == reflect.Ptr && embeddedField.Type() == reflect.TypeOf((*Container)(nil)) {
						if !embeddedField.IsNil() {
							container := embeddedField.Interface().(*Container)
							resultContainers[container.Name] = container
						}
						break
					}
				}
			}
		}
	}

	return resultOptions, resultContainers
}

// JSONSchema generates a JSON Schema representation of the container and its contents.
// It recursively includes schemas for all options and nested containers.
func (oc *Container) JSONSchema() map[string]any {
	properties := make(map[string]any)

	for name, opt := range oc.Options {
		properties[name] = opt.JSONSchemaProperty()
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

	return schema
}

// JSONSchemaWithMetadata generates a JSON Schema with $schema metadata field.
// This produces a complete, standalone JSON Schema document.
func (oc *Container) JSONSchemaWithMetadata() map[string]any {
	schema := oc.JSONSchema()
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	return schema
}
