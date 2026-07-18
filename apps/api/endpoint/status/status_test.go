package status

import (
	"net/http"
	"net/http/httptest"
	"testing"
		"github.com/gin-gonic/gin"
)

func TestStatusHandler(t *testing.T) {
	router := gin.New()

	router.GET("/status", StatusHandler())

	req := httptest.NewRequest(
		http.MethodGet,
		"/status",
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
}