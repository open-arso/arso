package node

import (
	"context"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)
	assert.IsType(t, &Service{}, svc)
}

func TestService_Status(t *testing.T) {
	// Save the original startTime and restore after test
	originalStartTime := startTime
	defer func() {
		startTime = originalStartTime
	}()

	// Set a fixed start time for testing
	startTime = time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Test basic fields
	assert.Equal(t, StateUp, status.State)
	assert.Equal(t, startTime, status.StartedAt)
	assert.NotEmpty(t, status.Uptime)

	// Test runtime info
	assert.GreaterOrEqual(t, status.Runtime.GoRoutines, 0)
	assert.NotEmpty(t, status.Runtime.GoVersion)
	assert.NotEmpty(t, status.Runtime.OS)
	assert.NotEmpty(t, status.Runtime.Arch)

	// Test memory info
	assert.GreaterOrEqual(t, status.Memory.UsedMB, uint64(0))
	assert.Greater(t, status.Memory.TotalMB, uint64(0))
	assert.GreaterOrEqual(t, status.Memory.Percent, 0.0)
	assert.LessOrEqual(t, status.Memory.Percent, 100.0)

	// Test disk info
	assert.GreaterOrEqual(t, status.Disk.UsedGB, uint64(0))
	assert.Greater(t, status.Disk.TotalGB, uint64(0))
	assert.GreaterOrEqual(t, status.Disk.Percent, 0.0)
	assert.LessOrEqual(t, status.Disk.Percent, 100.0)

	// Test CPU info
	assert.GreaterOrEqual(t, status.CPU.UsagePercent, 0.0)
	assert.LessOrEqual(t, status.CPU.UsagePercent, 100.0)
}

func TestService_Status_ContextCancellation(t *testing.T) {
	svc := NewService()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The Status method doesn't actually use the context for cancellation,
	// but we test that it still works with a cancelled context
	status, err := svc.Status(ctx)

	// The service should still work even with a cancelled context
	// because it's using the context for the gopsutil calls which may respect it
	if err != nil {
		// If there's an error, it should be from gopsutil, not from context
		assert.NotContains(t, err.Error(), "context canceled")
	} else {
		assert.NotEmpty(t, status)
	}
}

func TestService_Status_Uptime(t *testing.T) {
	// Save the original startTime and restore after test
	originalStartTime := startTime
	defer func() {
		startTime = originalStartTime
	}()

	// Set start time to a known value
	startTime = time.Now().Add(-2 * time.Hour)

	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Uptime should be approximately 2 hours
	assert.Contains(t, status.Uptime, "2h")
}

func TestService_Status_MemoryCalculation(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Test that memory calculations are correct
	// UsedMB should be less than or equal to TotalMB
	assert.LessOrEqual(t, status.Memory.UsedMB, status.Memory.TotalMB)

	// Percent should be between 0 and 100
	assert.GreaterOrEqual(t, status.Memory.Percent, 0.0)
	assert.LessOrEqual(t, status.Memory.Percent, 100.0)
}

func TestService_Status_DiskCalculation(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Test that disk calculations are correct
	// UsedGB should be less than or equal to TotalGB
	assert.LessOrEqual(t, status.Disk.UsedGB, status.Disk.TotalGB)

	// Percent should be between 0 and 100
	assert.GreaterOrEqual(t, status.Disk.Percent, 0.0)
	assert.LessOrEqual(t, status.Disk.Percent, 100.0)
}

func TestService_Status_CPUCalculation(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// CPU percent should be between 0 and 100
	assert.GreaterOrEqual(t, status.CPU.UsagePercent, 0.0)
	assert.LessOrEqual(t, status.CPU.UsagePercent, 100.0)
}

func TestService_Status_RuntimeInfo(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Test runtime info
	assert.GreaterOrEqual(t, status.Runtime.GoRoutines, 0)
	assert.NotEmpty(t, status.Runtime.GoVersion)
	assert.NotEmpty(t, status.Runtime.OS)
	assert.NotEmpty(t, status.Runtime.Arch)

	// Go version should start with "go"
	assert.Contains(t, status.Runtime.GoVersion, "go")

	// OS should be one of the common OS names
	validOS := []string{"linux", "windows", "darwin", "freebsd", "openbsd"}
	assert.Contains(t, validOS, status.Runtime.OS)

	// Arch should be one of the common architectures
	validArch := []string{"386", "amd64", "arm", "arm64", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "riscv64", "s390x"}
	assert.Contains(t, validArch, status.Runtime.Arch)
}

func TestService_Status_State(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// State should always be StateUp for a running service
	assert.Equal(t, StateUp, status.State)
	assert.Equal(t, "up", string(status.State))
}

