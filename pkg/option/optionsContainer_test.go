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
