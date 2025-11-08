package option

import (
	"testing"
)

func TestOption_SetAny_SameType(t *testing.T) {
	opt := NewSimple[int]("test")
	err := opt.SetAny(42)
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != 42 {
		t.Errorf("Expected 42, got %d", opt.Get())
	}
}

func TestOption_SetAny_ConvertibleTypes(t *testing.T) {
	opt := NewSimple[int64]("test")
	err := opt.SetAny(int32(42))
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != 42 {
		t.Errorf("Expected 42, got %d", opt.Get())
	}
}

func TestOption_SetAny_StringToInt(t *testing.T) {
	opt := NewSimple[int]("test")
	err := opt.SetAny("123")
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != 123 {
		t.Errorf("Expected 123, got %d", opt.Get())
	}
}

func TestOption_SetAny_StringToBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"True", true},
		{"False", false},
	}

	for _, tt := range tests {
		opt := NewSimple[bool]("test")
		err := opt.SetAny(tt.input)
		if err != nil {
			t.Errorf("SetAny(%q) failed: %v", tt.input, err)
			continue
		}
		if opt.Get() != tt.expected {
			t.Errorf("SetAny(%q): expected %v, got %v", tt.input, tt.expected, opt.Get())
		}
	}
}

func TestOption_SetAny_StringToFloat(t *testing.T) {
	opt := NewSimple[float64]("test")
	err := opt.SetAny("3.14")
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != 3.14 {
		t.Errorf("Expected 3.14, got %f", opt.Get())
	}
}

func TestOption_SetAny_IntToString(t *testing.T) {
	opt := NewSimple[string]("test")
	err := opt.SetAny(42)
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != "42" {
		t.Errorf("Expected '42', got %q", opt.Get())
	}
}

func TestOption_SetAny_BoolToString(t *testing.T) {
	opt := NewSimple[string]("test")
	err := opt.SetAny(true)
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != "true" {
		t.Errorf("Expected 'true', got %q", opt.Get())
	}
}

func TestOption_SetAny_InvalidStringToInt(t *testing.T) {
	opt := NewSimple[int]("test")
	err := opt.SetAny("not-a-number")
	if err == nil {
		t.Error("Expected error for invalid int string, got nil")
	}
}

func TestOption_SetAny_InvalidStringToBool(t *testing.T) {
	opt := NewSimple[bool]("test")
	err := opt.SetAny("not-a-bool")
	if err == nil {
		t.Error("Expected error for invalid bool string, got nil")
	}
}

func TestOption_SetAny_CustomStringType(t *testing.T) {
	type CustomString string
	opt := NewSimple[CustomString]("test")
	err := opt.SetAny("hello")
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != "hello" {
		t.Errorf("Expected 'hello', got %q", opt.Get())
	}
}

func TestOption_SetAny_IntToCustomStringType(t *testing.T) {
	type CustomString string
	opt := NewSimple[CustomString]("test")
	err := opt.SetAny(42)
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	if opt.Get() != "42" {
		t.Errorf("Expected '42', got %q", opt.Get())
	}
}

func TestOption_Enum_ValidValue(t *testing.T) {
	opt := NewSimple[string]("test").Enum("red", "green", "blue")
	opt.Set("red")
	if !opt.IsValid() {
		t.Error("Expected 'red' to be valid")
	}
	if opt.Get() != "red" {
		t.Errorf("Expected 'red', got %q", opt.Get())
	}
}

func TestOption_Enum_InvalidValue(t *testing.T) {
	opt := NewSimple[string]("test").Enum("red", "green", "blue")
	opt.Set("yellow")
	if opt.IsValid() {
		t.Error("Expected 'yellow' to be invalid")
	}
}

func TestOption_Enum_IntValues(t *testing.T) {
	opt := NewSimple[int]("test").Enum(1, 2, 3, 5, 8)
	opt.Set(5)
	if !opt.IsValid() {
		t.Error("Expected 5 to be valid")
	}
	opt.Set(4)
	if opt.IsValid() {
		t.Error("Expected 4 to be invalid")
	}
}

func TestOption_Enum_EmptyList(t *testing.T) {
	opt := NewSimple[string]("test").Enum()
	opt.Set("anything")
	if opt.IsValid() {
		t.Error("Expected any value to be invalid with empty enum")
	}
}

func TestOption_Enum_WithDefault(t *testing.T) {
	opt := NewSimple[string]("test").Enum("red", "green", "blue").Default("green")
	if !opt.IsValid() {
		t.Error("Expected default 'green' to be valid")
	}
	if opt.Get() != "green" {
		t.Errorf("Expected 'green', got %q", opt.Get())
	}
}
