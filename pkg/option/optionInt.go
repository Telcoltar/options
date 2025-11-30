package option

import "fmt"

// Integer constraint interface for all integer types
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Int represents an option for integer values with type safety.
// T can be any integer type (int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64).
type Int[T Integer] struct {
	baseOption[T, *Int[T]]
}

// NewInt creates a new Int option with the specified name.
func NewInt[T Integer](name string) *Int[T] {
	opt := &Int[T]{}
	opt.initCommon(name, opt)
	return opt
}

// Minimum sets the minimum allowed value for the integer option.
// It adds both a validation check and the corresponding JSON schema property.
func (o *Int[T]) Minimum(min T) *Int[T] {
	o.Check(func(value T) string {
		if value >= min {
			return ""
		}
		return fmt.Sprintf("value %v is less than minimum %v", value, min)
	}, map[string]any{"minimum": min})
	return o
}

// Maximum sets the maximum allowed value for the integer option.
// It adds both a validation check and the corresponding JSON schema property.
func (o *Int[T]) Maximum(max T) *Int[T] {
	o.Check(func(value T) string {
		if value <= max {
			return ""
		}
		return fmt.Sprintf("value %v exceeds maximum %v", value, max)
	}, map[string]any{"maximum": max})
	return o
}
