package option

// String represents an option for string values.
type String struct {
	baseOption[string, *String]
}

func NewString(name string) *String {
	om := &String{}
	om.initCommon(name, om)
	return om
}

func (o *String) JSONSchemaType() string {
	return "string"
}

func (o *String) JSONSchemaProperty() map[string]any {
	property := map[string]any{
		"type": o.JSONSchemaType(),
	}
	if o.defaultValue != nil {
		property["default"] = *o.defaultValue
	}
	return property
}
