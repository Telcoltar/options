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
	factory    func() T
}

// NewSlice creates a new Slice option with the specified name.
func NewSlice[T any](name string, factory ...func() T) *Slice[T] {
	os := &Slice[T]{}
	os.initCommon(name, os)
	if len(factory) > 0 {
		os.factory = factory[0]
	} else {
		os.ItemOption = &baseOption[T, *Slice[T]]{}
		os.ItemOption.initCommon("", os)
	}
	return os
}

// IsValid checks if a set value passes all check function
// if no value is set, but a defaultValue return true
// if neither is set, return false
func (os *Slice[T]) IsValid() bool {
	return !os.Validate("$").HasErrors()
}

// Validate performs validation and returns all errors with their JSONPath locations.
// The path parameter is the JSONPath prefix for this option (e.g., "$.items" or "$").
func (os *Slice[T]) Validate(path string) *ValidationErrors {
	errs := NewValidationErrors()

	// Validate base option (required check, custom checks on the slice itself)
	baseErrs := os.baseOption.Validate(path)
	errs.Merge(baseErrs, "")

	if os.value != nil {
		for i, elem := range *os.value {
			elemPath := JoinPath(path, IndexPath(i))
			if os.factory != nil {
				if opt, ok := any(elem).(OptionInterface); ok {
					elemErrs := opt.Validate(elemPath)
					errs.Merge(elemErrs, "")
				}
			} else {
				os.ItemOption.Set(elem)
				elemErrs := os.ItemOption.Validate(elemPath)
				errs.Merge(elemErrs, "")
			}
		}
	}
	return errs
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

		if os.factory != nil {
			newElem := os.factory()
			if opt, ok := any(newElem).(OptionInterface); ok {
				if err := opt.SetAny(elem); err != nil {
					return err
				}
				result[i] = newElem
			} else {
				return fmt.Errorf("factory returned type %T which does not implement CommonInterface", newElem)
			}
		} else {
			if err := os.ItemOption.SetAny(elem); err != nil {
				return err
			}

			result[i] = os.ItemOption.Get()
		}
	}

	os.Set(result)
	return nil
}

func (os *Slice[T]) JSONSchema() map[string]any {
	var items map[string]any
	if os.factory != nil {
		newElem := os.factory()
		if opt, ok := any(newElem).(OptionInterface); ok {
			items = opt.JSONSchema()
		} else {
			items = make(map[string]any)
		}
	} else {
		items = os.ItemOption.JSONSchema()
	}

	property := map[string]any{
		"type":  "array",
		"items": items,
	}

	maps.Copy(property, os.jsonSchemaProperties)

	return property
}

// GetAny returns the slice value as any, or nil if no value is set.
// If using a factory for nested options, recursively calls GetAny() on each element.
func (os *Slice[T]) GetAny() any {
	if !os.HasValue() {
		return nil
	}

	slice := os.Get()

	// If using factory, elements are OptionInterface - recurse
	if os.factory != nil {
		result := make([]any, len(slice))
		for i, elem := range slice {
			if opt, ok := any(elem).(OptionInterface); ok {
				result[i] = opt.GetAny()
			} else {
				result[i] = elem
			}
		}
		return result
	}

	// No factory - return the slice directly
	return slice
}
