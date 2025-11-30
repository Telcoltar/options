package option

// OptionInterface defines the interface that all option types must implement.
// It provides methods for name retrieval, validation, string conversion, setting values,
// and JSON Schema generation.
type OptionInterface interface {
	GetName() string
	IsValid() bool
	IsRequired() bool
	SetAny(value any) error
	GetAny() any
	JSONSchema() map[string]any
	// HasValue returns true if the option has a value (either explicitly set or via default).
	HasValue() bool
	// NotZero returns true if the option has a value and that value is not the zero value.
	NotZero() bool
	// Validate performs validation and returns all errors with their JSONPath locations.
	// The path parameter is the JSONPath prefix for this option (e.g., "$.config" or "$").
	Validate(path string) *ValidationErrors
}

type TypedOptionInterface[T any] interface {
	OptionInterface
	Get() T
}
