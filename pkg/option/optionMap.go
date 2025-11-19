package option

import (
	"fmt"
	"maps"
	"reflect"
)

// Map represents an option type that holds a map of string keys to values.
type Map[T any] struct {
	baseOption[map[string]T, *Map[T]]
	ValueOption *baseOption[T, *Map[T]]
	factory     func() T
}

// NewMap creates a new Map option with the specified name.
func NewMap[T any](name string, factory ...func() T) *Map[T] {
	om := &Map[T]{}
	om.initCommon(name, om)
	if len(factory) > 0 {
		om.factory = factory[0]
	} else {
		om.ValueOption = &baseOption[T, *Map[T]]{}
		om.ValueOption.initCommon("", om)
	}
	return om
}

func (om *Map[T]) EmptyDefault() *Map[T] {
	return om.Default(make(map[string]T))
}

// IsValid checks if a set value passes all check function
// if no value is set, but a defaultValue return true
// if neither is set, return false
func (om *Map[T]) IsValid() bool {
	if !om.baseOption.IsValid() {
		return false
	}
	if om.value != nil {
		for _, elem := range *om.value {
			if om.factory != nil {
				if opt, ok := any(elem).(OptionInterface); ok {
					if !opt.IsValid() {
						return false
					}
				}
			} else {
				om.ValueOption.Set(elem)
				if !om.ValueOption.IsValid() {
					return false
				}
			}
		}
	}
	return true
}

func (om *Map[T]) SetAny(value any) error {
	if typedValue, ok := value.(map[string]T); ok {
		om.Set(typedValue)
		return nil
	}

	mapValue := reflect.ValueOf(value)
	if mapValue.Kind() != reflect.Map {
		return fmt.Errorf("cannot convert %T to map[string]%T", value, *new(T))
	}

	keyType := mapValue.Type().Key()
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("map key type must be string, got %v", keyType)
	}

	result := make(map[string]T)
	iter := mapValue.MapRange()

	for iter.Next() {
		key := iter.Key().String()
		elem := iter.Value().Interface()

		if om.factory != nil {
			newElem := om.factory()
			if opt, ok := any(newElem).(OptionInterface); ok {
				if err := opt.SetAny(elem); err != nil {
					return err
				}
				result[key] = newElem
			} else {
				return fmt.Errorf("factory returned type %T which does not implement CommonInterface", newElem)
			}
		} else {
			if err := om.ValueOption.SetAny(elem); err != nil {
				return err
			}

			result[key] = om.ValueOption.Get()
		}
	}

	om.Set(result)
	return nil
}

func (om *Map[T]) JSONSchema() map[string]any {
	var additional map[string]any
	if om.factory != nil {
		newElem := om.factory()
		if opt, ok := any(newElem).(OptionInterface); ok {
			additional = opt.JSONSchema()
		} else {
			additional = make(map[string]any)
		}
	} else {
		additional = om.ValueOption.JSONSchema()
	}

	property := map[string]any{
		"type":                 "object",
		"additionalProperties": additional,
	}

	maps.Copy(property, om.jsonSchemaProperties)

	return property
}
