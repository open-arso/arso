package config

import (
	"github.com/gin-gonic/gin"
	"github.com/openarso/arso/apps/internal/config"
	"net/http"
)

type Loader func() (config.Config, error)

func ConfigHandler(load Loader) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := load()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load configuration",
			})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}
