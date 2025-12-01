package option

import (
	"testing"
)

func TestString_Basic(t *testing.T) {
	opt := NewString("test")
	opt.Set("hello")
	if opt.Get() != "hello" {
		t.Errorf("Expected 'hello', got %s", opt.Get())
	}
}

func TestString_JSONSchemaType(t *testing.T) {
	opt := NewString("test")
	schema := opt.JSONSchema()
	if schema["type"] != "string" {
		t.Errorf("Expected type 'string', got %v", schema["type"])
	}
}

func TestString_MinLength_Valid(t *testing.T) {
	opt := NewString("test").MinLength(3)
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with minLength 3")
	}
}

func TestString_MinLength_Invalid(t *testing.T) {
	opt := NewString("test").MinLength(5)
	opt.Set("hi")
	if opt.IsValid() {
		t.Error("Expected value 'hi' to be invalid with minLength 5")
	}
}

func TestString_MinLength_Boundary(t *testing.T) {
	opt := NewString("test").MinLength(5)
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with minLength 5 (exact length)")
	}
}

func TestString_MaxLength_Valid(t *testing.T) {
	opt := NewString("test").MaxLength(10)
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with maxLength 10")
	}
}

func TestString_MaxLength_Invalid(t *testing.T) {
	opt := NewString("test").MaxLength(3)
	opt.Set("hello")
	if opt.IsValid() {
		t.Error("Expected value 'hello' to be invalid with maxLength 3")
	}
}

func TestString_MaxLength_Boundary(t *testing.T) {
	opt := NewString("test").MaxLength(5)
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with maxLength 5 (exact length)")
	}
}

func TestString_MinLengthMaxLength_Valid(t *testing.T) {
	opt := NewString("test").MinLength(3).MaxLength(10)
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with minLength 3 and maxLength 10")
	}
}

func TestString_MinLengthMaxLength_TooShort(t *testing.T) {
	opt := NewString("test").MinLength(5).MaxLength(10)
	opt.Set("hi")
	if opt.IsValid() {
		t.Error("Expected value 'hi' to be invalid with minLength 5")
	}
}

func TestString_MinLengthMaxLength_TooLong(t *testing.T) {
	opt := NewString("test").MinLength(3).MaxLength(5)
	opt.Set("hello world")
	if opt.IsValid() {
		t.Error("Expected value 'hello world' to be invalid with maxLength 5")
	}
}

func TestString_JSONSchema_MinLength(t *testing.T) {
	opt := NewString("test").MinLength(5)
	schema := opt.JSONSchema()
	if schema["minLength"] != 5 {
		t.Errorf("Expected minLength property to be 5, got %v", schema["minLength"])
	}
}

func TestString_JSONSchema_MaxLength(t *testing.T) {
	opt := NewString("test").MaxLength(100)
	schema := opt.JSONSchema()
	if schema["maxLength"] != 100 {
		t.Errorf("Expected maxLength property to be 100, got %v", schema["maxLength"])
	}
}

func TestString_JSONSchema_MinLengthMaxLength(t *testing.T) {
	opt := NewString("test").MinLength(5).MaxLength(100)
	schema := opt.JSONSchema()
	if schema["minLength"] != 5 {
		t.Errorf("Expected minLength property to be 5, got %v", schema["minLength"])
	}
	if schema["maxLength"] != 100 {
		t.Errorf("Expected maxLength property to be 100, got %v", schema["maxLength"])
	}
}

func TestString_WithDefault(t *testing.T) {
	opt := NewString("test").MinLength(3).MaxLength(10).Default("hello")
	if opt.Get() != "hello" {
		t.Errorf("Expected default value 'hello', got %s", opt.Get())
	}
	if !opt.IsValid() {
		t.Error("Expected default value to be valid")
	}
}

func TestString_EmptyString_MinLength(t *testing.T) {
	opt := NewString("test").MinLength(1)
	opt.Set("")
	if opt.IsValid() {
		t.Error("Expected empty string to be invalid with minLength 1")
	}
}

func TestString_EmptyString_MinLengthZero(t *testing.T) {
	opt := NewString("test").MinLength(0)
	opt.Set("")
	if !opt.IsValid() {
		t.Error("Expected empty string to be valid with minLength 0")
	}
}

func TestString_WithEnum(t *testing.T) {
	opt := NewString("test").MinLength(2).MaxLength(10).Enum("dev", "staging", "prod")
	opt.Set("dev")
	if !opt.IsValid() {
		t.Error("Expected value 'dev' to be valid")
	}

	// Value in length range but not in enum
	opt.Set("test")
	if opt.IsValid() {
		t.Error("Expected value 'test' to be invalid (not in enum)")
	}
}

func TestString_Regex_WithMinLength(t *testing.T) {
	opt := NewString("test").MinLength(3).Regex("^[a-z]+$")
	opt.Set("hello")
	if !opt.IsValid() {
		t.Error("Expected value 'hello' to be valid with minLength 3 and lowercase regex")
	}

	opt.Set("ab")
	if opt.IsValid() {
		t.Error("Expected value 'ab' to be invalid with minLength 3")
	}

	opt.Set("ABC")
	if opt.IsValid() {
		t.Error("Expected value 'ABC' to be invalid (not matching lowercase regex)")
	}
}

func TestString_MinLength_ErrorMessage(t *testing.T) {
	opt := NewString("test").MinLength(5)
	opt.Set("hi")
	errs := opt.Validate("$.test")
	if !errs.HasErrors() {
		t.Fatal("Expected validation errors")
	}
	if len(errs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs.Errors))
	}
	expected := `value "hi" has length 2, which is less than minimum length 5`
	if errs.Errors[0].Message != expected {
		t.Errorf("Expected error message %q, got %q", expected, errs.Errors[0].Message)
	}
}

func TestString_MaxLength_ErrorMessage(t *testing.T) {
	opt := NewString("test").MaxLength(3)
	opt.Set("hello")
	errs := opt.Validate("$.test")
	if !errs.HasErrors() {
		t.Fatal("Expected validation errors")
	}
	if len(errs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs.Errors))
	}
	expected := `value "hello" has length 5, which exceeds maximum length 3`
	if errs.Errors[0].Message != expected {
		t.Errorf("Expected error message %q, got %q", expected, errs.Errors[0].Message)
	}
}
