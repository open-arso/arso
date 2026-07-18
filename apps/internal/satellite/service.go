package satellite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode"

	sgp4 "github.com/akhenakh/sgp4"
	"github.com/openarso/arso/apps/internal/config"
)

const DefaultBaseURL = "https://celestrak.org/NORAD/elements/gp.php"

type rawFetcher func(
	ctx context.Context,
	queryKey string,
	queryValue string,
) ([]byte, error)

type targetResolver func(
	ctx context.Context,
	target string,
) (ResolvedTarget, error)

type Client struct {
	baseURL       string
	httpClient    *http.Client
	targetCache   *TargetCache
	fetchRaw      rawFetcher
	resolveTarget targetResolver
}

func NewClient() *Client {
	targetCache, err := LoadDefaultTargetCache()
	if err != nil {
		// Cache failure should not prevent ARSO from working.
		// Use an in-memory cache with no path as fallback.
		targetCache = NewTargetCache("")
	}

	client := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		targetCache: targetCache,
	}

	client.fetchRaw = client.fetchRawHTTP
	client.resolveTarget = client.ResolveTarget

	return client
}

func (c *Client) buildURL(queryKey string, queryValue string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	query := baseURL.Query()
	query.Set(queryKey, queryValue)
	query.Set("FORMAT", "JSON")

	baseURL.RawQuery = query.Encode()

	return baseURL.String(), nil
}

func (c *Client) fetchRawHTTP(ctx context.Context, queryKey string, queryValue string) ([]byte, error) {
	apiURL, err := c.buildURL(queryKey, queryValue)
	if err != nil {
		return nil, fmt.Errorf("build CelesTrak URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create CelesTrak request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call CelesTrak API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CelesTrak returned status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read CelesTrak response: %w", err)
	}

	return body, nil
}

func (c *Client) Elements(ctx context.Context, target string) ([]GPElement, error) {
	queryKey, queryValue, err := BuildCelesTrakQuery(target)
	if err != nil {
		return nil, err
	}

	elements, err := c.Fetch(ctx, queryKey, queryValue)
	if err != nil {
		return nil, err
	}

	return elements, nil
}

func (c *Client) Fetch(ctx context.Context, queryKey string, queryValue string) ([]GPElement, error) {
	body, err := c.fetchRaw(ctx, queryKey, queryValue)
	if err != nil {
		return nil, err
	}

	var elements []GPElement
	if err := json.Unmarshal(body, &elements); err != nil {
		return nil, fmt.Errorf("decode CelesTrak response: %w", err)
	}

	return elements, nil
}

func (c *Client) PassPredictions(
	ctx context.Context,
	target string,
	observer config.Observer,
	at string,
	lookahead string,
	minElevation int,
) (PassPredictionResult, error) {
	startTime, err := parseTimeStr(at)
	if err != nil {
		return PassPredictionResult{}, err
	}

	stopTime, err := computeLookaheadTime(startTime, lookahead)
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("compute stop time based on lookahead: %w", err)
	}

	resolved, err := c.resolveTarget(ctx, target)
	if err != nil {
		return PassPredictionResult{}, err
	}

	body, err := c.fetchRaw(ctx, QueryCATNR, strconv.Itoa(resolved.NoradID))
	if err != nil {
		return PassPredictionResult{}, err
	}

	omms, err := sgp4.ParseOMMs(body)
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("parse CelesTrak OMM data: %w", err)
	}

	if len(omms) == 0 {
		return PassPredictionResult{}, fmt.Errorf("no object found for %q", target)
	}

	omm := omms[0]

	tle, err := omm.ToTLE()
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("convert OMM to TLE for %s: %w", omm.ObjectName, err)
	}

	const stepSeconds = 30

	passes, err := tle.GeneratePasses(
		observer.LatitudeDeg,
		observer.LongitudeDeg,
		observer.ElevationMeters,
		startTime,
		stopTime,
		stepSeconds,
	)
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("generate passes: %w", err)
	}

	predictedPasses := make([]PredictedPass, 0, len(passes))

	for _, pass := range passes {
		if pass.MaxElevation < float64(minElevation) {
			continue
		}

		predictedPasses = append(predictedPasses, PredictedPass{
			AcquisitionOfSignal: pass.AOS,
			LossOfSignal:        pass.LOS,
			Duration:            pass.Duration,
			MaxElevation:        pass.MaxElevation,
			MaxElevationTime:    pass.MaxElevationTime,
			AzimuthAtAOS:        pass.AOSAzimuth,
			AzimuthAtLOS:        pass.LOSAzimuth,
		})
	}

	return PassPredictionResult{
		Name:         omm.ObjectName,
		Kind:         "satellite",
		Source:       "celestrak_sgp4",
		NoradID:      omm.NoradCatID,
		ObjectID:     omm.ObjectID,
		ObserverName: observer.Name,
		Passes:       predictedPasses,
	}, nil
}

