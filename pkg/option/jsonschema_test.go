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

		*Container
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
		*Container
	}

	type Outer struct {
		Name  *Simple[string]
		Inner *Inner
		*Container
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
		*Container
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
		*Container
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
