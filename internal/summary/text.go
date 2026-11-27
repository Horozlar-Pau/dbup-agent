package summary

import (
	"strings"
	"unicode/utf8"
)

// Limits from docs/summary.md.
const (
	// MaxInput is how many characters of the commit subject are sent to CLIProxy.
	MaxInput = 200
	// MaxSummary is the longest sentence that may leave the machine.
	MaxSummary = 140
)

const ellipsis = "…"

// Subject returns the first non-empty line of a commit message, at most MaxInput characters.
func Subject(message string) string {
	for _, line := range strings.Split(message, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return truncate(s, MaxInput)
		}
	}
	return ""
}

// clean turns the model's answer into one sentence of at most MaxSummary characters.
func clean(answer string) string {
	s := strings.Join(strings.Fields(answer), " ")
	s = strings.Trim(s, "\"'“”")
	if utf8.RuneCountInString(s) <= MaxSummary {
		return s
	}
	cut := truncate(s, MaxSummary-utf8.RuneCountInString(ellipsis))
	// Prefer ending on a word boundary when one is close.
	if i := strings.LastIndex(cut, " "); i > len(cut)*2/3 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,;:.") + ellipsis
}

// truncate keeps the first n runes of s.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