func parseTimeStr(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"invalid datetime value %q: expected RFC3339 format like 2026-06-03T22:00:00Z",
			value,
		)
	}

	return t.UTC(), nil
}

func computeLookaheadTime(start time.Time, lookahead string) (time.Time, error) {
	timeValue, unitString, err := splitLookahead(lookahead)
	if err != nil {
		return time.Time{}, err
	}

	switch unitString {
	case "Y":
		return start.AddDate(timeValue, 0, 0), nil
	case "M":
		return start.AddDate(0, timeValue, 0), nil
	case "d":
		return start.AddDate(0, 0, timeValue), nil
	case "h":
		return start.Add(time.Hour * time.Duration(timeValue)), nil
	case "m":
		return start.Add(time.Minute * time.Duration(timeValue)), nil
	case "s":
		return start.Add(time.Second * time.Duration(timeValue)), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported lookahead unit %q", unitString)
	}
}

func splitLookahead(lookahead string) (int, string, error) {
	if lookahead == "" {
		return 0, "", fmt.Errorf("lookahead cannot be empty")
	}

	splitIndex := 0

	for _, r := range lookahead {
		if !unicode.IsDigit(r) {
			break
		}

		splitIndex++
	}

	if splitIndex == 0 {
		return 0, "", fmt.Errorf("lookahead must start with a number: %q", lookahead)
	}

	if splitIndex == len(lookahead) {
		return 0, "", fmt.Errorf("lookahead must contain a unit: %q", lookahead)
	}

	numberPart := lookahead[:splitIndex]
	unitPart := lookahead[splitIndex:]

	value, err := strconv.Atoi(numberPart)
	if err != nil {
		return 0, "", fmt.Errorf("invalid lookahead value %q: %w", numberPart, err)
	}

	if value <= 0 {
		return 0, "", fmt.Errorf("lookahead value must be positive: %d", value)
	}

	// Validate unit
	validUnits := map[string]bool{
		"Y": true,
		"M": true,
		"d": true,
		"h": true,
		"m": true,
		"s": true,
	}
	if !validUnits[unitPart] {
		return 0, "", fmt.Errorf("invalid lookahead unit %q, must be one of: Y, M, d, h, m, s", unitPart)
	}

	return value, unitPart, nil
}

func (c *Client) NextPass(
	ctx context.Context,
	target string,
	observer config.Observer,
	at string,
	lookahead string,
	minElevation int,
) (PassPredictionResult, error) {
	startTime, err := parseTimeStr(at)
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("invalid start search time: %w", err)
	}

	maxEnd, err := computeLookaheadTime(startTime, lookahead)
	if err != nil {
		return PassPredictionResult{}, fmt.Errorf("invalid lookahead value: %w", err)
	}

	maxDuration := maxEnd.Sub(startTime)
	if maxDuration <= 0 {
		return PassPredictionResult{}, fmt.Errorf("lookahead must be positive")
	}

	window := time.Hour

	for {
		if window > maxDuration {
			window = maxDuration
		}

		currentLookahead, err := formatLookaheadDuration(window)
		if err != nil {
			return PassPredictionResult{}, err
		}

		predictions, err := c.PassPredictions(
			ctx,
			target,
			observer,
			at,
			currentLookahead,
			minElevation,
		)
		if err != nil {
			return PassPredictionResult{}, fmt.Errorf("error in next pass predictions: %w", err)
		}

		if len(predictions.Passes) > 0 {
			predictions.Passes = predictions.Passes[:1]
			return predictions, nil
		}

		if window == maxDuration {
			break
		}

		window *= 2
	}

	return PassPredictionResult{}, fmt.Errorf(
		"no pass found for %q above %d° in the next %s",
		target,
		minElevation,
		lookahead,
	)
}

