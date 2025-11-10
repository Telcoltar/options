package option

import (
	"testing"
)

func TestNumber_Basic(t *testing.T) {
	opt := NewNumber[float64]("test")
	opt.Set(3.14)
	if opt.Get() != 3.14 {
		t.Errorf("Expected 3.14, got %f", opt.Get())
	}
}

func TestNumber_JSONSchemaType(t *testing.T) {
	opt := NewNumber[float64]("test")
	schema := opt.JSONSchema()
	if schema["type"] != "number" {
		t.Errorf("Expected type 'number', got %v", schema["type"])
	}
}

func TestNumber_Minimum_Valid(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(0.0)
	opt.Set(5.5)
	if !opt.IsValid() {
		t.Error("Expected value 5.5 to be valid with minimum 0.0")
	}
}

func TestNumber_Minimum_Invalid(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.0)
	opt.Set(5.5)
	if opt.IsValid() {
		t.Error("Expected value 5.5 to be invalid with minimum 10.0")
	}
}

func TestNumber_Minimum_Boundary(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.5)
	opt.Set(10.5)
	if !opt.IsValid() {
		t.Error("Expected value 10.5 to be valid with minimum 10.5")
	}
}

func TestNumber_Maximum_Valid(t *testing.T) {
	opt := NewNumber[float64]("test").Maximum(100.0)
	opt.Set(50.5)
	if !opt.IsValid() {
		t.Error("Expected value 50.5 to be valid with maximum 100.0")
	}
}

func TestNumber_Maximum_Invalid(t *testing.T) {
	opt := NewNumber[float64]("test").Maximum(100.0)
	opt.Set(150.5)
	if opt.IsValid() {
		t.Error("Expected value 150.5 to be invalid with maximum 100.0")
	}
}

func TestNumber_Maximum_Boundary(t *testing.T) {
	opt := NewNumber[float64]("test").Maximum(100.5)
	opt.Set(100.5)
	if !opt.IsValid() {
		t.Error("Expected value 100.5 to be valid with maximum 100.5")
	}
}

func TestNumber_MinimumMaximum_Valid(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(0.0).Maximum(100.0)
	opt.Set(50.5)
	if !opt.IsValid() {
		t.Error("Expected value 50.5 to be valid with minimum 0.0 and maximum 100.0")
	}
}

func TestNumber_MinimumMaximum_TooLow(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.0).Maximum(100.0)
	opt.Set(5.5)
	if opt.IsValid() {
		t.Error("Expected value 5.5 to be invalid with minimum 10.0")
	}
}

func TestNumber_MinimumMaximum_TooHigh(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.0).Maximum(100.0)
	opt.Set(150.5)
	if opt.IsValid() {
		t.Error("Expected value 150.5 to be invalid with maximum 100.0")
	}
}

func TestNumber_JSONSchema_Minimum(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.5)
	schema := opt.JSONSchema()
	if schema["minimum"] != 10.5 {
		t.Errorf("Expected minimum property to be 10.5, got %v", schema["minimum"])
	}
}

func TestNumber_JSONSchema_Maximum(t *testing.T) {
	opt := NewNumber[float64]("test").Maximum(100.5)
	schema := opt.JSONSchema()
	if schema["maximum"] != 100.5 {
		t.Errorf("Expected maximum property to be 100.5, got %v", schema["maximum"])
	}
}

func TestNumber_JSONSchema_MinimumMaximum(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(10.5).Maximum(100.5)
	schema := opt.JSONSchema()
	if schema["minimum"] != 10.5 {
		t.Errorf("Expected minimum property to be 10.5, got %v", schema["minimum"])
	}
	if schema["maximum"] != 100.5 {
		t.Errorf("Expected maximum property to be 100.5, got %v", schema["maximum"])
	}
}

func TestNumber_Float32(t *testing.T) {
	opt := NewNumber[float32]("test").Minimum(10.5).Maximum(100.5)
	opt.Set(50.5)
	if !opt.IsValid() {
		t.Error("Expected float32 value to be valid")
	}

	opt.Set(5.5)
	if opt.IsValid() {
		t.Error("Expected float32 value 5.5 to be invalid with minimum 10.5")
	}

	opt.Set(150.5)
	if opt.IsValid() {
		t.Error("Expected float32 value 150.5 to be invalid with maximum 100.5")
	}
}

func TestNumber_WithDefault(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(0.0).Maximum(100.0).Default(50.5)
	if opt.Get() != 50.5 {
		t.Errorf("Expected default value 50.5, got %f", opt.Get())
	}
	if !opt.IsValid() {
		t.Error("Expected default value to be valid")
	}
}

func TestNumber_NegativeRange(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(-100.0).Maximum(-10.0)
	opt.Set(-50.5)
	if !opt.IsValid() {
		t.Error("Expected value -50.5 to be valid with minimum -100.0 and maximum -10.0")
	}

	opt.Set(-5.0)
	if opt.IsValid() {
		t.Error("Expected value -5.0 to be invalid (above maximum -10.0)")
	}

	opt.Set(-150.0)
	if opt.IsValid() {
		t.Error("Expected value -150.0 to be invalid (below minimum -100.0)")
	}
}

func TestNumber_WithEnum(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(0.0).Maximum(100.0).Enum(10.5, 20.5, 30.5, 40.5)
	opt.Set(20.5)
	if !opt.IsValid() {
		t.Error("Expected value 20.5 to be valid")
	}

	// Value in range but not in enum
	opt.Set(25.0)
	if opt.IsValid() {
		t.Error("Expected value 25.0 to be invalid (not in enum)")
	}
}

func TestNumber_SmallValues(t *testing.T) {
	opt := NewNumber[float64]("test").Minimum(0.001).Maximum(0.999)
	opt.Set(0.5)
	if !opt.IsValid() {
		t.Error("Expected value 0.5 to be valid")
	}

	opt.Set(0.0005)
	if opt.IsValid() {
		t.Error("Expected value 0.0005 to be invalid (below minimum)")
	}

	opt.Set(1.0)
	if opt.IsValid() {
		t.Error("Expected value 1.0 to be invalid (above maximum)")
	}
}
