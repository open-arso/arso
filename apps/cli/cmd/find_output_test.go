package cmd

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/openarso/arso/apps/cli/internal/clioutput"
	"github.com/openarso/arso/apps/cli/internal/satellite"
)

func TestPrintApparentPositionsSupportsTextJSONAndNDJSON(t *testing.T) {
	positions := sampleApparentPositions()

	textCmd, textStdout, _ := newTestCommandIO()
	if err := printApparentPositions(textCmd, positions, clioutput.Text); err != nil {
		t.Fatalf("printApparentPositions(TEXT) unexpected error: %v", err)
	}

	text := textStdout.String()
	for _, fragment := range []string{
		"Name:          ISS (ZARYA)",
		"Observer:      paris-node",
		"Time UTC:      2026-06-10T22:00:00Z",
		"Above horizon: true",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("text output missing fragment %q in %q", fragment, text)
		}
	}

	jsonCmd, jsonStdout, _ := newTestCommandIO()
	if err := printApparentPositions(jsonCmd, positions, clioutput.JSON); err != nil {
		t.Fatalf("printApparentPositions(JSON) unexpected error: %v", err)
	}

	var decodedJSON []satellite.ApparentPosition
	if err := json.Unmarshal(jsonStdout.Bytes(), &decodedJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got, want := decodedJSON, positions; !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded JSON = %#v, want %#v", got, want)
	}

	ndjsonCmd, ndjsonStdout, _ := newTestCommandIO()
	if err := printApparentPositions(ndjsonCmd, positions, clioutput.NDJSON); err != nil {
		t.Fatalf("printApparentPositions(NDJSON) unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(ndjsonStdout.String()), "\n")
	if got, want := len(lines), len(positions); got != want {
		t.Fatalf("NDJSON line count = %d, want %d", got, want)
	}
}

func TestPrintElementsSupportsTextJSONAndNDJSON(t *testing.T) {
	elements := sampleElements()

	textCmd, textStdout, _ := newTestCommandIO()
	if err := printElements(textCmd, elements, clioutput.Text); err != nil {
		t.Fatalf("printElements(TEXT) unexpected error: %v", err)
	}

	text := textStdout.String()
	for _, fragment := range []string{
		"Name:              ISS (ZARYA)",
		"Source:            celestrak",
		"Epoch UTC:         2026-06-10T22:00:00Z",
		"Mean motion:       15.48912345 rev/day",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("text output missing fragment %q in %q", fragment, text)
		}
	}

	jsonCmd, jsonStdout, _ := newTestCommandIO()
	if err := printElements(jsonCmd, elements, clioutput.JSON); err != nil {
		t.Fatalf("printElements(JSON) unexpected error: %v", err)
	}

	var decodedJSON []map[string]any
	if err := json.Unmarshal(jsonStdout.Bytes(), &decodedJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got, want := len(decodedJSON), 1; got != want {
		t.Fatalf("decoded JSON length = %d, want %d", got, want)
	}
	if got, want := decodedJSON[0]["epoch_utc"], "2026-06-10T22:00:00Z"; got != want {
		t.Fatalf("epoch_utc = %#v, want %#v", got, want)
	}

	ndjsonCmd, ndjsonStdout, _ := newTestCommandIO()
	if err := printElements(ndjsonCmd, elements, clioutput.NDJSON); err != nil {
		t.Fatalf("printElements(NDJSON) unexpected error: %v", err)
	}
	if got := strings.Count(strings.TrimSpace(ndjsonStdout.String()), "\n") + 1; got != 1 {
		t.Fatalf("NDJSON line count = %d, want 1", got)
	}
}

func TestSelectResolvedTargetRetriesAndReturnsChoice(t *testing.T) {
	candidates := []satellite.ResolvedTarget{
		{Name: "ISS (ZARYA)", NoradID: 25544, ObjectID: "1998-067A"},
		{Name: "HUBBLE", NoradID: 20580, ObjectID: "1990-037B"},
	}

	selected, err := selectResolvedTarget(
		strings.NewReader("bad\n0\n2\n"),
		io.Discard,
		candidates,
	)
	if err != nil {
		t.Fatalf("selectResolvedTarget() unexpected error: %v", err)
	}

	if got, want := selected, candidates[1]; got != want {
		t.Fatalf("selected target = %#v, want %#v", got, want)
	}
}

func TestSelectResolvedTargetReturnsEOFWhenInputEnds(t *testing.T) {
	_, err := selectResolvedTarget(strings.NewReader(""), io.Discard, []satellite.ResolvedTarget{{Name: "ISS"}})
	if err == nil {
		t.Fatal("selectResolvedTarget() expected error, got nil")
	}
	if err != io.EOF {
		t.Fatalf("selectResolvedTarget() error = %v, want %v", err, io.EOF)
	}
}

func TestSelectResolvedTargetRejectsEmptyCandidates(t *testing.T) {
	_, err := selectResolvedTarget(strings.NewReader("1\n"), io.Discard, nil)
	if err == nil {
		t.Fatal("selectResolvedTarget() expected error, got nil")
	}
}

func TestParseFindAtAndNormalizeEpochValidation(t *testing.T) {
	if _, err := parseFindAt("not-a-time"); err == nil {
		t.Fatal("parseFindAt() expected error, got nil")
	}

	tests := []struct {
		name  string
		epoch string
		want  string
	}{
		{name: "empty epoch", epoch: "", want: ""},
		{name: "already UTC", epoch: "2026-06-10T22:00:00Z", want: "2026-06-10T22:00:00Z"},
		{name: "appends UTC", epoch: "2026-06-10T22:00:00", want: "2026-06-10T22:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCelesTrakEpoch(tt.epoch); got != tt.want {
				t.Fatalf("normalizeCelesTrakEpoch(%q) = %q, want %q", tt.epoch, got, tt.want)
			}
		})
	}
}
