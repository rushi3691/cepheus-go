package utils

import (
	"strings"
)

func GetInitials(name string) string {
	var sb strings.Builder
	f := 0
	for i := 0; i < len(name); i++ {
		if f == 0 && name[i] != ' ' {
			f = 1
			sb.WriteByte(name[i])
		} else if name[i] == ' ' {
			f = 0
		}
	}

	return sb.String()
}
