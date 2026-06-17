package appconfig

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const (
	// AppName names the CLI application directory inside platform config paths.
	AppName = "arso"
	// ConfigEnvVar overrides the default config file path when it is set.
	ConfigEnvVar = "ARSO_CONFIG"
	// ConfigFile is the default filename used inside the ARSO config directory.
	ConfigFile = "config.yaml"
)

// Config contains the persisted CLI settings used by ARSO commands.
type Config struct {
	Node        NodeConfig        `mapstructure:"node" json:"node"`
	API         APIConfig         `mapstructure:"api" json:"api"`
	Observatory ObservatoryConfig `mapstructure:"observatory" json:"observatory"`
	Output      OutputConfig      `mapstructure:"output" json:"output"`
}

// NodeConfig identifies the local node profile used by CLI requests.
type NodeConfig struct {
	Name string `mapstructure:"name" json:"name"`
	ID   string `mapstructure:"id" json:"id"`
}

// APIConfig configures how the CLI reaches the ARSO backend API.
type APIConfig struct {
	URL string `mapstructure:"url" json:"url"`
}

// ObservatoryConfig stores the observer position used for pass and position
// calculations.
type ObservatoryConfig struct {
	Latitude        *float64 `mapstructure:"latitude" json:"latitude"`
	Longitude       *float64 `mapstructure:"longitude" json:"longitude"`
	ElevationMeters float64  `mapstructure:"elevation_meters" json:"elevation_meters"`
}

// IsConfigured reports whether both latitude and longitude are available.
func (o ObservatoryConfig) IsConfigured() bool {
	return o.Latitude != nil && o.Longitude != nil
}

// RequireConfigured returns a user-facing error that explains which
// observatory coordinates are still missing.
func (o ObservatoryConfig) RequireConfigured() error {
	if o.Latitude == nil && o.Longitude == nil {
		return fmt.Errorf(
			"observatory location is not configured. Run:\n" +
				"  arso config set observatory.latitude <latitude>\n" +
				"  arso config set observatory.longitude <longitude>\n" +
				"  arso config set observatory.elevation_meters <meters>",
		)
	}

	if o.Latitude == nil {
		return fmt.Errorf(
			"observatory latitude is not configured. Run:\n" +
				"  arso config set observatory.latitude <latitude>",
		)
	}

	if o.Longitude == nil {
		return fmt.Errorf(
			"observatory longitude is not configured. Run:\n" +
				"  arso config set observatory.longitude <longitude>",
		)
	}

	return nil
}

// OutputConfig stores the preferred default output format for commands that
// support machine-readable responses.
type OutputConfig struct {
	Format string `mapstructure:"format" json:"format"`
}

// Default returns the baseline configuration used when no config file exists.
func Default() Config {
	return Config{
		Node: NodeConfig{
			Name: "local-node",
			ID:   "local",
		},
		API: APIConfig{
			URL: "http://localhost:8080",
		},
		Observatory: ObservatoryConfig{
			Latitude:        nil,
			Longitude:       nil,
			ElevationMeters: 0,
		},
		Output: OutputConfig{
			Format: "text",
		},
	}
}