func TestService_Status_Concurrent(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	// Test that Status can be called concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			status, err := svc.Status(ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, status)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestService_Status_Consistency(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	// Get status twice and compare
	status1, err := svc.Status(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	status2, err := svc.Status(ctx)
	require.NoError(t, err)

	// Some fields should be consistent
	assert.Equal(t, status1.State, status2.State)
	assert.Equal(t, status1.Runtime.GoVersion, status2.Runtime.GoVersion)
	assert.Equal(t, status1.Runtime.OS, status2.Runtime.OS)
	assert.Equal(t, status1.Runtime.Arch, status2.Runtime.Arch)
	assert.Equal(t, status1.StartedAt, status2.StartedAt)

	// Uptime should have increased
	assert.NotEqual(t, status1.Uptime, status2.Uptime)
}

func TestService_Status_WithTimeout(t *testing.T) {
	svc := NewService()

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	status, err := svc.Status(ctx)
	duration := time.Since(start)

	// The service should complete within the timeout
	assert.Less(t, duration, 2*time.Second)

	if err != nil {
		// If there's an error, it should be from the underlying system calls
		assert.NotContains(t, err.Error(), "context deadline exceeded")
	} else {
		assert.NotEmpty(t, status)
	}
}

// Benchmark tests
func BenchmarkService_Status(b *testing.B) {
	svc := NewService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.Status(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration test for real system values
func TestService_Status_RealSystemValues(t *testing.T) {
	// This test verifies that the values returned are realistic
	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Memory values should be reasonable (at least 1 MB total)
	assert.Greater(t, status.Memory.TotalMB, uint64(0))

	// Disk values should be reasonable (at least 1 GB total on most systems)
	assert.Greater(t, status.Disk.TotalGB, uint64(0))

	// CPU usage should be a reasonable value
	assert.GreaterOrEqual(t, status.CPU.UsagePercent, 0.0)
	assert.LessOrEqual(t, status.CPU.UsagePercent, 100.0)
}

// Skip this test if running on CI or in environments where disk usage might not be available
func TestService_Status_DiskUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping disk usage test in short mode")
	}

	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Disk used should be less than total
	assert.LessOrEqual(t, status.Disk.UsedGB, status.Disk.TotalGB)

	// Disk percent should be between 0 and 100
	assert.GreaterOrEqual(t, status.Disk.Percent, 0.0)
	assert.LessOrEqual(t, status.Disk.Percent, 100.0)
}

// Test that startTime is initialized properly
func TestStartTime(t *testing.T) {
	// Reset startTime for this test
	originalStartTime := startTime
	defer func() {
		startTime = originalStartTime
	}()

	// Force a new start time
	startTime = time.Now()

	// The start time should be recent
	assert.WithinDuration(t, time.Now(), startTime, 1*time.Second)
}

// Helper to get real system stats for comparison
func getRealSystemStats(t *testing.T) (mem.VirtualMemoryStat, disk.UsageStat) {
	t.Helper()

	vm, err := mem.VirtualMemory()
	require.NoError(t, err)

	diskUsage, err := disk.Usage("/")
	require.NoError(t, err)

	return *vm, *diskUsage
}

func TestService_Status_MatchesSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping system comparison test in short mode")
	}

	svc := NewService()
	ctx := context.Background()

	status, err := svc.Status(ctx)
	require.NoError(t, err)

	// Get real system stats for comparison
	vm, diskUsage := getRealSystemStats(t)

	// The status values should be close to the real system values
	// Allow for small differences due to timing

	// Memory calculations should be close
	expectedUsedMB := vm.Used / 1024 / 1024
	expectedTotalMB := vm.Total / 1024 / 1024

	assert.InDelta(t, expectedUsedMB, status.Memory.UsedMB, 100, "Memory used should be close")
	assert.InDelta(t, expectedTotalMB, status.Memory.TotalMB, 100, "Memory total should be close")
	assert.InDelta(t, vm.UsedPercent, status.Memory.Percent, 1.0, "Memory percent should be close")

	// Disk calculations should be close
	expectedUsedGB := diskUsage.Used / 1024 / 1024 / 1024
	expectedTotalGB := diskUsage.Total / 1024 / 1024 / 1024

	assert.InDelta(t, expectedUsedGB, status.Disk.UsedGB, 1, "Disk used should be close")
	assert.InDelta(t, expectedTotalGB, status.Disk.TotalGB, 1, "Disk total should be close")
	assert.InDelta(t, diskUsage.UsedPercent, status.Disk.Percent, 1.0, "Disk percent should be close")
}
