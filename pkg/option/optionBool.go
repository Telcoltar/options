package option

// Bool represents an option for boolean values.
type Bool struct {
	baseOption[bool, *Bool]
}

func NewBool(name string) *Bool {
	om := &Bool{}
	om.initCommon(name, om)
	return om
}

func (o *Bool) JSONSchemaType() string {
	return "boolean"
}

func (o *Bool) JSONSchemaProperty() map[string]any {
	property := map[string]any{
		"type": o.JSONSchemaType(),
	}
	if o.defaultValue != nil {
		property["default"] = *o.defaultValue
	}
	return property
}
