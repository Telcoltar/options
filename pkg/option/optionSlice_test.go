package option

import (
	"testing"
)

func TestOptionSlice_SetAny_SameType(t *testing.T) {
	opt := NewSlice[int]("test")
	err := opt.SetAny([]int{1, 2, 3})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected [1 2 3], got %v", result)
	}
}

func TestOptionSlice_SetAny_ConvertibleTypes(t *testing.T) {
	opt := NewSlice[int64]("test")
	err := opt.SetAny([]int32{1, 2, 3})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected [1 2 3], got %v", result)
	}
}

func TestOptionSlice_SetAny_InterfaceSliceToInt(t *testing.T) {
	opt := NewSlice[int]("test")
	err := opt.SetAny([]any{1, 2, 3})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected [1 2 3], got %v", result)
	}
}

func TestOptionSlice_SetAny_MixedTypesToString(t *testing.T) {
	opt := NewSlice[string]("test")
	err := opt.SetAny([]any{"hello", 42, true, 3.14})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	expected := []string{"hello", "42", "true", "3.14"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected result[%d] = %q, got %q", i, v, result[i])
		}
	}
}

func TestOptionSlice_SetAny_StringSliceToInt(t *testing.T) {
	opt := NewSlice[int]("test")
	err := opt.SetAny([]any{"1", "2", "3"})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected [1 2 3], got %v", result)
	}
}

func TestOptionSlice_SetAny_StringSliceToBool(t *testing.T) {
	opt := NewSlice[bool]("test")
	err := opt.SetAny([]any{"true", "false", "1", "0"})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	expected := []bool{true, false, true, false}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected result[%d] = %v, got %v", i, v, result[i])
		}
	}
}

func TestOptionSlice_SetAny_InvalidStringToInt(t *testing.T) {
	opt := NewSlice[int]("test")
	err := opt.SetAny([]any{"1", "not-a-number", "3"})
	if err == nil {
		t.Error("Expected error for invalid int string, got nil")
	}
}

func TestOptionSlice_SetAny_InvalidType(t *testing.T) {
	opt := NewSlice[int]("test")
	err := opt.SetAny("not-a-slice")
	if err == nil {
		t.Error("Expected error for non-slice type, got nil")
	}
}

func TestOptionSlice_InnerChecks(t *testing.T) {
	opt := NewSlice[int]("test").InnerChecks(func(val int) bool {
		return val > 0
	})

	opt.Set([]int{1, 2, 3})
	if !opt.IsValid() {
		t.Error("Expected valid for all positive values")
	}

	opt.Set([]int{1, -1, 3})
	if opt.IsValid() {
		t.Error("Expected invalid for negative value")
	}
}

func TestOptionSlice_Enum_ValidValues(t *testing.T) {
	opt := NewSlice[string]("test").ItemOption.Enum("IPv4", "IPv6")
	opt.Set([]string{"IPv4", "IPv6"})
	if !opt.IsValid() {
		t.Error("Expected valid for allowed values")
	}
}

func TestOptionSlice_Enum_InvalidValue(t *testing.T) {
	opt := NewSlice[string]("test").ItemOption.Enum("IPv4", "IPv6")
	opt.Set([]string{"IPv4", "IPv7"})
	if opt.IsValid() {
		t.Error("Expected invalid for disallowed value")
	}
}

func TestOptionSlice_Enum_IntValues(t *testing.T) {
	opt := NewSlice[int]("test").ItemOption.Enum(1, 2, 3, 5, 8)
	opt.Set([]int{1, 3, 5})
	if !opt.IsValid() {
		t.Error("Expected valid for allowed values")
	}
	opt.Set([]int{1, 4, 5})
	if opt.IsValid() {
		t.Error("Expected invalid for disallowed value 4")
	}
}

func TestOptionSlice_Enum_EmptySlice(t *testing.T) {
	opt := NewSlice[string]("test").ItemOption.Enum("red", "green", "blue")
	opt.Set([]string{})
	if !opt.IsValid() {
		t.Error("Expected valid for empty slice")
	}
}
