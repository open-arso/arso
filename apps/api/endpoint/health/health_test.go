package health

import (
	"net/http"
	"net/http/httptest"
	"github.com/alexliesenfeld/health"
	"testing"
	"context"

	"github.com/gin-gonic/gin"
)


type fakeHealthChecker struct {
	result health.CheckerResult
}

func (f *fakeHealthChecker) Start() {}

func (f *fakeHealthChecker) Stop() {}

func (f *fakeHealthChecker) Check(ctx context.Context) health.CheckerResult {
	return f.result
}

func (f *fakeHealthChecker) GetRunningPeriodicCheckCount() int {
	return 0
}

func (f *fakeHealthChecker) IsStarted() bool {
	return true
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		checkerResult  health.CheckerResult
		expectedStatus int
	}{
		{
			name: "healthy",
			checkerResult: health.CheckerResult{
				Status: "up",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unhealthy",
			checkerResult: health.CheckerResult{
				Status: "down",
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeHealthChecker{
				result: tt.checkerResult,
			}

			router := gin.New()
			router.GET("/health", HealthHandler(checker))

			req := httptest.NewRequest(
				http.MethodGet,
				"/health",
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}
		})
	}
}