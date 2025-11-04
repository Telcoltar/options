package option

import (
	"fmt"
	"reflect"
)

// Map represents an option type that holds a map of string keys to values.
type Map[T any] struct {
	baseOption[map[string]T, *Map[T]]
	valueEnumValues []T
}

// NewMap creates a new Map option with the specified name.
func NewMap[T any](name string) *Map[T] {
	om := &Map[T]{}
	om.initCommon(name, om)
	return om
}

// InnerTransform applies a transform to each value inside the map on Set.
func (om *Map[T]) InnerTransform(transform func(val T) T) *Map[T] {
	om.transform = func(m map[string]T) map[string]T {
		if m == nil {
			return nil
		}
		res := make(map[string]T, len(m))
		for k, v := range m {
			res[k] = transform(v)
		}
		return res
	}
	return om
}

func (om *Map[T]) InnerChecks(checks ...func(val T) bool) *Map[T] {
	for _, check := range checks {
		om.checks = append(om.checks, func(val map[string]T) bool {
			for _, elem := range val {
				if !check(elem) {
					return false
				}
			}
			return true
		})
	}
	return om
}

func (om *Map[T]) KeyChecks(checks ...func(key string) bool) *Map[T] {
	for _, check := range checks {
		om.checks = append(om.checks, func(val map[string]T) bool {
			for key := range val {
				if !check(key) {
					return false
				}
			}
			return true
		})
	}
	return om
}

func (om *Map[T]) EmptyDefault() *Map[T] {
	emptyMap := make(map[string]T)
	om.defaultValue = &emptyMap
	return om
}

func (om *Map[T]) Enum(allowedValues ...T) *Map[T] {
	om.valueEnumValues = allowedValues
	return om.InnerChecks(func(val T) bool {
		for _, allowed := range allowedValues {
			if reflect.DeepEqual(val, allowed) {
				return true
			}
		}
		return false
	})
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

	var emptyT T
	targetType := reflect.TypeOf(emptyT)
	result := make(map[string]T)
	iter := mapValue.MapRange()

	for iter.Next() {
		key := iter.Key().String()
		elem := iter.Value().Interface()

		if typedElem, ok := elem.(T); ok {
			result[key] = typedElem
			continue
		}

		converted, err := tryConvert(elem, targetType)
		if err != nil {
			return fmt.Errorf("cannot convert map value at key %q: %w", key, err)
		}
		result[key] = converted.Interface().(T)
	}

	om.Set(result)
	return nil
}

func (om *Map[T]) JSONSchemaType() string {
	return "object"
}

func (om *Map[T]) JSONSchemaProperty() map[string]any {
	var zero T
	valueType := reflectTypeToJSONSchemaType(reflect.TypeOf(zero))

	additional := map[string]any{
		"type": valueType,
	}

	if len(om.valueEnumValues) > 0 {
		enumValues := make([]any, len(om.valueEnumValues))
		for i, v := range om.valueEnumValues {
			enumValues[i] = v
		}
		additional["enum"] = enumValues
	}

	property := map[string]any{
		"type":                 "object",
		"additionalProperties": additional,
	}

	if om.defaultValue != nil {
		property["default"] = *om.defaultValue
	}

	return property
}
