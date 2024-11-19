package common

import "strings"

func EqualsIgnoreCase(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	return a == b
}
