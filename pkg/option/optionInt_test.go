package option

import (
	"testing"
)

func TestInt_Basic(t *testing.T) {
	opt := NewInt[int]("test")
	opt.Set(42)
	if opt.Get() != 42 {
		t.Errorf("Expected 42, got %d", opt.Get())
	}
}

func TestInt_JSONSchemaType(t *testing.T) {
	opt := NewInt[int]("test")
	schema := opt.JSONSchema()
	if schema["type"] != "integer" {
		t.Errorf("Expected type 'integer', got %v", schema["type"])
	}
}

func TestInt_Minimum_Valid(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10)
	opt.Set(15)
	if !opt.IsValid() {
		t.Error("Expected value 15 to be valid with minimum 10")
	}
}

func TestInt_Minimum_Invalid(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10)
	opt.Set(5)
	if opt.IsValid() {
		t.Error("Expected value 5 to be invalid with minimum 10")
	}
}

func TestInt_Minimum_Boundary(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10)
	opt.Set(10)
	if !opt.IsValid() {
		t.Error("Expected value 10 to be valid with minimum 10")
	}
}

func TestInt_Maximum_Valid(t *testing.T) {
	opt := NewInt[int]("test").Maximum(100)
	opt.Set(50)
	if !opt.IsValid() {
		t.Error("Expected value 50 to be valid with maximum 100")
	}
}

func TestInt_Maximum_Invalid(t *testing.T) {
	opt := NewInt[int]("test").Maximum(100)
	opt.Set(150)
	if opt.IsValid() {
		t.Error("Expected value 150 to be invalid with maximum 100")
	}
}

func TestInt_Maximum_Boundary(t *testing.T) {
	opt := NewInt[int]("test").Maximum(100)
	opt.Set(100)
	if !opt.IsValid() {
		t.Error("Expected value 100 to be valid with maximum 100")
	}
}

func TestInt_MinimumMaximum_Valid(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10).Maximum(100)
	opt.Set(50)
	if !opt.IsValid() {
		t.Error("Expected value 50 to be valid with minimum 10 and maximum 100")
	}
}

func TestInt_MinimumMaximum_TooLow(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10).Maximum(100)
	opt.Set(5)
	if opt.IsValid() {
		t.Error("Expected value 5 to be invalid with minimum 10")
	}
}

func TestInt_MinimumMaximum_TooHigh(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10).Maximum(100)
	opt.Set(150)
	if opt.IsValid() {
		t.Error("Expected value 150 to be invalid with maximum 100")
	}
}

func TestInt_JSONSchema_Minimum(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10)
	schema := opt.JSONSchema()
	if schema["minimum"] != 10 {
		t.Errorf("Expected minimum property to be 10, got %v", schema["minimum"])
	}
}

func TestInt_JSONSchema_Maximum(t *testing.T) {
	opt := NewInt[int]("test").Maximum(100)
	schema := opt.JSONSchema()
	if schema["maximum"] != 100 {
		t.Errorf("Expected maximum property to be 100, got %v", schema["maximum"])
	}
}

func TestInt_JSONSchema_MinimumMaximum(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10).Maximum(100)
	schema := opt.JSONSchema()
	if schema["minimum"] != 10 {
		t.Errorf("Expected minimum property to be 10, got %v", schema["minimum"])
	}
	if schema["maximum"] != 100 {
		t.Errorf("Expected maximum property to be 100, got %v", schema["maximum"])
	}
}

func TestInt_DifferentTypes(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"int8", func(t *testing.T) {
			opt := NewInt[int8]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected int8 value to be valid")
			}
		}},
		{"int16", func(t *testing.T) {
			opt := NewInt[int16]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected int16 value to be valid")
			}
		}},
		{"int32", func(t *testing.T) {
			opt := NewInt[int32]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected int32 value to be valid")
			}
		}},
		{"int64", func(t *testing.T) {
			opt := NewInt[int64]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected int64 value to be valid")
			}
		}},
		{"uint", func(t *testing.T) {
			opt := NewInt[uint]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected uint value to be valid")
			}
		}},
		{"uint8", func(t *testing.T) {
			opt := NewInt[uint8]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected uint8 value to be valid")
			}
		}},
		{"uint16", func(t *testing.T) {
			opt := NewInt[uint16]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected uint16 value to be valid")
			}
		}},
		{"uint32", func(t *testing.T) {
			opt := NewInt[uint32]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected uint32 value to be valid")
			}
		}},
		{"uint64", func(t *testing.T) {
			opt := NewInt[uint64]("test").Minimum(10).Maximum(100)
			opt.Set(50)
			if !opt.IsValid() {
				t.Error("Expected uint64 value to be valid")
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestInt_WithDefault(t *testing.T) {
	opt := NewInt[int]("test").Minimum(10).Maximum(100).Default(50)
	if opt.Get() != 50 {
		t.Errorf("Expected default value 50, got %d", opt.Get())
	}
	if !opt.IsValid() {
		t.Error("Expected default value to be valid")
	}
}

func TestInt_WithEnum(t *testing.T) {
	opt := NewInt[int]("test").Minimum(0).Maximum(100).Enum(10, 20, 30, 40)
	opt.Set(20)
	if !opt.IsValid() {
		t.Error("Expected value 20 to be valid")
	}

	// Value in range but not in enum
	opt.Set(25)
	if opt.IsValid() {
		t.Error("Expected value 25 to be invalid (not in enum)")
	}

	// Value in enum but out of range should fail minimum check
	opt.Set(5)
	if opt.IsValid() {
		t.Error("Expected value 5 to be invalid (below minimum)")
	}
}
