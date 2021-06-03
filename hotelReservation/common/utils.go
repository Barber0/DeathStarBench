package common

import "strings"

func HeadToUpper(str string) string {
	if len(str) < 2 {
		return strings.ToUpper(str)
	}
	return strings.ToUpper(str[:1]) + str[1:]
}
