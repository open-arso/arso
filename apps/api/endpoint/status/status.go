package status

import (
	"github.com/gin-gonic/gin"
	"github.com/openarso/arso/apps/internal/node"
	"net/http"
)

func StatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		service := node.NewService()
		status, err := service.Status(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Error getting node status")
			return
		}

		c.JSON(http.StatusOK, status)
	}
}
