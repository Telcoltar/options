package option

import (
	"encoding/json"
	"testing"
)

func TestBaseJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		option   OptionInterface
		expected map[string]any
	}{
		{
			name:   "string option",
			option: NewSimple[string]("name"),
			expected: map[string]any{
				"type": "string",
			},
		},
		{
			name:   "int option with default",
			option: NewSimple[int]("count").Default(10),
			expected: map[string]any{
				"type":    "integer",
				"default": 10,
			},
		},
		{
			name:   "bool option",
			option: NewSimple[bool]("enabled").Default(true),
			expected: map[string]any{
				"type":    "boolean",
				"default": true,
			},
		},
		{
			name:   "float option",
			option: NewSimple[float64]("ratio").Default(3.14),
			expected: map[string]any{
				"type":    "number",
				"default": 3.14,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.option.JSONSchema()
			if !mapsEqual(result, tt.expected) {
				t.Errorf("JSONSchemaProperty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSliceJSONSchema(t *testing.T) {
	option := NewSlice[string]("tags")
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "string",
		},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestMapJSONSchema(t *testing.T) {
	option := NewMap[string]("labels")
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "object",
		"additionalProperties": map[string]any{
			"type": "string",
		},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestSliceEmptyDefaultJSONSchema(t *testing.T) {
	option := NewSlice[string]("tags").EmptyDefault()
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "string",
		},
		"default": []string{},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}

	defaultValue := result["default"]
	if defaultValue == nil {
		t.Error("default should not be nil, should be empty slice")
	}

	if sliceVal, ok := defaultValue.([]string); !ok {
		t.Errorf("default should be []string, got %T", defaultValue)
	} else if len(sliceVal) != 0 {
		t.Errorf("default slice should be empty, got length %d", len(sliceVal))
	}
}

func TestMapEmptyDefaultJSONSchema(t *testing.T) {
	option := NewMap[string]("labels").EmptyDefault()
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "object",
		"additionalProperties": map[string]any{
			"type": "string",
		},
		"default": map[string]string{},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchema() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}

	defaultValue := result["default"]
	if defaultValue == nil {
		t.Error("default should not be nil, should be empty map")
	}

	if mapVal, ok := defaultValue.(map[string]string); !ok {
		t.Errorf("default should be map[string]string, got %T", defaultValue)
	} else if len(mapVal) != 0 {
		t.Errorf("default map should be empty, got length %d", len(mapVal))
	}
}

func TestContainerJSONSchema(t *testing.T) {
	type TestConfig struct {
		Name    *Simple[string]
		Count   *Simple[int]
		Enabled *Simple[bool]

		*Container[TestConfig]
	}

	config := &TestConfig{
		Name:    NewSimple[string]("name").Default("test"),
		Count:   NewSimple[int]("count").Default(5),
		Enabled: NewSimple[bool]("enabled").Default(true),
	}
	config.Container = NewContainer("TestConfig", config)

	schema := config.JSONSchema()

	if schema["type"] != "object" {
		t.Errorf("schema type = %v, want object", schema["type"])
	}

	if schema["title"] != "TestConfig" {
		t.Errorf("schema title = %v, want TestConfig", schema["title"])
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties is not a map")
	}

	if len(properties) != 3 {
		t.Errorf("properties length = %d, want 3", len(properties))
	}

	nameProps, ok := properties["name"].(map[string]any)
	if !ok {
		t.Fatalf("name property is not a map")
	}
	if nameProps["type"] != "string" || nameProps["default"] != "test" {
		t.Errorf("name property = %v, want {type: string, default: test}", nameProps)
	}
}

func TestNestedContainerJSONSchema(t *testing.T) {
	type Inner struct {
		Value *Simple[string]
		*Container[Inner]
	}

	type Outer struct {
		Name  *Simple[string]
		Inner *Inner
		*Container[Outer]
	}

	inner := &Inner{
		Value: NewSimple[string]("value").Default("inner"),
	}
	inner.Container = NewContainer("inner", inner)

	outer := &Outer{
		Name:  NewSimple[string]("name").Default("outer"),
		Inner: inner,
	}
	outer.Container = NewContainer("outer", outer)

	schema := outer.JSONSchema()

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties is not a map")
	}

	innerSchema, ok := properties["inner"].(map[string]any)
	if !ok {
		t.Fatalf("inner property is not a map")
	}

	if innerSchema["type"] != "object" {
		t.Errorf("inner type = %v, want object", innerSchema["type"])
	}

	innerProps, ok := innerSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("inner properties is not a map")
	}

	valueProps, ok := innerProps["value"].(map[string]any)
	if !ok {
		t.Fatalf("value property is not a map")
	}

	if valueProps["type"] != "string" || valueProps["default"] != "inner" {
		t.Errorf("value property = %v, want {type: string, default: inner}", valueProps)
	}
}

