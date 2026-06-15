package adb

import "strings"

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func trimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

func splitByWhitespace(s string) []string {
	fields := strings.Fields(s)
	return fields
}
