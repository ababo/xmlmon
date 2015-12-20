package data

import (
	"strings"
)

func encodeValue(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func encodeName(name string) string {
	return "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
}
