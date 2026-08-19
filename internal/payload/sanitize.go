package payload

import (
	"regexp"
	"strings"
	"unicode"
)

// ansiRegex matches ANSI escape codes, CSI sequences, OSC sequences, and terminal control sequences.
var ansiRegex = regexp.MustCompile(`(?:\x1b\[[0-9:;<=>?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\|\x1b)|[\x1b\x9b][\[\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\d\/#&.:=?%@~_]+)*|[a-zA-Z\d]+(?:;[-a-zA-Z\d\/#&.:=?%@~_]*)*)?\x07)|(?:(?:\d{1,4}(?:;\d{0,4})*)?[\dA-PR-TZcf-ntqry=><~]))|\x1b[@-Z\\-_])`)

// Sanitize removes ANSI escape sequences and unicode control characters from input string,
// returning a clean, single-line safe string.
func Sanitize(s string) string {
	if s == "" {
		return ""
	}

	// 1. Strip ANSI escape sequences
	cleaned := ansiRegex.ReplaceAllString(s, "")

	// 2. Strip control characters (including \r, \n, \t, null bytes, C0, C1 controls)
	var sb strings.Builder
	sb.Grow(len(cleaned))
	for _, r := range cleaned {
		if !unicode.IsControl(r) {
			sb.WriteRune(r)
		}
	}

	return strings.TrimSpace(sb.String())
}

// StripANSI removes only ANSI escape sequences from the string without stripping other control characters.
func StripANSI(s string) string {
	if s == "" {
		return ""
	}
	return ansiRegex.ReplaceAllString(s, "")
}

// StripControlChars removes unicode control characters from the string.
func StripControlChars(s string) string {
	if s == "" {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if !unicode.IsControl(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
