package option

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name *Simple[string]
	Port *Simple[int]
}

func newTestConfig() *testConfig {
	return &testConfig{
		Name: NewSimple[string]("name"),
		Port: NewSimple[int]("port").Default(8080),
	}
}

func TestContainer_ParseAndValidate_Success(t *testing.T) {
	cfg := newTestConfig()
	container := NewContainer("test", cfg)

	data := []byte(`
name: myapp
port: 3000
`)

	err := container.ParseAndValidate(data)
	if err != nil {
		t.Errorf("ParseAndValidate failed: %v", err)
	}

	if cfg.Name.Get() != "myapp" {
		t.Errorf("Expected name 'myapp', got %q", cfg.Name.Get())
	}
	if cfg.Port.Get() != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Port.Get())
	}
}

func TestContainer_ParseAndValidateFile_Success(t *testing.T) {
	cfg := newTestConfig()
	container := NewContainer("test", cfg)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	data := []byte(`
name: fileapp
port: 4000
`)
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	err := container.ParseAndValidateFile(tmpFile)
	if err != nil {
		t.Errorf("ParseAndValidateFile failed: %v", err)
	}

	if cfg.Name.Get() != "fileapp" {
		t.Errorf("Expected name 'fileapp', got %q", cfg.Name.Get())
	}
	if cfg.Port.Get() != 4000 {
		t.Errorf("Expected port 4000, got %d", cfg.Port.Get())
	}
}

func TestContainer_ParseAndValidateFile_ReadError(t *testing.T) {
	cfg := newTestConfig()
	container := NewContainer("test", cfg)

	// Non-existent file
	err := container.ParseAndValidateFile("non_existent_file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestContainer_ParseAndValidate_ParseError(t *testing.T) {
	cfg := newTestConfig()
	container := NewContainer("test", cfg)

	data := []byte(`invalid: yaml: content: [`)

	err := container.ParseAndValidate(data)
	if err == nil {
		t.Error("Expected parse error, got nil")
	}

	// Should not be a ValidationErrors
	var validationErr *ValidationErrors
	if errors.As(err, &validationErr) {
		t.Error("Expected parse error, not ValidationErrors")
	}
}

func TestContainer_ParseAndValidate_ValidationError_CheckFailed(t *testing.T) {
	cfg := &struct {
		Value *Simple[int]
	}{
		Value: NewSimple[int]("value").Checks(func(v int) string {
			if v > 0 {
				return ""
			}
			return "value must be positive"
		}),
	}
	container := NewContainer("test", cfg)

	data := []byte(`value: -5`)

	err := container.ParseAndValidate(data)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	var validationErr *ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationErrors, got %T: %v", err, err)
	}
}

func TestContainer_ParseAndValidate_ValidationError_EnumFailed(t *testing.T) {
	cfg := &struct {
		Color *Simple[string]
	}{
		Color: NewSimple[string]("color").Enum("red", "green", "blue"),
	}
	container := NewContainer("test", cfg)

	data := []byte(`color: yellow`)

	err := container.ParseAndValidate(data)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	var validationErr *ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationErrors, got %T: %v", err, err)
	}
}

func TestContainer_ParseAndValidate_ValidationError_MultipleInvalid(t *testing.T) {
	cfg := &struct {
		A *Simple[int]
		B *Simple[int]
	}{
		A: NewSimple[int]("a").Checks(func(v int) string {
			if v > 0 {
				return ""
			}
			return "value must be positive"
		}),
		B: NewSimple[int]("b").Checks(func(v int) string {
			if v > 0 {
				return ""
			}
			return "value must be positive"
		}),
	}
	container := NewContainer("test", cfg)

	data := []byte(`
a: -1
b: -2
`)

	err := container.ParseAndValidate(data)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	var validationErr *ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationErrors, got %T: %v", err, err)
	}

	// With the new system, both errors should be collected
	if len(validationErr.Errors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d: %v", len(validationErr.Errors), validationErr)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationErrors
		contains string
	}{
		{
			name:     "no errors",
			err:      NewValidationErrors(),
			contains: "validation passed",
		},
		{
			name: "single error",
			err: &ValidationErrors{Errors: []FieldError{
				{Path: "$.value", Message: "required field is missing"},
			}},
			contains: "$.value: required field is missing",
		},
		{
			name: "multiple errors",
			err: &ValidationErrors{Errors: []FieldError{
				{Path: "$.a", Message: "value must be positive"},
				{Path: "$.b", Message: "value must be positive"},
			}},
			contains: "validation failed with 2 errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Expected non-empty error message")
			}
			if !contains(msg, tt.contains) {
				t.Errorf("Expected error to contain %q, got %q", tt.contains, msg)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test isSet tracking and JSON Schema-consistent validation behavior

func TestContainer_UnsetNonRequired_WithRequiredChildren_IsValid(t *testing.T) {
	// JSON Schema behavior: if an object property is not required and not present,
	// the nested required fields are not validated
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name").Required(),
	}
	container := NewContainer("nested", cfg)
	// Container is not required and not set - should be valid even though child is required

	if !container.IsValid() {
		t.Error("Unset non-required container with required children should be valid (JSON Schema semantics)")
	}
}

