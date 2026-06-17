package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/openarso/arso/apps/cli/internal/clioutput"
	"github.com/openarso/arso/apps/cli/internal/satellite"
)

func TestPrintPassPredictionsTextFormatsMetadataAndPasses(t *testing.T) {
	cmd, stdout, _ := newTestCommandIO()
	result := samplePassPredictionResult()

	if err := printPassPredictions(cmd, result, clioutput.Text); err != nil {
		t.Fatalf("printPassPredictions() unexpected error: %v", err)
	}

	wantLines := []string{
		formatPassLabelLine("Name:", result.Name),
		formatPassLabelLine("Kind:", result.Kind),
		formatPassLabelLine("NORAD ID:", fmt.Sprintf("%d", result.NoradID)),
		formatPassLabelLine("Object ID:", result.ObjectID),
		formatPassLabelLine("Observer:", result.ObserverName),
		formatPassLabelLine("Acquisition of signal (AOS):", "2026-06-10 22:15:00 UTC"),
		formatPassLabelLine("Loss of signal (LOS):", "2026-06-10 22:27:30 UTC"),
		formatPassLabelLine("Duration:", "12m30s"),
		formatPassLabelLine("Maximum elevation:", "47.2°"),
		formatPassLabelLine("Maximum elevation time:", "2026-06-10 22:21:15 UTC"),
		formatPassLabelLine("Azimuth at AOS:", "130.4°"),
		formatPassLabelLine("Azimuth at LOS:", "44.8°"),
		"",
		formatPassLabelLine("Acquisition of signal (AOS):", "2026-06-11 00:03:00 UTC"),
		formatPassLabelLine("Loss of signal (LOS):", "2026-06-11 00:14:00 UTC"),
		formatPassLabelLine("Duration:", "11m0s"),
		formatPassLabelLine("Maximum elevation:", "31.5°"),
		formatPassLabelLine("Maximum elevation time:", "2026-06-11 00:08:45 UTC"),
		formatPassLabelLine("Azimuth at AOS:", "214.2°"),
		formatPassLabelLine("Azimuth at LOS:", "98.6°"),
		"",
	}

	if got, want := stdout.String(), strings.Join(wantLines, "\n"); got != want {
		t.Fatalf("text output mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrintPassPredictionsJSONAndNDJSONStayConsistent(t *testing.T) {
	result := samplePassPredictionResult()

	jsonCmd, jsonStdout, _ := newTestCommandIO()
	if err := printPassPredictions(jsonCmd, result, clioutput.JSON); err != nil {
		t.Fatalf("printPassPredictions(JSON) unexpected error: %v", err)
	}

	var jsonResult satellite.PassPredictionResult
	if err := json.Unmarshal(jsonStdout.Bytes(), &jsonResult); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got, want := jsonResult, result; !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON result mismatch: got %#v want %#v", got, want)
	}

	ndjsonCmd, ndjsonStdout, _ := newTestCommandIO()
	if err := printPassPredictions(ndjsonCmd, result, clioutput.NDJSON); err != nil {
		t.Fatalf("printPassPredictions(NDJSON) unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(ndjsonStdout.String()), "\n")
	if got, want := len(lines), len(result.Passes); got != want {
		t.Fatalf("NDJSON line count = %d, want %d", got, want)
	}

	for i, line := range lines {
		var pass satellite.PredictedPass
		if err := json.Unmarshal([]byte(line), &pass); err != nil {
			t.Fatalf("json.Unmarshal(line %d) error = %v", i, err)
		}

		if got, want := pass, result.Passes[i]; got != want {
			t.Fatalf("NDJSON pass %d mismatch: got %#v want %#v", i, got, want)
		}
	}
}

func TestPrintPassPredictionsHandlesEmptyResults(t *testing.T) {
	result := samplePassPredictionResult()
	result.Passes = nil

	textCmd, textStdout, _ := newTestCommandIO()
	if err := printPassPredictions(textCmd, result, clioutput.Text); err != nil {
		t.Fatalf("printPassPredictions(TEXT) unexpected error: %v", err)
	}

	if strings.Contains(textStdout.String(), "Acquisition of signal") {
		t.Fatalf("text output unexpectedly included a pass: %q", textStdout.String())
	}

	jsonCmd, jsonStdout, _ := newTestCommandIO()
	if err := printPassPredictions(jsonCmd, result, clioutput.JSON); err != nil {
		t.Fatalf("printPassPredictions(JSON) unexpected error: %v", err)
	}

	var decoded satellite.PassPredictionResult
	if err := json.Unmarshal(jsonStdout.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.Passes != nil && len(decoded.Passes) != 0 {
		t.Fatalf("decoded passes = %#v, want empty", decoded.Passes)
	}

	ndjsonCmd, ndjsonStdout, _ := newTestCommandIO()
	if err := printPassPredictions(ndjsonCmd, result, clioutput.NDJSON); err != nil {
		t.Fatalf("printPassPredictions(NDJSON) unexpected error: %v", err)
	}

	if got := ndjsonStdout.String(); got != "" {
		t.Fatalf("NDJSON output = %q, want empty string", got)
	}
}

func TestPrintPassPredictionsRejectsUnsupportedOutput(t *testing.T) {
	cmd, _, _ := newTestCommandIO()

	err := printPassPredictions(cmd, samplePassPredictionResult(), "yaml")
	if err == nil {
		t.Fatal("printPassPredictions() expected error, got nil")
	}

	if !strings.Contains(err.Error(), `unhandled output format "yaml"`) {
		t.Fatalf("error = %q, want unsupported format message", err)
	}
}

func formatPassLabelLine(label string, value string) string {
	return fmt.Sprintf("%-30s %s", label, value)
}
