package placeholder

import "strings"

func ContainsPlaceholder(input string) bool {
	return strings.Contains(input, startTag)
}
