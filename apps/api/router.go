package main

import (
    "time"
	healthChecker "github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"github.com/openarso/arso/apps/api/endpoint/health"
	"github.com/openarso/arso/apps/api/endpoint/config"
	internalConfig "github.com/openarso/arso/apps/internal/config"
	"github.com/openarso/arso/apps/api/endpoint/status"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	checker := healthChecker.NewChecker(
		healthChecker.WithCacheDuration(time.Second),
		healthChecker.WithTimeout(10*time.Second),
	)

	r.GET("/health", health.HealthHandler(checker))
	r.GET("/status", status.StatusHandler())
	r.GET("/config", config.ConfigHandler(internalConfig.Load))

	return r
}