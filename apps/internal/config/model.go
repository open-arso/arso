package config

const (
	AppName      = "arso"
	ConfigEnvVar = "ARSO_CONFIG"
	ConfigFile   = "config.yaml"
)

type Observer struct {
	Name            string
	LatitudeDeg     float64
	LongitudeDeg    float64
	ElevationMeters float64
}

type Config struct {
	Node        NodeConfig        `mapstructure:"node" json:"node"`
	API         APIConfig         `mapstructure:"api" json:"api"`
	Observatory ObservatoryConfig `mapstructure:"observatory" json:"observatory"`
	Output      OutputConfig      `mapstructure:"output" json:"output"`
}

type NodeConfig struct {
	Name string `mapstructure:"name" json:"name"`
	ID   string `mapstructure:"id" json:"id"`
}

type APIConfig struct {
	URL string `mapstructure:"url" json:"url"`
}

type ObservatoryConfig struct {
	Latitude        *float64 `mapstructure:"latitude" json:"latitude"`
	Longitude       *float64 `mapstructure:"longitude" json:"longitude"`
	ElevationMeters float64  `mapstructure:"elevation_meters" json:"elevation_meters"`
}

type OutputConfig struct {
	Format string `mapstructure:"format" json:"format"`
}
