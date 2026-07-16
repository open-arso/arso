package config

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

func (o ObservatoryConfig) IsConfigured() bool {
	return o.Latitude != nil && o.Longitude != nil
}

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
