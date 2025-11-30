package option

import "fmt"

// Float constraint interface for all floating-point types
type Float interface {
	~float32 | ~float64
}

// Number represents an option for floating-point number values with type safety.
// T can be any float type (float32, float64).
type Number[T Float] struct {
	baseOption[T, *Number[T]]
}

// NewNumber creates a new Number option with the specified name.
func NewNumber[T Float](name string) *Number[T] {
	opt := &Number[T]{}
	opt.initCommon(name, opt)
	return opt
}

// Minimum sets the minimum allowed value for the number option.
// It adds both a validation check and the corresponding JSON schema property.
func (o *Number[T]) Minimum(min T) *Number[T] {
	o.Check(func(value T) string {
		if value >= min {
			return ""
		}
		return fmt.Sprintf("value %v is less than minimum %v", value, min)
	}, map[string]any{"minimum": min})
	return o
}

// Maximum sets the maximum allowed value for the number option.
// It adds both a validation check and the corresponding JSON schema property.
func (o *Number[T]) Maximum(max T) *Number[T] {
	o.Check(func(value T) string {
		if value <= max {
			return ""
		}
		return fmt.Sprintf("value %v exceeds maximum %v", value, max)
	}, map[string]any{"maximum": max})
	return o
}
