package endpoint

import (
	"github.com/gin-gonic/gin"
	"github.com/openarso/arso/apps/internal/config"
	"net/http"
)

func ConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := config.Load()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load configuration",
			})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}
