package option

import (
	"fmt"
	"reflect"
)

// Slice represents an option type that holds a slice of values.
type Slice[T any] struct {
	baseOption[[]T, *Slice[T]]
	itemEnumValues []T
}

// NewSlice creates a new Slice option with the specified name.
func NewSlice[T any](name string) *Slice[T] {
	os := &Slice[T]{}
	os.self = os
	os.name = name
	os.valueFormatFunc = func(value []T) string {
		return fmt.Sprintf("%v", value)
	}
	return os
}

// InnerTransform applies a transform to each element inside the slice on Set.
func (os *Slice[T]) InnerTransform(transform func(val T) T) *Slice[T] {
	os.transform = func(s []T) []T {
		if s == nil {
			return nil
		}
		res := make([]T, len(s))
		for i, v := range s {
			res[i] = transform(v)
		}
		return res
	}
	return os
}

func (os *Slice[T]) InnerChecks(checks ...func(val T) bool) *Slice[T] {
	for _, check := range checks {
		os.checks = append(os.checks, func(val []T) bool {
			for _, elem := range val {
				if !check(elem) {
					return false
				}
			}
			return true
		})
	}
	return os
}

func (os *Slice[T]) Enum(allowedValues ...T) *Slice[T] {
	os.itemEnumValues = allowedValues
	return os.InnerChecks(func(val T) bool {
		for _, allowed := range allowedValues {
			if reflect.DeepEqual(val, allowed) {
				return true
			}
		}
		return false
	})
}

func (os *Slice[T]) EmptyDefault() *Slice[T] {
	emptySlice := make([]T, 0)
	os.defaultValue = &emptySlice
	return os
}

func (os *Slice[T]) SetAny(value any) error {
	if typedValue, ok := value.([]T); ok {
		os.Set(typedValue)
		return nil
	}

	sliceValue := reflect.ValueOf(value)
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Array {
		return fmt.Errorf("cannot convert %T to []%T", value, *new(T))
	}

	var emptyT T
	targetType := reflect.TypeOf(emptyT)
	result := make([]T, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		elem := sliceValue.Index(i).Interface()

		if typedElem, ok := elem.(T); ok {
			result[i] = typedElem
			continue
		}

		converted, err := tryConvert(elem, targetType)
		if err != nil {
			return fmt.Errorf("cannot convert slice element at index %d: %w", i, err)
		}
		result[i] = converted.Interface().(T)
	}

	os.Set(result)
	return nil
}

func (os *Slice[T]) JSONSchemaType() string {
	return "array"
}

func (os *Slice[T]) JSONSchemaProperty() map[string]any {
	var zero T
	itemType := reflectTypeToJSONSchemaType(reflect.TypeOf(zero))

	items := map[string]any{
		"type": itemType,
	}

	if len(os.itemEnumValues) > 0 {
		enumValues := make([]any, len(os.itemEnumValues))
		for i, v := range os.itemEnumValues {
			enumValues[i] = v
		}
		items["enum"] = enumValues
	}

	property := map[string]any{
		"type":  "array",
		"items": items,
	}

	if os.defaultValue != nil {
		property["default"] = *os.defaultValue
	}

	return property
}
