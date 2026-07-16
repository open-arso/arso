package output

import (
	"fmt"
	"strings"
)

const (
	// Text selects human-readable CLI output.
	Text = "text"
	// JSON selects a single indented JSON document.
	JSON = "json"
	// NDJSON selects newline-delimited JSON records.
	NDJSON = "ndjson"
)

// Validate reports whether format is one of the allowed output formats.
// When allowed is empty, Validate accepts the default text and json formats.
func Validate(format string, allowed ...string) error {
	format = strings.TrimSpace(format)

	if len(allowed) == 0 {
		allowed = []string{Text, JSON}
	}

	for _, candidate := range allowed {
		if format == candidate {
			return nil
		}
	}

	return fmt.Errorf(
		"unsupported output format %q, expected one of: %s",
		format,
		strings.Join(allowed, ", "),
	)
}

// Normalize trims surrounding whitespace and lowercases a CLI output format.
func Normalize(format string) string {
	return strings.ToLower(strings.TrimSpace(format))
}