// Path resolves the config file path, honoring ARSO_CONFIG when it is set.
func Path() (string, error) {
	if customPath := strings.TrimSpace(os.Getenv(ConfigEnvVar)); customPath != "" {
		return customPath, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(configDir, AppName, ConfigFile), nil
}

// Exists reports whether the config file currently exists on disk.
func Exists() (bool, error) {
	path, err := Path()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func newViper() (*viper.Viper, string, error) {
	path, err := Path()
	if err != nil {
		return nil, "", err
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	return v, path, nil
}

// Init creates a config file populated with default values. When overwrite is
// false, Init returns an error if the file already exists.
func Init(overwrite bool) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	exists, err := Exists()
	if err != nil {
		return "", err
	}

	if exists && !overwrite {
		return path, fmt.Errorf("config already exists at %s", path)
	}

	v, _, err := newViper()
	if err != nil {
		return "", err
	}

	setDefaults(v)

	if err := v.WriteConfigAs(path); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return path, nil
}

// Load reads the config file and merges it with default values. Missing config
// files are treated as an empty config instead of an error.
func Load() (Config, error) {
	cfg := Default()

	v, _, err := newViper()
	if err != nil {
		return cfg, err
	}

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}

		if os.IsNotExist(err) {
			return cfg, nil
		}

		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Get returns the current value for a supported config key.
func Get(key string) (any, error) {
	v, _, err := newViper()
	if err != nil {
		return nil, err
	}

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	if !v.IsSet(key) {
		return nil, fmt.Errorf("unknown config key %q", key)
	}

	return v.Get(key), nil
}

// Set validates and persists a supported config key/value pair, returning the
// path to the saved config file.
func Set(key string, rawValue string) (string, error) {
	value, err := parseConfigValue(key, rawValue)
	if err != nil {
		return "", err
	}

	path, err := Path()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	v, _, err := newViper()
	if err != nil {
		return "", err
	}

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if !isConfigNotFound(err) {
			return "", fmt.Errorf("read config: %w", err)
		}
	}

	v.Set(key, value)

	if err := v.WriteConfigAs(path); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return path, nil
}

func parseConfigValue(key string, rawValue string) (any, error) {
	value := strings.TrimSpace(rawValue)

	switch key {
	case "node.name":
		return parseNonEmptyString(key, value)

	case "node.id":
		return parseNodeID(key, value)

	case "api.url":
		return parseAPIURL(key, value)

	case "observatory.latitude":
		return parseFloatInRange(key, value, -90, 90)

	case "observatory.longitude":
		return parseFloatInRange(key, value, -180, 180)

	case "observatory.elevation_meters":
		return parseFiniteFloat(key, value)

	case "output.format":
		return parseOutputFormat(key, value)

	default:
		return nil, fmt.Errorf("unsupported config key %q", key)
	}
}

func parseNonEmptyString(key string, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", key)
	}

	return value, nil
}

func parseNodeID(key string, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", key)
	}

	if strings.ContainsAny(value, " \t\n\r") {
		return "", fmt.Errorf("%s cannot contain whitespace", key)
	}

	return value, nil
}

func parseAPIURL(key string, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", key)
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", fmt.Errorf("%s must be a valid URL: %w", key, err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%s must include scheme and host, for example http://localhost:8080", key)
	}

	return value, nil
}

func parseFloatInRange(key string, value string, min float64, max float64) (float64, error) {
	parsed, err := parseFiniteFloat(key, value)
	if err != nil {
		return 0, err
	}

	if parsed < min || parsed > max {
		return 0, fmt.Errorf("%s must be between %.2f and %.2f", key, min, max)
	}

	return parsed, nil
}

func parseFiniteFloat(key string, value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("%s cannot be empty", key)
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}

	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("%s must be a finite number", key)
	}

	return parsed, nil
}

func parseOutputFormat(key string, value string) (string, error) {
	switch value {
	case "text", "json":
		return value, nil
	default:
		return "", fmt.Errorf("%s must be one of: text, json", key)
	}
}

func isConfigNotFound(err error) bool {
	if os.IsNotExist(err) {
		return true
	}

	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}

func setDefaults(v *viper.Viper) {
	defaultConfig := Default()

	v.SetDefault("node.name", defaultConfig.Node.Name)
	v.SetDefault("node.id", defaultConfig.Node.ID)

	v.SetDefault("api.url", defaultConfig.API.URL)

	v.SetDefault("observatory.elevation_meters", defaultConfig.Observatory.ElevationMeters)

	v.SetDefault("output.format", defaultConfig.Output.Format)
}

func validateKey(key string) error {
	allowed := map[string]bool{
		"node.name":                    true,
		"node.id":                      true,
		"api.url":                      true,
		"observatory.latitude":         true,
		"observatory.longitude":        true,
		"observatory.elevation_meters": true,
		"output.format":                true,
	}

	if !allowed[key] {
		return fmt.Errorf("unsupported config key %q", key)
	}

	return nil
}
