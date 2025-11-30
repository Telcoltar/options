package option

import (
	"fmt"
	"regexp"
)

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
	os.Check(func(value string) string {
		if regexp.MustCompile(regexStr).Match([]byte(os.Get())) {
			return ""
		}
		return fmt.Sprintf("value %q does not match pattern %q", value, regexStr)
	}, map[string]any{"pattern": regexStr})
	return os
}
