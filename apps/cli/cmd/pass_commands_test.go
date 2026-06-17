package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/openarso/arso/apps/cli/internal/satellite"
)

type stubSatelliteClient struct {
	elementsFunc            func(context.Context, string) ([]satellite.GPElement, error)
	locateFunc              func(context.Context, string, satellite.Observer, time.Time) ([]satellite.ApparentPosition, error)
	passPredictionsFunc     func(context.Context, string, satellite.Observer, string, string, int) (satellite.PassPredictionResult, error)
	nextPassFunc            func(context.Context, string, satellite.Observer, string, string, int) (satellite.PassPredictionResult, error)
	cacheResolvedTargetFunc func(string, satellite.ResolvedTarget) error
}

func (s stubSatelliteClient) Elements(ctx context.Context, target string) ([]satellite.GPElement, error) {
	if s.elementsFunc == nil {
		return nil, errors.New("unexpected Elements call")
	}
	return s.elementsFunc(ctx, target)
}

func (s stubSatelliteClient) Locate(ctx context.Context, target string, observer satellite.Observer, at time.Time) ([]satellite.ApparentPosition, error) {
	if s.locateFunc == nil {
		return nil, errors.New("unexpected Locate call")
	}
	return s.locateFunc(ctx, target, observer, at)
}

func (s stubSatelliteClient) PassPredictions(ctx context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error) {
	if s.passPredictionsFunc == nil {
		return satellite.PassPredictionResult{}, errors.New("unexpected PassPredictions call")
	}
	return s.passPredictionsFunc(ctx, target, observer, at, lookahead, minElevation)
}

func (s stubSatelliteClient) NextPass(ctx context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error) {
	if s.nextPassFunc == nil {
		return satellite.PassPredictionResult{}, errors.New("unexpected NextPass call")
	}
	return s.nextPassFunc(ctx, target, observer, at, lookahead, minElevation)
}

func (s stubSatelliteClient) CacheResolvedTarget(query string, resolved satellite.ResolvedTarget) error {
	if s.cacheResolvedTargetFunc == nil {
		return errors.New("unexpected CacheResolvedTarget call")
	}
	return s.cacheResolvedTargetFunc(query, resolved)
}

