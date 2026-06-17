package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCmdOutputsTextAndJSON(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalOutput := versionOutput

	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		versionOutput = originalOutput
	})

	Version = "1.2.3"
	Commit = "abc123"
	Date = "2026-06-10T22:00:00Z"

	textCmd, textStdout, _ := newTestCommandIO()
	versionOutput = "text"
	if err := versionCmd.RunE(textCmd, nil); err != nil {
		t.Fatalf("versionCmd.RunE(TEXT) unexpected error: %v", err)
	}
	for _, fragment := range []string{
		"ARSO 1.2.3",
		"Commit: abc123",
		"Built:  2026-06-10T22:00:00Z",
	} {
		if !strings.Contains(textStdout.String(), fragment) {
			t.Fatalf("text output missing fragment %q in %q", fragment, textStdout.String())
		}
	}

	jsonCmd, jsonStdout, _ := newTestCommandIO()
	versionOutput = " JSON "
	if err := versionCmd.RunE(jsonCmd, nil); err != nil {
		t.Fatalf("versionCmd.RunE(JSON) unexpected error: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(jsonStdout.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded["version"] != Version || decoded["commit"] != Commit || decoded["date"] != Date {
		t.Fatalf("decoded JSON = %#v, want version=%q commit=%q date=%q", decoded, Version, Commit, Date)
	}
}

func TestVersionCmdRejectsUnsupportedOutput(t *testing.T) {
	originalOutput := versionOutput
	t.Cleanup(func() {
		versionOutput = originalOutput
	})

	versionOutput = "ndjson"

	if err := versionCmd.RunE(versionCmd, nil); err == nil {
		t.Fatal("versionCmd.RunE() expected error, got nil")
	}
}
