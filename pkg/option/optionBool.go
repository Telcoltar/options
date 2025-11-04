package option

type Bool struct {
	baseOption[bool, *Bool]
}

func NewBool(name string) *Bool {
	om := &Bool{}
	om.initCommon(name, om)
	return om
}
