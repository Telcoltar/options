package option

import (
	"testing"
)

func TestOptionMap_SetAny_SameType(t *testing.T) {
	opt := NewMap[int]("test")
	err := opt.SetAny(map[string]int{"a": 1, "b": 2})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 2 || result["a"] != 1 || result["b"] != 2 {
		t.Errorf("Expected map[a:1 b:2], got %v", result)
	}
}

func TestOptionMap_SetAny_ConvertibleTypes(t *testing.T) {
	opt := NewMap[int64]("test")
	err := opt.SetAny(map[string]int32{"a": 1, "b": 2})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 2 || result["a"] != 1 || result["b"] != 2 {
		t.Errorf("Expected map[a:1 b:2], got %v", result)
	}
}

func TestOptionMap_SetAny_InterfaceMapToString(t *testing.T) {
	opt := NewMap[string]("test")
	err := opt.SetAny(map[string]any{
		"name":    "grafana",
		"version": 10,
		"enabled": true,
		"port":    3000,
	})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	expected := map[string]string{
		"name":    "grafana",
		"version": "10",
		"enabled": "true",
		"port":    "3000",
	}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Expected result[%q] = %q, got %q", k, v, result[k])
		}
	}
}

func TestOptionMap_SetAny_IntToString(t *testing.T) {
	opt := NewMap[string]("test")
	err := opt.SetAny(map[string]any{
		"port":  8080,
		"count": 42,
	})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if result["port"] != "8080" || result["count"] != "42" {
		t.Errorf("Expected map[port:8080 count:42], got %v", result)
	}
}

func TestOptionMap_SetAny_StringMapToInt(t *testing.T) {
	opt := NewMap[int]("test")
	err := opt.SetAny(map[string]any{
		"a": "1",
		"b": "2",
		"c": "3",
	})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if len(result) != 3 || result["a"] != 1 || result["b"] != 2 || result["c"] != 3 {
		t.Errorf("Expected map[a:1 b:2 c:3], got %v", result)
	}
}

func TestOptionMap_SetAny_StringMapToBool(t *testing.T) {
	opt := NewMap[bool]("test")
	err := opt.SetAny(map[string]any{
		"enabled":  "true",
		"disabled": "false",
		"on":       "1",
		"off":      "0",
	})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if !result["enabled"] || result["disabled"] || !result["on"] || result["off"] {
		t.Errorf("Unexpected bool conversion results: %v", result)
	}
}

func TestOptionMap_SetAny_InvalidStringToInt(t *testing.T) {
	opt := NewMap[int]("test")
	err := opt.SetAny(map[string]any{
		"valid":   "123",
		"invalid": "not-a-number",
	})
	if err == nil {
		t.Error("Expected error for invalid int string, got nil")
	}
}

func TestOptionMap_SetAny_InvalidType(t *testing.T) {
	opt := NewMap[int]("test")
	err := opt.SetAny("not-a-map")
	if err == nil {
		t.Error("Expected error for non-map type, got nil")
	}
}

func TestOptionMap_SetAny_NonStringKeys(t *testing.T) {
	opt := NewMap[int]("test")
	err := opt.SetAny(map[int]int{1: 100, 2: 200})
	if err == nil {
		t.Error("Expected error for non-string keys, got nil")
	}
}

func TestOptionMap_InnerChecks(t *testing.T) {
	opt := NewMap[int]("test").InnerChecks(func(val int) bool {
		return val > 0
	})

	opt.Set(map[string]int{"a": 1, "b": 2})
	if !opt.IsValid() {
		t.Error("Expected valid for all positive values")
	}

	opt.Set(map[string]int{"a": 1, "b": -1})
	if opt.IsValid() {
		t.Error("Expected invalid for negative value")
	}
}

func TestOptionMap_KeyChecks(t *testing.T) {
	opt := NewMap[int]("test").KeyChecks(func(key string) bool {
		return len(key) > 0 && key[0] != '_'
	})

	opt.Set(map[string]int{"valid": 1, "another": 2})
	if !opt.IsValid() {
		t.Error("Expected valid for keys without underscore prefix")
	}

	opt.Set(map[string]int{"valid": 1, "_invalid": 2})
	if opt.IsValid() {
		t.Error("Expected invalid for key with underscore prefix")
	}
}

func TestOptionMap_SetAny_CustomStringType(t *testing.T) {
	type CustomString string
	opt := NewMap[CustomString]("test")
	err := opt.SetAny(map[string]any{
		"name": "hello",
		"num":  42,
	})
	if err != nil {
		t.Errorf("SetAny failed: %v", err)
	}
	result := opt.Get()
	if result["name"] != "hello" || result["num"] != "42" {
		t.Errorf("Expected map[name:hello num:42], got %v", result)
	}
}
