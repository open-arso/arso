package main

import (
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"github.com/openarso/arso/apps/api/endpoint"
)

func main() {
	r := gin.Default()

	checker := health.NewChecker(
		health.WithCacheDuration(time.Second),
		health.WithTimeout(10*time.Second),
	)

	r.GET("/health", endpoint.HealthHandler(checker))
	r.GET("/status", endpoint.StatusHandler())
	r.GET("/config", endpoint.ConfigHandler())

	r.Run()
}
