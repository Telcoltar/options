package option

import (
	"fmt"
	"maps"
	"reflect"
	"strconv"
)

func tryConvert(value any, targetType reflect.Type) (reflect.Value, error) {
	// Handle nil values
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	inputValue := reflect.ValueOf(value)

	if inputValue.Type().AssignableTo(targetType) {
		return inputValue, nil
	}

	if targetType.Kind() == reflect.String && inputValue.Kind() == reflect.String {
		if inputValue.Type().ConvertibleTo(targetType) {
			return inputValue.Convert(targetType), nil
		}
		return inputValue, nil
	}

	if targetType.Kind() == reflect.String {
		strValue := reflect.ValueOf(fmt.Sprintf("%v", value))
		if strValue.Type().ConvertibleTo(targetType) {
			return strValue.Convert(targetType), nil
		}
		return strValue, nil
	}

	if inputValue.Type().ConvertibleTo(targetType) {
		return inputValue.Convert(targetType), nil
	}

	if inputValue.Kind() == reflect.String {
		str := inputValue.String()

		switch targetType.Kind() {
		case reflect.Bool:
			b, err := strconv.ParseBool(str)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse %q as bool: %w", str, err)
			}
			return reflect.ValueOf(b), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse %q as int: %w", str, err)
			}
			return reflect.ValueOf(i).Convert(targetType), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse %q as uint: %w", str, err)
			}
			return reflect.ValueOf(u).Convert(targetType), nil
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse %q as float: %w", str, err)
			}
			return reflect.ValueOf(f).Convert(targetType), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", value, targetType)
}

type OptionGet[T any] interface {
	Get() T
	HasValue() bool
}

type baseOption[T any, Self any] struct {
	defaultValue         *T
	jsonSchemaProperties map[string]any
	jsonSchemaType       string
	name                 string
	value                *T
	valueFormatFunc      func(value T) string
	checks               []func(value T) bool
	transform            func(value T) T
	self                 Self
}

// initCommon sets common defaults for all option types
func (o *baseOption[T, Self]) initCommon(name string, self Self) {
	o.self = self
	o.name = name
	o.jsonSchemaProperties = make(map[string]any)
	o.jsonSchemaType = getJSONSchemaType[T]()
	o.valueFormatFunc = func(value T) string {
		return fmt.Sprintf("%v", value)
	}
}

func getJSONSchemaType[T any]() string {
	t := reflect.TypeFor[T]()
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "string"
	}
}

func (o *baseOption[T, Self]) JSONSchema() map[string]any {
	prop := map[string]any{
		"type": o.jsonSchemaType,
	}
	maps.Copy(prop, o.jsonSchemaProperties)
	return prop
}

// Get returns in order: value, defaultValue, empty Value of Type T
func (o *baseOption[T, Self]) Get() T {
	if o.value != nil {
		return *o.value
	}
	if o.defaultValue != nil {
		return *o.defaultValue
	}
	return *new(T)
}

// Returns if value or defaultValue is set think != nil
func (o *baseOption[T, Self]) HasValue() bool {
	return o.value != nil || o.defaultValue != nil
}

// NotZero returns true if the option has a value and that value is not the zero value of T
func (o *baseOption[T, Self]) NotZero() bool {
	if !o.HasValue() {
		return false
	}
	var zero T
	return !reflect.DeepEqual(o.Get(), zero)
}

// GetPointer returns in order: &value, &defaultValue, nil
func (o *baseOption[T, Self]) GetPointer() *T {
	if o.value != nil {
		return o.value
	}
	if o.defaultValue != nil {
		return o.defaultValue
	}
	return nil
}

func (o *baseOption[T, Self]) StrValue() string {
	return o.valueFormatFunc(o.Get())
}

func (o *baseOption[T, Self]) Name() string {
	return o.name
}

// IsValid checks if a set value passes all check function
// if no value is set, but a defaultValue return true
// if neither is set, return false
func (o *baseOption[T, Self]) IsValid() bool {
	if o.value != nil {
		for _, check := range o.checks {
			if !check(*o.value) {
				return false
			}
		}
		return true
	}
	if o.defaultValue != nil {
		return true
	}
	return false
}

// GetValid returns Get(),IsValid()
func (o *baseOption[T, Self]) GetValid() (T, bool) {
	return o.Get(), o.IsValid()
}

func (o *baseOption[T, Self]) Default(value T) Self {
	o.defaultValue = &value
	o.jsonSchemaProperties["default"] = value
	return o.self
}

func (o *baseOption[T, Self]) EmptyDefault() Self {
	o.Default(*new(T))
	return o.self
}

func (o *baseOption[T, Self]) Checks(checks ...func(value T) bool) Self {
	o.checks = append(o.checks, checks...)
	return o.self
}

func (o *baseOption[T, Self]) Transform(transform func(value T) T) Self {
	o.transform = transform
	return o.self
}

func (o *baseOption[T, Self]) Enum(allowedValues ...T) Self {
	o.jsonSchemaProperties["enum"] = allowedValues
	o.checks = append(o.checks, func(value T) bool {
		for _, allowed := range allowedValues {
			if reflect.DeepEqual(value, allowed) {
				return true
			}
		}
		return false
	})
	return o.self
}

func (o *baseOption[T, Self]) Set(value T) {
	if o.transform != nil {
		value = o.transform(value)
	}
	o.value = &value
}

func (o *baseOption[T, Self]) SetAny(value any) error {
	if typedValue, ok := value.(T); ok {
		o.Set(typedValue)
		return nil
	}

	var result T
	targetType := reflect.TypeOf(result)

	converted, err := tryConvert(value, targetType)
	if err != nil {
		return err
	}

	if !converted.Type().AssignableTo(targetType) {
		return fmt.Errorf("converted value type %v is not assignable to %v", converted.Type(), targetType)
	}

	o.Set(converted.Interface().(T))
	return nil
}
