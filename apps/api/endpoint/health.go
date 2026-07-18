package endpoint

import (
	"net/http"

	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
)

func HealthHandler(checker health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := checker.Check(c.Request.Context())

		httpStatusCode := http.StatusServiceUnavailable
		if result.Status == "up" {
			httpStatusCode = http.StatusOK
		}

		c.JSON(httpStatusCode, result)
	}
}