func TestListCmdRunEUsesNormalizedNDJSONOutput(t *testing.T) {
	cfg := configuredTestConfig()
	result := sampleSinglePassPredictionResult()
	var called bool

	restoreCommandDeps(t, stubSatelliteClient{
		passPredictionsFunc: func(_ context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error) {
			called = true

			if target != "ISS" {
				t.Fatalf("target = %q, want %q", target, "ISS")
			}
			if observer.Name != cfg.Node.Name {
				t.Fatalf("observer.Name = %q, want %q", observer.Name, cfg.Node.Name)
			}
			if observer.LatitudeDeg != *cfg.Observatory.Latitude {
				t.Fatalf("observer.LatitudeDeg = %v, want %v", observer.LatitudeDeg, *cfg.Observatory.Latitude)
			}
			if at != "2026-06-10T22:00:00Z" {
				t.Fatalf("at = %q, want %q", at, "2026-06-10T22:00:00Z")
			}
			if lookahead != "72h" {
				t.Fatalf("lookahead = %q, want %q", lookahead, "72h")
			}
			if minElevation != 15 {
				t.Fatalf("minElevation = %d, want %d", minElevation, 15)
			}

			return result, nil
		},
	}, cfg, nil)

	restoreListState(t)
	fromTimeList = "2026-06-10T22:00:00Z"
	lookaheadList = "72h"
	minElevationList = 15
	outputList = " NDJSON "

	cmd, stdout, _ := newTestCommandIO()
	if err := listCmd.RunE(cmd, []string{"ISS"}); err != nil {
		t.Fatalf("listCmd.RunE() unexpected error: %v", err)
	}

	if !called {
		t.Fatal("PassPredictions was not called")
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if got, want := len(lines), 1; got != want {
		t.Fatalf("NDJSON line count = %d, want %d", got, want)
	}

	var pass satellite.PredictedPass
	if err := json.Unmarshal([]byte(lines[0]), &pass); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got, want := pass, result.Passes[0]; got != want {
		t.Fatalf("decoded pass = %#v, want %#v", got, want)
	}
}

func TestNextCmdRunEUsesNormalizedJSONOutput(t *testing.T) {
	cfg := configuredTestConfig()
	result := sampleSinglePassPredictionResult()
	var called bool

	restoreCommandDeps(t, stubSatelliteClient{
		nextPassFunc: func(_ context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error) {
			called = true

			if target != "ISS" {
				t.Fatalf("target = %q, want %q", target, "ISS")
			}
			if observer.LongitudeDeg != *cfg.Observatory.Longitude {
				t.Fatalf("observer.LongitudeDeg = %v, want %v", observer.LongitudeDeg, *cfg.Observatory.Longitude)
			}
			if at != "2026-06-10T22:00:00Z" {
				t.Fatalf("at = %q, want %q", at, "2026-06-10T22:00:00Z")
			}
			if lookahead != "24h" {
				t.Fatalf("lookahead = %q, want %q", lookahead, "24h")
			}
			if minElevation != 25 {
				t.Fatalf("minElevation = %d, want %d", minElevation, 25)
			}

			return result, nil
		},
	}, cfg, nil)

	restoreNextState(t)
	fromTime = "2026-06-10T22:00:00Z"
	lookahead = "24h"
	minElevation = 25
	output = " JSON "

	cmd, stdout, _ := newTestCommandIO()
	if err := nextCmd.RunE(cmd, []string{"ISS"}); err != nil {
		t.Fatalf("nextCmd.RunE() unexpected error: %v", err)
	}

	if !called {
		t.Fatal("NextPass was not called")
	}

	var decoded satellite.PassPredictionResult
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got, want := decoded, result; !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded result = %#v, want %#v", got, want)
	}
}

func TestPassCommandsRejectUnsupportedOutputBeforeCallingDependencies(t *testing.T) {
	cfg := configuredTestConfig()

	restoreCommandDeps(t, stubSatelliteClient{}, cfg, nil)

	restoreListState(t)
	outputList = "yaml"

	if err := listCmd.RunE(listCmd, []string{"ISS"}); err == nil {
		t.Fatal("listCmd.RunE() expected error, got nil")
	}

	restoreNextState(t)
	output = "yaml"

	if err := nextCmd.RunE(nextCmd, []string{"ISS"}); err == nil {
		t.Fatal("nextCmd.RunE() expected error, got nil")
	}
}

func restoreCommandDeps(t *testing.T, client satelliteClient, cfg appconfig.Config, loadErr error) {
	originalClientFactory := newSatelliteClient
	originalLoadConfig := loadConfig

	newSatelliteClient = func() satelliteClient {
		return client
	}
	loadConfig = func() (appconfig.Config, error) {
		return cfg, loadErr
	}

	t.Cleanup(func() {
		newSatelliteClient = originalClientFactory
		loadConfig = originalLoadConfig
	})
}

func restoreListState(t *testing.T) {
	originalFromTimeList := fromTimeList
	originalLookaheadList := lookaheadList
	originalMinElevationList := minElevationList
	originalOutputList := outputList

	t.Cleanup(func() {
		fromTimeList = originalFromTimeList
		lookaheadList = originalLookaheadList
		minElevationList = originalMinElevationList
		outputList = originalOutputList
	})
}

func restoreNextState(t *testing.T) {
	originalFromTime := fromTime
	originalLookahead := lookahead
	originalMinElevation := minElevation
	originalOutput := output

	t.Cleanup(func() {
		fromTime = originalFromTime
		lookahead = originalLookahead
		minElevation = originalMinElevation
		output = originalOutput
	})
}
