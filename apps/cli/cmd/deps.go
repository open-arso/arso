package cmd

import (
	"context"
	"time"

	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/openarso/arso/apps/cli/internal/satellite"
)

type satelliteClient interface {
	Elements(ctx context.Context, target string) ([]satellite.GPElement, error)
	Locate(ctx context.Context, target string, observer satellite.Observer, at time.Time) ([]satellite.ApparentPosition, error)
	PassPredictions(ctx context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error)
	NextPass(ctx context.Context, target string, observer satellite.Observer, at string, lookahead string, minElevation int) (satellite.PassPredictionResult, error)
	CacheResolvedTarget(query string, resolved satellite.ResolvedTarget) error
}

var (
	newSatelliteClient = func() satelliteClient {
		return satellite.NewClient()
	}
	loadConfig = appconfig.Load
)