func TestContainer_UnsetRequired_IsInvalid(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("required", cfg).Required()
	// Container is required but not set - should be invalid

	if container.IsValid() {
		t.Error("Unset required container should be invalid")
	}
}

func TestContainer_SetEmpty_NonRequired_IsValid(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	// Parse empty object
	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !container.IsValid() {
		t.Error("Set empty non-required container should be valid")
	}
}

func TestContainer_SetWithMissingRequiredChildren_IsInvalid(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name").Required(),
	}
	container := NewContainer("test", cfg)

	// Parse empty object - container is set but required child is missing
	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if container.IsValid() {
		t.Error("Set container with missing required children should be invalid")
	}
}

func TestContainer_HasValue_UnsetContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	if container.HasValue() {
		t.Error("Unset container should have HasValue() == false")
	}
}

func TestContainer_HasValue_SetContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !container.HasValue() {
		t.Error("Set container (even empty) should have HasValue() == true")
	}
}

func TestContainer_NotZero_UnsetContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	if container.NotZero() {
		t.Error("Unset container should have NotZero() == false")
	}
}

func TestContainer_NotZero_SetEmptyContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if container.NotZero() {
		t.Error("Set but empty container should have NotZero() == false")
	}
}

func TestContainer_NotZero_SetWithValues(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.Parse([]byte(`name: hello`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !container.NotZero() {
		t.Error("Set container with values should have NotZero() == true")
	}
}

func TestContainer_GetAny_UnsetContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	result := container.GetAny()
	if result != nil {
		t.Errorf("Unset container GetAny() should return nil, got %v", result)
	}
}

func TestContainer_GetAny_SetEmptyContainer(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := container.GetAny()
	if result == nil {
		t.Error("Set empty container GetAny() should return empty map, not nil")
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Errorf("GetAny() should return map[string]any, got %T", result)
	}
	if len(resultMap) != 0 {
		t.Errorf("GetAny() for empty container should return empty map, got %v", resultMap)
	}
}

func TestContainer_GetAny_SetWithValues(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.Parse([]byte(`name: hello`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := container.GetAny()
	if result == nil {
		t.Fatal("Set container GetAny() should not return nil")
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("GetAny() should return map[string]any, got %T", result)
	}
	if resultMap["name"] != "hello" {
		t.Errorf("Expected name='hello', got %v", resultMap["name"])
	}
}

func TestContainer_NestedContainer_UnsetParent_WithRequiredNestedChild(t *testing.T) {
	// Simulates JSON Schema: { "properties": { "nested": { "properties": { "name": {} }, "required": ["name"] } } }
	// If "nested" is not in input, validation passes (nested.name not validated)
	type nestedConfig struct {
		Name *Simple[string]
	}
	type parentConfig struct {
		Nested *Container[nestedConfig]
	}

	nestedCfg := &nestedConfig{
		Name: NewSimple[string]("name").Required(),
	}
	parentCfg := &parentConfig{
		Nested: NewContainer("nested", nestedCfg),
	}
	parent := NewContainer("parent", parentCfg)

	// Parent not set - nested container not validated
	if !parent.IsValid() {
		t.Error("Unset parent with nested required children should be valid (JSON Schema semantics)")
	}
}

func TestContainer_NestedContainer_SetParent_UnsetNested_WithRequiredNestedChild(t *testing.T) {
	// If parent is set but "nested" key is not in input, nested container remains unset
	type nestedConfig struct {
		Name *Simple[string]
	}
	type parentConfig struct {
		Nested *Container[nestedConfig]
	}

	nestedCfg := &nestedConfig{
		Name: NewSimple[string]("name").Required(),
	}
	parentCfg := &parentConfig{
		Nested: NewContainer("nested", nestedCfg),
	}
	parent := NewContainer("parent", parentCfg)

	// Parse parent with no nested key
	err := parent.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Parent is set, nested is not set (and not required) - should be valid
	if !parent.IsValid() {
		t.Error("Set parent with unset non-required nested container should be valid")
	}
}

func TestContainer_NestedContainer_SetParent_SetNestedEmpty_WithRequiredNestedChild(t *testing.T) {
	// If nested is set to {} but has required child, should fail
	type nestedConfig struct {
		Name *Simple[string]
	}
	type parentConfig struct {
		Nested *Container[nestedConfig]
	}

	nestedCfg := &nestedConfig{
		Name: NewSimple[string]("name").Required(),
	}
	parentCfg := &parentConfig{
		Nested: NewContainer("nested", nestedCfg),
	}
	parent := NewContainer("parent", parentCfg)

	// Parse parent with empty nested object
	err := parent.Parse([]byte(`nested: {}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Nested is set but missing required child - should be invalid
	if parent.IsValid() {
		t.Error("Set parent with set nested container missing required child should be invalid")
	}
}

func TestContainer_SetAny_SetsIsSet(t *testing.T) {
	cfg := &struct {
		Name *Simple[string]
	}{
		Name: NewSimple[string]("name"),
	}
	container := NewContainer("test", cfg)

	err := container.SetAny(map[string]any{})
	if err != nil {
		t.Fatalf("SetAny failed: %v", err)
	}

	if !container.HasValue() {
		t.Error("SetAny should set isSet to true")
	}
}

// Tests for ValidationErrors and JSONPath collection

func TestValidate_SimpleField_ReturnsPathInError(t *testing.T) {
	cfg := &struct {
		Port *Int[int]
	}{
		Port: NewInt[int]("port").Maximum(65535),
	}
	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`port: 70000`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation error")
	}

	if len(errs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs.Errors))
	}

	if errs.Errors[0].Path != "$.port" {
		t.Errorf("Expected path '$.port', got %q", errs.Errors[0].Path)
	}
}

func TestValidate_NestedContainer_ReturnsNestedPath(t *testing.T) {
	type serverConfig struct {
		Port *Int[int]
	}
	type rootConfig struct {
		Server *Container[serverConfig]
	}

	serverCfg := &serverConfig{
		Port: NewInt[int]("port").Maximum(65535),
	}
	rootCfg := &rootConfig{
		Server: NewContainer("server", serverCfg),
	}
	container := NewContainer("root", rootCfg)

	err := container.Parse([]byte(`
server:
  port: 70000
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation error")
	}

	if len(errs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs.Errors))
	}

	if errs.Errors[0].Path != "$.server.port" {
		t.Errorf("Expected path '$.server.port', got %q", errs.Errors[0].Path)
	}
}

func TestValidate_Slice_ReturnsIndexInPath(t *testing.T) {
	cfg := &struct {
		Ports *Slice[int]
	}{
		Ports: NewSlice[int]("ports"),
	}
	cfg.Ports.ItemOption.Checks(func(v int) string {
		if v > 0 && v <= 65535 {
			return ""
		}
		return "port must be between 1 and 65535"
	})

	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`
ports:
  - 8080
  - -1
  - 70000
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation errors")
	}

	if len(errs.Errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs.Errors), errs)
	}

	// Check that both invalid items are reported with correct paths
	paths := make(map[string]bool)
	for _, e := range errs.Errors {
		paths[e.Path] = true
	}

	if !paths["$.ports[1]"] {
		t.Error("Expected error for $.ports[1]")
	}
	if !paths["$.ports[2]"] {
		t.Error("Expected error for $.ports[2]")
	}
}

func TestValidate_Map_ReturnsKeyInPath(t *testing.T) {
	cfg := &struct {
		Settings *Map[int]
	}{
		Settings: NewMap[int]("settings"),
	}
	cfg.Settings.ValueOption.Checks(func(v int) string {
		if v >= 0 {
			return ""
		}
		return "value must be non-negative"
	})

	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`
settings:
  timeout: 30
  retries: -1
  maxConnections: -5
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation errors")
	}

	if len(errs.Errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs.Errors), errs)
	}

	// Check that both invalid items are reported with correct paths
	paths := make(map[string]bool)
	for _, e := range errs.Errors {
		paths[e.Path] = true
	}

	if !paths["$.settings.retries"] {
		t.Error("Expected error for $.settings.retries")
	}
	if !paths["$.settings.maxConnections"] {
		t.Error("Expected error for $.settings.maxConnections")
	}
}

