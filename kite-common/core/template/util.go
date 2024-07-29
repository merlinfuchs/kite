package template

import "strings"

func ContainsVariable(value string) bool {
	return strings.Contains(value, DelimLeft)
}
