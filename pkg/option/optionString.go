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

// MinLength sets the minimum allowed length for the string option.
// It adds both a validation check and the corresponding JSON schema property.
func (os *String) MinLength(min int) *String {
	os.Check(func(value string) string {
		if len(value) >= min {
			return ""
		}
		return fmt.Sprintf("value %q has length %d, which is less than minimum length %d", value, len(value), min)
	}, map[string]any{"minLength": min})
	return os
}

// MaxLength sets the maximum allowed length for the string option.
// It adds both a validation check and the corresponding JSON schema property.
func (os *String) MaxLength(max int) *String {
	os.Check(func(value string) string {
		if len(value) <= max {
			return ""
		}
		return fmt.Sprintf("value %q has length %d, which exceeds maximum length %d", value, len(value), max)
	}, map[string]any{"maxLength": max})
	return os
}
