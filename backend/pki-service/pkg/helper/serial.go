package helper

import (
	"strings"
)

// NormalizeSerial removes all colons and transforms the serial number to lowercase.
func NormalizeSerial(serial string) string {
	return strings.ToLower(strings.ReplaceAll(serial, ":", ""))
}