func TestJSONSchemaWithMetadata(t *testing.T) {
	type TestConfig struct {
		Name *Simple[string]
		*Container[TestConfig]
	}

	config := &TestConfig{
		Name: NewSimple[string]("name").Default("test"),
	}
	config.Container = NewContainer("TestConfig", config)

	schema := config.JSONSchemaWithMetadata()

	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("$schema = %v, want https://json-schema.org/draft/2020-12/schema", schema["$schema"])
	}

	if schema["type"] != "object" {
		t.Errorf("type = %v, want object", schema["type"])
	}
}

func TestJSONSchemaMarshaling(t *testing.T) {
	type TestConfig struct {
		Name  *Simple[string]
		Count *Simple[int]
		*Container[TestConfig]
	}

	config := &TestConfig{
		Name:  NewSimple[string]("name").Default("test"),
		Count: NewSimple[int]("count").Default(5),
	}
	config.Container = NewContainer("TestConfig", config)

	schema := config.JSONSchemaWithMetadata()

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	if unmarshaled["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("unmarshaled $schema = %v", unmarshaled["$schema"])
	}
}

func TestBaseEnumJSONSchema(t *testing.T) {
	option := NewSimple[string]("color").Enum("red", "green", "blue")
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "string",
		"enum": []any{"red", "green", "blue"},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestBaseEnumWithDefaultJSONSchema(t *testing.T) {
	option := NewSimple[int]("size").Enum(1, 2, 3, 5, 8).Default(3)
	result := option.JSONSchema()

	expected := map[string]any{
		"type":    "integer",
		"enum":    []any{1, 2, 3, 5, 8},
		"default": 3,
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestSliceEnumJSONSchema(t *testing.T) {
	option := NewSlice[string]("ipFamilies").ItemOption.Enum("IPv4", "IPv6")
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "string",
			"enum": []any{"IPv4", "IPv6"},
		},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestSliceEnumWithDefaultJSONSchema(t *testing.T) {
	option := NewSlice[int]("numbers").ItemOption.Enum(1, 2, 3, 5, 8).Default([]int{1, 2})
	result := option.JSONSchema()

	expected := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "integer",
			"enum": []any{1, 2, 3, 5, 8},
		},
		"default": []int{1, 2},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestMapEnumJSONSchema(t *testing.T) {
	m := NewMap[string]("labels")
	m.ValueOption.Enum("dev", "prod")
	result := m.JSONSchema()

	expected := map[string]any{
		"type": "object",
		"additionalProperties": map[string]any{
			"type": "string",
			"enum": []any{"dev", "prod"},
		},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestMapEnumWithDefaultJSONSchema(t *testing.T) {
	m := NewMap[string]("labels")
	m.ValueOption.Enum("dev", "prod")
	m.Default(map[string]string{"env": "dev"})
	result := m.JSONSchema()

	expected := map[string]any{
		"type": "object",
		"additionalProperties": map[string]any{
			"type": "string",
			"enum": []any{"dev", "prod"},
		},
		"default": map[string]string{"env": "dev"},
	}

	if !mapsEqual(result, expected) {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("JSONSchemaProperty() =\n%s\nwant\n%s", resultJSON, expectedJSON)
	}
}

func TestContainerCheckWithIfElseSchema(t *testing.T) {
	type DeploymentConfig struct {
		Environment *Simple[string]
		MinReplicas *Simple[int]
		MaxReplicas *Simple[int]

		*Container[DeploymentConfig]
	}

	config := &DeploymentConfig{
		Environment: NewSimple[string]("environment").Enum("dev", "staging", "prod").Default("dev"),
		MinReplicas: NewSimple[int]("minReplicas").Default(1),
		MaxReplicas: NewSimple[int]("maxReplicas").Default(10),
	}
	config.Container = NewContainer("DeploymentConfig", config)

	// Add a validation check where BOTH the Go function AND the JSON Schema
	// express the SAME validation logic:
	// - In production: minReplicas >= 3 and maxReplicas >= 10
	// - In non-production: minReplicas >= 1
	// - Always: minReplicas < maxReplicas
	config.Container.Check(
		func(d *DeploymentConfig) string {
			// Validate minReplicas < maxReplicas
			if d.MinReplicas.HasValue() && d.MaxReplicas.HasValue() {
				if d.MinReplicas.Get() >= d.MaxReplicas.Get() {
					return "minReplicas must be less than maxReplicas"
				}
			}

			// Environment-specific validation
			if d.Environment.HasValue() {
				env := d.Environment.Get()
				if env == "prod" {
					// Production requires higher minimums
					if d.MinReplicas.HasValue() && d.MinReplicas.Get() < 3 {
						return "production requires minReplicas >= 3"
					}
					if d.MaxReplicas.HasValue() && d.MaxReplicas.Get() < 10 {
						return "production requires maxReplicas >= 10"
					}
				} else {
					// Non-production requires at least 1 replica
					if d.MinReplicas.HasValue() && d.MinReplicas.Get() < 1 {
						return "non-production requires minReplicas >= 1"
					}
				}
			}

			return ""
		},
		// JSON Schema that reflects the SAME validation logic
		map[string]any{
			"if": map[string]any{
				"properties": map[string]any{
					"environment": map[string]any{
						"const": "prod",
					},
				},
			},
			"then": map[string]any{
				"properties": map[string]any{
					"minReplicas": map[string]any{
						"minimum": 3,
					},
					"maxReplicas": map[string]any{
						"minimum": 10,
					},
				},
			},
			"else": map[string]any{
				"properties": map[string]any{
					"minReplicas": map[string]any{
						"minimum": 1,
					},
				},
			},
		},
	)

	// Test 1: Valid production config
	t.Run("valid production config", func(t *testing.T) {
		config.Environment.Set("prod")
		config.MinReplicas.Set(3)
		config.MaxReplicas.Set(10)
		if !config.IsValid() {
			t.Error("Expected valid production config (minReplicas=3, maxReplicas=10)")
		}
	})

	// Test 2: Invalid production config (too few minReplicas)
	t.Run("invalid production - low minReplicas", func(t *testing.T) {
		config.Environment.Set("prod")
		config.MinReplicas.Set(2) // Less than required 3
		config.MaxReplicas.Set(10)
		if config.IsValid() {
			t.Error("Expected invalid: prod requires minReplicas >= 3")
		}
	})

	// Test 3: Invalid production config (too few maxReplicas)
	t.Run("invalid production - low maxReplicas", func(t *testing.T) {
		config.Environment.Set("prod")
		config.MinReplicas.Set(3)
		config.MaxReplicas.Set(9) // Less than required 10
		if config.IsValid() {
			t.Error("Expected invalid: prod requires maxReplicas >= 10")
		}
	})

	// Test 4: Valid dev config
	t.Run("valid dev config", func(t *testing.T) {
		config.Environment.Set("dev")
		config.MinReplicas.Set(1)
		config.MaxReplicas.Set(5)
		if !config.IsValid() {
			t.Error("Expected valid dev config (minReplicas=1, maxReplicas=5)")
		}
	})

	// Test 5: Invalid dev config (minReplicas < 1)
	t.Run("invalid dev - zero replicas", func(t *testing.T) {
		config.Environment.Set("dev")
		config.MinReplicas.Set(0)
		config.MaxReplicas.Set(5)
		if config.IsValid() {
			t.Error("Expected invalid: dev requires minReplicas >= 1")
		}
	})

	// Test 6: Invalid regardless of environment (minReplicas >= maxReplicas)
	t.Run("invalid - minReplicas >= maxReplicas", func(t *testing.T) {
		config.Environment.Set("dev")
		config.MinReplicas.Set(10)
		config.MaxReplicas.Set(5)
		if config.IsValid() {
			t.Error("Expected invalid: minReplicas must be < maxReplicas")
		}
	})

	// Test 7: Verify JSON Schema includes the if/then/else
	t.Run("schema generation", func(t *testing.T) {
		schema := config.JSONSchema()

		if schema["type"] != "object" {
			t.Errorf("schema type = %v, want object", schema["type"])
		}

		// Verify if/then/else properties exist in schema
		if _, ok := schema["if"]; !ok {
			t.Error("Expected schema to contain 'if' property")
		}

		if _, ok := schema["then"]; !ok {
			t.Error("Expected schema to contain 'then' property")
		}

		if _, ok := schema["else"]; !ok {
			t.Error("Expected schema to contain 'else' property")
		}

		// Verify the if condition
		ifClause, ok := schema["if"].(map[string]any)
		if !ok {
			t.Fatal("if clause is not a map")
		}

		ifProps, ok := ifClause["properties"].(map[string]any)
		if !ok {
			t.Fatal("if.properties is not a map")
		}

		envIf, ok := ifProps["environment"].(map[string]any)
		if !ok {
			t.Fatal("if.properties.environment is not a map")
		}

		if envIf["const"] != "prod" {
			t.Errorf("if.properties.environment.const = %v, want 'prod'", envIf["const"])
		}

		// Verify the then clause (production requirements)
		thenClause, ok := schema["then"].(map[string]any)
		if !ok {
			t.Fatal("then clause is not a map")
		}

		thenProps, ok := thenClause["properties"].(map[string]any)
		if !ok {
			t.Fatal("then.properties is not a map")
		}

		minReplicasThen, ok := thenProps["minReplicas"].(map[string]any)
		if !ok {
			t.Fatal("then.properties.minReplicas is not a map")
		}

		if minReplicasThen["minimum"] != 3 {
			t.Errorf("then.properties.minReplicas.minimum = %v, want 3", minReplicasThen["minimum"])
		}

		maxReplicasThen, ok := thenProps["maxReplicas"].(map[string]any)
		if !ok {
			t.Fatal("then.properties.maxReplicas is not a map")
		}

		if maxReplicasThen["minimum"] != 10 {
			t.Errorf("then.properties.maxReplicas.minimum = %v, want 10", maxReplicasThen["minimum"])
		}

		// Verify the else clause (non-production requirements)
		elseClause, ok := schema["else"].(map[string]any)
		if !ok {
			t.Fatal("else clause is not a map")
		}

		elseProps, ok := elseClause["properties"].(map[string]any)
		if !ok {
			t.Fatal("else.properties is not a map")
		}

		minReplicasElse, ok := elseProps["minReplicas"].(map[string]any)
		if !ok {
			t.Fatal("else.properties.minReplicas is not a map")
		}

		if minReplicasElse["minimum"] != 1 {
			t.Errorf("else.properties.minReplicas.minimum = %v, want 1", minReplicasElse["minimum"])
		}

		// Verify that standard properties still exist
		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties is not a map")
		}

		if len(properties) != 3 {
			t.Errorf("properties length = %d, want 3", len(properties))
		}
	})
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}

		switch av := v.(type) {
		case map[string]any:
			if bvm, ok := bv.(map[string]any); ok {
				if !mapsEqual(av, bvm) {
					return false
				}
			} else {
				return false
			}
		default:
			aJSON, _ := json.Marshal(av)
			bJSON, _ := json.Marshal(bv)
			if string(aJSON) != string(bJSON) {
				return false
			}
		}
	}

	return true
}
