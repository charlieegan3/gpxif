package utils

import "strings"

func IsJPEGFile(filename string) bool {
	lowered := strings.ToLower(filename)

	if strings.HasSuffix(lowered, ".jpeg") {
		return true
	}

	if strings.HasSuffix(lowered, ".jpg") {
		return true
	}

	return false
}
