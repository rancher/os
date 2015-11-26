package util

import (
	"strings"
)

func GetValue(index int, args []string) string {
	val := args[index]
	parts := strings.SplitN(val, "=", 2)
	if len(parts) == 1 {
		if len(args) > index+1 {
			return args[index+1]
		} else {
			return ""
		}
	} else {
		return parts[1]
	}
}
