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
