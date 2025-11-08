package option

import (
	"fmt"
	"maps"
	"reflect"
)

// Slice represents an option type that holds a slice of values.
type Slice[T any] struct {
	baseOption[[]T, *Slice[T]]
	ItemOption *baseOption[T, *Slice[T]]
}

// NewSlice creates a new Slice option with the specified name.
func NewSlice[T any](name string) *Slice[T] {
	os := &Slice[T]{}
	os.initCommon(name, os)
	os.ItemOption = &baseOption[T, *Slice[T]]{}
	os.ItemOption.initCommon("", os)
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

// IsValid checks if a set value passes all check function
// if no value is set, but a defaultValue return true
// if neither is set, return false
func (os *Slice[T]) IsValid() bool {
	if !os.baseOption.IsValid() {
		return false
	}
	if os.value != nil {
		for _, elem := range *os.value {
			os.ItemOption.Set(elem)
			if !os.ItemOption.IsValid() {
				return false
			}
		}
	}
	return true
}

func (os *Slice[T]) EmptyDefault() *Slice[T] {
	return os.Default(make([]T, 0))
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

	result := make([]T, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		elem := sliceValue.Index(i).Interface()

		if err := os.ItemOption.SetAny(elem); err != nil {
			return err
		}

		result[i] = os.ItemOption.Get()
	}

	os.Set(result)
	return nil
}

func (os *Slice[T]) JSONSchema() map[string]any {
	items := os.ItemOption.JSONSchema()

	property := map[string]any{
		"type":  "array",
		"items": items,
	}

	maps.Copy(property, os.jsonSchemaProperties)

	return property
}
