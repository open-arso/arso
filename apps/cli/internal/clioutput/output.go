package clioutput

import (
	"fmt"
	"strings"
)

const (
	Text   = "text"
	JSON   = "json"
	NDJSON = "ndjson"
)

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

func Normalize(format string) string {
	return strings.ToLower(strings.TrimSpace(format))
}