func TestValidate_SliceOfContainers_ReturnsFullPath(t *testing.T) {
	type userConfig struct {
		Email *String
		Age   *Int[int]
		*Container[userConfig]
	}

	newUserConfig := func() *userConfig {
		u := &userConfig{
			Email: NewString("email").Required(),
			Age:   NewInt[int]("age").Minimum(0).Maximum(150),
		}
		u.Container = NewContainer("user", u)
		return u
	}

	cfg := &struct {
		Users *Slice[*userConfig]
	}{
		Users: NewSlice[*userConfig]("users", newUserConfig),
	}

	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`
users:
  - email: "alice@example.com"
    age: 25
  - age: 200
  - email: "bob@example.com"
    age: -5
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation errors")
	}

	// Should have 3 errors:
	// - $.users[1].email: required
	// - $.users[1].age: exceeds max
	// - $.users[2].age: below min
	if len(errs.Errors) != 3 {
		t.Fatalf("Expected 3 errors, got %d: %v", len(errs.Errors), errs)
	}

	paths := make(map[string]bool)
	for _, e := range errs.Errors {
		paths[e.Path] = true
	}

	if !paths["$.users[1].email"] {
		t.Error("Expected error for $.users[1].email")
	}
	if !paths["$.users[1].age"] {
		t.Error("Expected error for $.users[1].age")
	}
	if !paths["$.users[2].age"] {
		t.Error("Expected error for $.users[2].age")
	}
}

func TestValidate_CollectsAllErrors_NotJustFirst(t *testing.T) {
	cfg := &struct {
		A *Int[int]
		B *Int[int]
		C *Int[int]
	}{
		A: NewInt[int]("a").Minimum(0),
		B: NewInt[int]("b").Minimum(0),
		C: NewInt[int]("c").Minimum(0),
	}

	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`
a: -1
b: -2
c: -3
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")

	// All 3 errors should be collected
	if len(errs.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d: %v", len(errs.Errors), errs)
	}
}

