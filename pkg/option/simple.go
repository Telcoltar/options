package option

// Simple represents a basic option type with value, default, and validation checks.
type Simple[T any] struct {
	baseOption[T, *Simple[T]]
}

// NewSimple creates a new Simple option with the specified name.
func NewSimple[T any](name string) *Simple[T] {
	opt := &Simple[T]{}
	opt.initCommon(name, opt)
	return opt
}
