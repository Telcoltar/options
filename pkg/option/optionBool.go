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