func TestValidate_RequiredField_IncludesDescriptiveMessage(t *testing.T) {
	cfg := &struct {
		Name *String
	}{
		Name: NewString("name").Required(),
	}

	container := NewContainer("config", cfg)

	err := container.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation error")
	}

	if errs.Errors[0].Path != "$.name" {
		t.Errorf("Expected path '$.name', got %q", errs.Errors[0].Path)
	}

	if !contains(errs.Errors[0].Message, "required") {
		t.Errorf("Expected message to contain 'required', got %q", errs.Errors[0].Message)
	}
}

func TestValidate_ContainerCheck_ReturnsErrorMessage(t *testing.T) {
	type rangeConfig struct {
		Min *Int[int]
		Max *Int[int]
		*Container[rangeConfig]
	}

	cfg := &rangeConfig{
		Min: NewInt[int]("min"),
		Max: NewInt[int]("max"),
	}
	cfg.Container = NewContainer("range", cfg)
	cfg.Container.Check(func(r *rangeConfig) string {
		if r.Min.HasValue() && r.Max.HasValue() && r.Min.Get() > r.Max.Get() {
			return "min must be less than or equal to max"
		}
		return ""
	}, nil)

	err := cfg.Container.Parse([]byte(`
min: 10
max: 5
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	errs := cfg.Container.Validate("$")
	if !errs.HasErrors() {
		t.Fatal("Expected validation error")
	}

	if errs.Errors[0].Path != "$" {
		t.Errorf("Expected path '$', got %q", errs.Errors[0].Path)
	}

	if !contains(errs.Errors[0].Message, "min must be less than or equal to max") {
		t.Errorf("Expected specific error message, got %q", errs.Errors[0].Message)
	}
}