func formatLookaheadDuration(d time.Duration) (string, error) {
	if d <= 0 {
		return "", fmt.Errorf("lookahead duration must be positive")
	}

	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", int(d/(24*time.Hour))), nil
	}

	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour)), nil
	}

	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute)), nil
	}

	if d%time.Second == 0 {
		return fmt.Sprintf("%ds", int(d/time.Second)), nil
	}

	return "", fmt.Errorf("unsupported lookahead duration precision: %s", d)
}

func (c *Client) Locate(
	ctx context.Context,
	target string,
	observer config.Observer,
	at time.Time,
) ([]ApparentPosition, error) {
	resolved, err := c.resolveTarget(ctx, target)
	if err != nil {
		return nil, err
	}

	body, err := c.fetchRaw(ctx, QueryCATNR, strconv.Itoa(resolved.NoradID))
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	omms, err := sgp4.ParseOMMs(body)
	if err != nil {
		return nil, fmt.Errorf("parse CelesTrak OMM data: %w", err)
	}

	if len(omms) == 0 {
		return nil, fmt.Errorf("no object found for %q", target)
	}

	location := &sgp4.Location{
		Latitude:  observer.LatitudeDeg,
		Longitude: observer.LongitudeDeg,
		Altitude:  observer.ElevationMeters,
	}

	at = at.UTC()

	results := make([]ApparentPosition, 0, len(omms))

	for _, omm := range omms {
		tle, err := omm.ToTLE()
		if err != nil {
			return nil, fmt.Errorf("convert OMM to TLE for %s: %w", omm.ObjectName, err)
		}

		eciState, err := tle.FindPositionAtTime(at)
		if err != nil {
			return nil, fmt.Errorf("propagate %s at %s: %w", omm.ObjectName, at.Format(time.RFC3339), err)
		}

		stateVector := &sgp4.StateVector{
			X:  eciState.Position.X,
			Y:  eciState.Position.Y,
			Z:  eciState.Position.Z,
			VX: eciState.Velocity.X,
			VY: eciState.Velocity.Y,
			VZ: eciState.Velocity.Z,
		}

		observation, err := stateVector.GetLookAngle(location, at)
		if err != nil {
			return nil, fmt.Errorf("calculate look angle for %s: %w", omm.ObjectName, err)
		}

		result := ApparentPosition{
			Name:         omm.ObjectName,
			Kind:         "satellite",
			Source:       "celestrak_sgp4",
			NoradID:      omm.NoradCatID,
			ObjectID:     omm.ObjectID,
			ObserverName: observer.Name,
			TimeUTC:      at.Format(time.RFC3339),

			AzimuthDeg:   observation.LookAngles.Azimuth,
			ElevationDeg: observation.LookAngles.Elevation,
			RangeKm:      observation.LookAngles.Range,
			RangeRateKms: observation.LookAngles.RangeRate / 1000.0,
			AboveHorizon: observation.LookAngles.Elevation > 0,

			SatelliteLatitudeDeg:  observation.SatellitePos.Latitude,
			SatelliteLongitudeDeg: observation.SatellitePos.Longitude,
			SatelliteAltitudeKm:   observation.SatellitePos.Altitude,
		}

		results = append(results, result)
	}

	return results, nil
}

func (c *Client) CacheResolvedTarget(query string, resolved ResolvedTarget) error {
	normalized := normalizeTarget(query)

	if normalized == "" {
		return fmt.Errorf("target cannot be empty")
	}

	if c.targetCache == nil {
		return nil
	}

	return c.targetCache.SetTarget(normalized, resolved)
}
