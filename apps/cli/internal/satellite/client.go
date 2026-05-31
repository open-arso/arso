package satellite

import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
	"io"

	sgp4 "github.com/akhenakh/sgp4"
)

const DefaultBaseURL = "https://celestrak.org/NORAD/elements/gp.php"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
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

func (c *Client) fetchRaw(ctx context.Context, queryKey string, queryValue string) ([]byte, error) {
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

func (c *Client) Locate(
	ctx context.Context,
	target string,
	observer Observer,
	at time.Time,
) ([]ApparentPosition, error) {
	queryKey, queryValue, err := ResolveTarget(target)
	if err != nil {
		return nil, err
	}

	body, err := c.fetchRaw(ctx, queryKey, queryValue)
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
			RangeRateKms: observation.LookAngles.RangeRate,
			Visible:      observation.LookAngles.Elevation > 0,

			SatelliteLatitudeDeg:  observation.SatellitePos.Latitude,
			SatelliteLongitudeDeg: observation.SatellitePos.Longitude,
			SatelliteAltitudeKm:   observation.SatellitePos.Altitude,
		}

		results = append(results, result)
	}

	return results, nil
}
