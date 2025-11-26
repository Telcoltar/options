package option

// OptionInterface defines the interface that all option types must implement.
// It provides methods for name retrieval, validation, string conversion, setting values,
// and JSON Schema generation.
type OptionInterface interface {
	GetName() string
	IsValid() bool
	SetAny(value any) error
	GetAny() any
	JSONSchema() map[string]any
}

type TypedOptionInterface[T any] interface {
	OptionInterface
	Get() T
}
