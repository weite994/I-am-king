package profiler

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiler_ProfileFunc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	profiler := New(logger, true)

	ctx := context.Background()
	executed := false

	profile, err := profiler.ProfileFunc(ctx, "test_operation", func() error {
		executed = true
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
	assert.NotNil(t, profile)
	assert.Equal(t, "test_operation", profile.Operation)
	assert.True(t, profile.Duration >= 10*time.Millisecond)
	assert.False(t, profile.Timestamp.IsZero())
}

func TestProfiler_ProfileFuncWithMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	profiler := New(logger, true)

	ctx := context.Background()
	executed := false

	profile, err := profiler.ProfileFuncWithMetrics(ctx, "test_operation_with_metrics", func() (int, int64, error) {
		executed = true
		time.Sleep(5 * time.Millisecond)
		return 100, 1024, nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
	assert.NotNil(t, profile)
	assert.Equal(t, "test_operation_with_metrics", profile.Operation)
	assert.True(t, profile.Duration >= 5*time.Millisecond)
	assert.Equal(t, 100, profile.LinesCount)
	assert.Equal(t, int64(1024), profile.BytesCount)
	assert.False(t, profile.Timestamp.IsZero())
}

func TestProfiler_Start(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	profiler := New(logger, true)

	ctx := context.Background()
	finish := profiler.Start(ctx, "test_start_operation")

	time.Sleep(5 * time.Millisecond)
	profile := finish(50, 512)

	assert.NotNil(t, profile)
	assert.Equal(t, "test_start_operation", profile.Operation)
	assert.True(t, profile.Duration >= 5*time.Millisecond)
	assert.Equal(t, 50, profile.LinesCount)
	assert.Equal(t, int64(512), profile.BytesCount)
	assert.False(t, profile.Timestamp.IsZero())
}

func TestProfiler_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	profiler := New(logger, false)

	ctx := context.Background()
	executed := false

	profile, err := profiler.ProfileFunc(ctx, "test_operation", func() error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
	assert.Nil(t, profile) // Should be nil when disabled
}

func TestProfile_String(t *testing.T) {
	profile := &Profile{
		Operation:   "test_operation",
		Duration:    100 * time.Millisecond,
		MemoryDelta: 1024,
		LinesCount:  100,
		BytesCount:  2048,
		Timestamp:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	str := profile.String()
	assert.Contains(t, str, "test_operation")
	assert.Contains(t, str, "100ms")
	assert.Contains(t, str, "1024B")
	assert.Contains(t, str, "lines=100")
	assert.Contains(t, str, "bytes=2048")
	assert.Contains(t, str, "12:00:00.000")
}

func TestGlobalProfiler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	Init(logger, true)

	ctx := context.Background()
	executed := false

	profile, err := ProfileFunc(ctx, "global_test", func() error {
		executed = true
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
	assert.NotNil(t, profile)
	assert.Equal(t, "global_test", profile.Operation)
	assert.True(t, profile.Duration >= 5*time.Millisecond)
}

func TestMemoryProfiling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	profiler := New(logger, true)

	ctx := context.Background()

	profile, err := profiler.ProfileFunc(ctx, "memory_allocation_test", func() error {
		data := make([]string, 1000)
		for i := range data {
			data[i] = strings.Repeat("test", 100)
		}
		_ = len(data)
		return nil
	})

	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "memory_allocation_test", profile.Operation)
	assert.NotEqual(t, int64(0), profile.MemoryBefore)
	assert.NotEqual(t, int64(0), profile.MemoryAfter)
}

func TestIsProfilingEnabled(t *testing.T) {
	assert.False(t, IsProfilingEnabled())

	os.Setenv("GITHUB_MCP_PROFILING_ENABLED", "true")
	defer os.Unsetenv("GITHUB_MCP_PROFILING_ENABLED")
	assert.True(t, IsProfilingEnabled())

	os.Setenv("GITHUB_MCP_PROFILING_ENABLED", "false")
	assert.False(t, IsProfilingEnabled())

	os.Setenv("GITHUB_MCP_PROFILING_ENABLED", "invalid")
	assert.False(t, IsProfilingEnabled())
}

func TestInitFromEnv(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	os.Setenv("GITHUB_MCP_PROFILING_ENABLED", "true")
	defer os.Unsetenv("GITHUB_MCP_PROFILING_ENABLED")

	InitFromEnv(logger)
	assert.NotNil(t, globalProfiler)
	assert.True(t, globalProfiler.enabled)

	os.Setenv("GITHUB_MCP_PROFILING_ENABLED", "false")
	InitFromEnv(logger)
	assert.NotNil(t, globalProfiler)
	assert.False(t, globalProfiler.enabled)
}
