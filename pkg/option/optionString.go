package option

import "regexp"

// String represents an option for string values.
type String struct {
	baseOption[string, *String]
}

func NewString(name string) *String {
	om := &String{}
	om.initCommon(name, om)
	return om
}

func (os *String) Regex(regexStr string) *String {
	os.Check(func(value string) bool {
		return regexp.MustCompile(regexStr).Match([]byte(os.Get()))
	}, map[string]any{"pattern": regexStr})
	return os
}
