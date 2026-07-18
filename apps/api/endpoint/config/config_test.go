package config

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	internalconfig "github.com/openarso/arso/apps/internal/config"
)

func TestConfigHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns configuration", func(t *testing.T) {
		latitude := 48.8566
		longitude := 2.3522

		expected := internalconfig.Config{
			Node: internalconfig.NodeConfig{
				Name: "test-node",
				ID:   "test-id",
			},
			API: internalconfig.APIConfig{
				URL: "http://localhost:8080",
			},
			Observatory: internalconfig.ObservatoryConfig{
				Latitude:        &latitude,
				Longitude:       &longitude,
				ElevationMeters: 35,
			},
			Output: internalconfig.OutputConfig{
				Format: "json",
			},
		}

		load := func() (internalconfig.Config, error) {
			return expected, nil
		}

		router := gin.New()
		router.GET("/config", ConfigHandler(load))

		req := httptest.NewRequest(
			http.MethodGet,
			"/config",
			nil,
		)

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				recorder.Code,
			)
		}

		var actual internalconfig.Config

		if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if actual.Node.Name != expected.Node.Name {
			t.Errorf(
				"expected node name %q, got %q",
				expected.Node.Name,
				actual.Node.Name,
			)
		}

		if actual.Node.ID != expected.Node.ID {
			t.Errorf(
				"expected node ID %q, got %q",
				expected.Node.ID,
				actual.Node.ID,
			)
		}

		if actual.API.URL != expected.API.URL {
			t.Errorf(
				"expected API URL %q, got %q",
				expected.API.URL,
				actual.API.URL,
			)
		}

		if actual.Observatory.Latitude == nil {
			t.Fatal("expected latitude, got nil")
		}

		if *actual.Observatory.Latitude != latitude {
			t.Errorf(
				"expected latitude %f, got %f",
				latitude,
				*actual.Observatory.Latitude,
			)
		}
	})

	t.Run("returns internal server error when loading fails", func(t *testing.T) {
		load := func() (internalconfig.Config, error) {
			return internalconfig.Config{}, errors.New("configuration unavailable")
		}

		router := gin.New()
		router.GET("/config", ConfigHandler(load))

		req := httptest.NewRequest(
			http.MethodGet,
			"/config",
			nil,
		)

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				recorder.Code,
			)
		}

		expectedBody := `{"error":"failed to load configuration"}`

		if recorder.Body.String() != expectedBody {
			t.Fatalf(
				"expected body %s, got %s",
				expectedBody,
				recorder.Body.String(),
			)
		}
	})
}