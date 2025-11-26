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

	// Should not be a ValidationError
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		t.Error("Expected parse error, not ValidationError")
	}
}

func TestContainer_ParseAndValidate_ValidationError_CheckFailed(t *testing.T) {
	cfg := &struct {
		Value *Simple[int]
	}{
		Value: NewSimple[int]("value").Checks(func(v int) bool { return v > 0 }),
	}
	container := NewContainer("test", cfg)

	data := []byte(`value: -5`)

	err := container.ParseAndValidate(data)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
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

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}
}

func TestContainer_ParseAndValidate_ValidationError_MultipleInvalid(t *testing.T) {
	cfg := &struct {
		A *Simple[int]
		B *Simple[int]
	}{
		A: NewSimple[int]("a").Checks(func(v int) bool { return v > 0 }),
		B: NewSimple[int]("b").Checks(func(v int) bool { return v > 0 }),
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

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		contains string
	}{
		{
			name:     "no invalid options",
			err:      &ValidationError{Container: "test"},
			contains: `validation failed for container "test"`,
		},
		{
			name:     "with invalid options (currently unused)",
			err:      &ValidationError{Container: "test", InvalidOptions: []string{}},
			contains: `validation failed for container "test"`,
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
