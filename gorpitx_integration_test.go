package gorpitx

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPITX_StartInGoroutineAndStop_Integration(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Test streaming by starting a real commander process
	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Start execution in a goroutine
	errCh := make(chan error, 1)

	go func() {
		err := rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 30*time.Second)
		errCh <- err
	}()

	// Wait a bit for execution to start
	time.Sleep(200 * time.Millisecond)

	// Verify it's executing
	assert.True(t, rpitx.isExecuting.Load(), "RPITX should be executing")

	// Stop execution
	stopErr := rpitx.Stop(ctx)
	// Stop should succeed or return expected termination errors
	if stopErr != nil && !errors.Is(stopErr, commonerrors.ErrTerminated) && !errors.Is(stopErr, commonerrors.ErrKilled) {
		t.Errorf("unexpected stop error: %v", stopErr)
	}

	// Wait for execution to complete
	select {
	case execErr := <-errCh:
		// Execution should complete (possibly with expected termination errors)
		if execErr != nil && !errors.Is(execErr, commonerrors.ErrTerminated) && !errors.Is(execErr, commonerrors.ErrKilled) {
			t.Errorf("unexpected execution error: %v", execErr)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("execution should have completed after stop")
	}

	// Verify no longer executing
	assert.False(t, rpitx.isExecuting.Load(), "RPITX should not be executing after stop")
}

func TestRPITX_ConcurrentExecution_Integration(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Test multiple streams with real commander
	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Start first execution in a goroutine - it will run the mock infinite loop
	errCh := make(chan error, 1)

	go func() {
		err := rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 500*time.Millisecond)
		errCh <- err
	}()

	// Wait a bit to ensure first execution started and acquired the lock
	time.Sleep(50 * time.Millisecond)

	// Try to start second execution - should fail with ErrExecuting
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 1*time.Second)

	assert.ErrorIs(t, err, ErrExecuting, "second execution should fail with ErrExecuting")

	// Wait for first execution to complete or timeout
	select {
	case firstErr := <-errCh:
		// First execution should complete (with possible timeout or termination)
		if firstErr != nil {
			// Timeout or termination errors are expected in this test
			t.Logf("First execution completed with: %v", firstErr)
		}
	case <-time.After(2 * time.Second):
		// If still running, stop it
		_ = rpitx.Stop(ctx) // Best effort stop

		<-errCh // Wait for completion
	}

	// Verify no longer executing
	assert.False(t, rpitx.isExecuting.Load(), "RPITX should not be executing after completion")
}

func TestRPITX_Stop_Integration(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Use real commander to actually execute shell commands
	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Start execution in a goroutine
	errCh := make(chan error, 1)

	go func() {
		err := rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 30*time.Second)
		errCh <- err
	}()

	// Wait for execution to start
	time.Sleep(100 * time.Millisecond)

	// Stop execution
	stopErr := rpitx.Stop(ctx)
	// Stop should succeed or return expected termination errors
	if stopErr != nil && !errors.Is(stopErr, commonerrors.ErrTerminated) && !errors.Is(stopErr, commonerrors.ErrKilled) {
		t.Errorf("unexpected stop error: %v", stopErr)
	}

	// Wait for execution to complete
	select {
	case <-errCh:
		// Execution completed
	case <-time.After(10 * time.Second):
		t.Log("execution did not complete within expected time, continuing")
	}

	// Verify no longer executing
	assert.False(t, rpitx.isExecuting.Load(), "RPITX should not be executing after stop")

	// Should be able to start new execution after stop
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 100*time.Millisecond)
	assert.Error(t, err) // Will timeout but should not be ErrExecuting
	assert.NotErrorIs(t, err, ErrExecuting, "Should not be busy after previous execution stopped")
}

func TestRPITX_DevExecution_Integration(t *testing.T) {
	// Test dev environment execution
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	// Create args with all defaults specified to avoid random generation
	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
		"pi":    "1234",
		"ps":    "TEST FM",
		"rt":    "Test Radio Text",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Should execute dev mock path and succeed
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 100*time.Millisecond)
	// Dev execution should succeed or timeout (both are acceptable)
	if err != nil {
		// Timeout is expected in dev mode due to mock infinite loop
		assert.Contains(t, err.Error(), "timeout", "Dev execution should timeout due to infinite loop")
	}
}

func TestRPITX_DevExecution_InvalidModule_Integration(t *testing.T) {
	// Test unknown module with real commander (should fail before execution)
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	ctx := context.Background()

	// Should fail with unknown module error
	err := rpitx.Exec(ctx, "unknown", []byte(`{"freq": 107.9}`), 1*time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown", "should contain unknown module error")
}

func TestRPITX_DevExecution_InvalidArgs_Integration(t *testing.T) {
	// Test invalid JSON with real commander
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	ctx := context.Background()

	// Should fail with JSON parse error
	err := rpitx.Exec(ctx, ModuleNamePIFMRDS, []byte(`{invalid json`), 1*time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse args", "should contain parse args error")
}

func TestRPITX_DevExecution_ParseArgsFailure_Integration(t *testing.T) {
	// Test module ParseArgs failure with real commander
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	realCommander := commander.New()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	ctx := context.Background()

	// Should fail because audio file is required
	err := rpitx.Exec(ctx, ModuleNamePIFMRDS, []byte(`{"freq": 107.9}`), 1*time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse args", "should contain parse args error")
}

func TestRPITX_ProductionExecution_WaitFailure_Integration(t *testing.T) {
	// Test production execution wait failure using real commander with non-existent command
	t.Setenv(env.EnvVarName, env.EnvTypeProd)

	// Use real commander to test actual wait failure
	realCommander := commander.New()
	rpitx := &RPITX{
		config: Config{Path: "/tmp/nonexistent"},
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: realCommander,
	}

	// Create args that would work but use non-existent audio file
	args := map[string]any{
		"freq":  107.9,
		"audio": "/nonexistent/file.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Should fail because audio file doesn't exist
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 1*time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse args", "should fail during args parsing")
}

func TestRPITX_ConcurrentStopCalls_Integration(t *testing.T) {
	// Test multiple stop calls don't cause issues
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	rpitx := createTestRPITXInstance()
	ctx := context.Background()
	argsBytes := createTestArgsBytes(t)

	// Start execution
	execErrCh := make(chan error, 1)

	go func() {
		err := rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 30*time.Second)
		execErrCh <- err
	}()

	time.Sleep(100 * time.Millisecond) // Let execution start

	// Call stop multiple times concurrently and collect results
	stopErrors := performConcurrentStops(ctx, t, rpitx)

	// Analyze stop results
	successCount, notExecutingCount := analyzeStopResults(t, stopErrors)

	// Verify expectations
	assert.GreaterOrEqual(t, successCount, 1, "At least one stop should succeed")
	assert.Equal(t, 3, successCount+notExecutingCount, "All stops should either succeed or return ErrNotExecuting")

	// Wait for execution to complete
	select {
	case <-execErrCh:
		// Execution completed
	case <-time.After(10 * time.Second):
		t.Fatal("Execution should have stopped")
	}
}

func createTestRPITXInstance() *RPITX {
	return &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: commander.New(),
	}
}

func createTestArgsBytes(t *testing.T) []byte {
	t.Helper()

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}
	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	return argsBytes
}

func performConcurrentStops(
	ctx context.Context,
	t *testing.T,
	rpitx *RPITX,
) []error {
	t.Helper()

	stopErrCh1 := make(chan error, 1)
	stopErrCh2 := make(chan error, 1)
	stopErrCh3 := make(chan error, 1)

	go func() { stopErrCh1 <- rpitx.Stop(ctx) }()
	go func() { stopErrCh2 <- rpitx.Stop(ctx) }()
	go func() { stopErrCh3 <- rpitx.Stop(ctx) }()

	return []error{<-stopErrCh1, <-stopErrCh2, <-stopErrCh3}
}

func analyzeStopResults(
	t *testing.T,
	stopErrors []error,
) (int, int) {
	t.Helper()

	successCount := 0
	notExecutingCount := 0

	for _, stopErr := range stopErrors {
		switch {
		case stopErr == nil:
			successCount++
		case errors.Is(stopErr, ErrNotExecuting):
			notExecutingCount++
		case errors.Is(stopErr, commonerrors.ErrTerminated) || errors.Is(stopErr, commonerrors.ErrKilled):
			// Process termination errors are expected when stopping
			successCount++

			t.Logf("Stop returned expected termination error: %v", stopErr)
		default:
			t.Errorf("Unexpected stop error: %v", stopErr)
		}
	}

	return successCount, notExecutingCount
}

func TestRPITX_StreamOutputs_WithRealExecution_Integration(t *testing.T) {
	// Use dev environment so RPITX will run the actual mock shell command
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	ctx := context.Background()
	rpitx := setupRealExecutionTest()

	// Start execution and streaming
	execDone, execErr := startExecutionForStreaming(ctx, rpitx)
	stdout, stderr := createStreamingChannels()

	// Set up output collection
	receivedLines, streamMu := setupOutputCollection()
	collectDone := startOutputCollection(stdout, stderr, receivedLines, streamMu)

	// Run test and collect output
	runStreamingTest(ctx, rpitx, stdout, stderr, collectDone, execDone)

	// Verify results
	verifyStreamingResults(t, execErr, rpitx, receivedLines, streamMu)
}

func setupRealExecutionTest() *RPITX {
	// Create RPITX manually to avoid root check in newRPITX()
	config, err := parseConfig()
	if err != nil {
		panic(err)
	}

	rpitx := &RPITX{
		config:    config,
		commander: commander.New(), // Real commander, not mock!
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
	}

	return rpitx
}

func startExecutionForStreaming(
	ctx context.Context,
	rpitx *RPITX,
) (chan struct{}, *error) {
	var execErr error

	execDone := make(chan struct{})

	go func() {
		defer close(execDone)

		execErr = rpitx.Exec(ctx, ModuleNamePIFMRDS, []byte(`{"freq": 107.9, "audio": ".fixtures/test.wav"}`), 10*time.Second)
	}()

	time.Sleep(50 * time.Millisecond) // Wait briefly for execution to start

	return execDone, &execErr
}

func createStreamingChannels() (chan string, chan string) {
	stdout := make(chan string, 50)
	stderr := make(chan string, 10)

	return stdout, stderr
}

func setupOutputCollection() (*[]string, *sync.Mutex) {
	var (
		receivedLines []string
		streamMu      sync.Mutex
	)

	return &receivedLines, &streamMu
}

func startOutputCollection(
	stdout, stderr chan string,
	receivedLines *[]string,
	streamMu *sync.Mutex,
) chan struct{} {
	collectDone := make(chan struct{})

	go func() {
		defer close(collectDone)

		collectStreamOutput(stdout, stderr, receivedLines, streamMu)
	}()

	return collectDone
}

func collectStreamOutput(
	stdout, stderr chan string,
	receivedLines *[]string,
	streamMu *sync.Mutex,
) {
	for {
		select {
		case line, ok := <-stdout:
			if !ok {
				drainStderr(stderr, receivedLines, streamMu)

				return
			}

			appendLine(receivedLines, streamMu, line)
		case line, ok := <-stderr:
			if !ok {
				continue // keep collecting from stdout
			}

			appendLine(receivedLines, streamMu, "STDERR: "+line)
		}
	}
}

func drainStderr(
	stderr chan string,
	receivedLines *[]string,
	streamMu *sync.Mutex,
) {
	for {
		select {
		case line, ok := <-stderr:
			if !ok {
				return // both channels closed
			}

			appendLine(receivedLines, streamMu, "STDERR: "+line)
		case <-time.After(100 * time.Millisecond):
			return // timeout waiting for stderr
		}
	}
}

func appendLine(
	receivedLines *[]string,
	streamMu *sync.Mutex,
	line string,
) {
	streamMu.Lock()

	*receivedLines = append(*receivedLines, line)

	streamMu.Unlock()
}

func runStreamingTest(
	ctx context.Context,
	rpitx *RPITX,
	stdout, stderr chan string,
	collectDone, execDone chan struct{},
) {
	rpitx.StreamOutputs(stdout, stderr)

	// Let the mock shell command run for a bit to generate output
	time.Sleep(2500 * time.Millisecond)

	// Stop execution
	_ = rpitx.Stop(ctx)

	// Wait for execution and collection to complete
	waitForCompletion(execDone, collectDone)
}

func waitForCompletion(execDone, collectDone chan struct{}) {
	select {
	case <-execDone:
	case <-time.After(5 * time.Second):
		// Execution didn't complete in time, continue anyway
	}

	select {
	case <-collectDone:
	case <-time.After(2 * time.Second):
		// Collection didn't complete in time, continue anyway
	}
}

func verifyStreamingResults(
	t *testing.T,
	execErr *error,
	rpitx *RPITX,
	receivedLines *[]string,
	streamMu *sync.Mutex,
) {
	t.Helper()
	// Verify execution completed (termination signals are expected)
	if *execErr != nil &&
		!errors.Is(*execErr, commonerrors.ErrTerminated) &&
		!errors.Is(*execErr, commonerrors.ErrKilled) {
		t.Errorf("unexpected exec error: %v", *execErr)
	}

	assert.False(t, rpitx.isExecuting.Load(), "should not be executing after completion")

	// Check collected output
	streamMu.Lock()

	totalLines := len(*receivedLines)
	receivedLinesCopy := make([]string, totalLines)
	copy(receivedLinesCopy, *receivedLines)
	streamMu.Unlock()

	t.Logf("Received %d lines during streaming", totalLines)

	for i, line := range receivedLinesCopy {
		t.Logf("Line %d: %s", i+1, line)
	}

	// Verify we got expected output
	assert.True(t, totalLines > 0, "should have received some output lines from the mock execution")

	expectedPattern := "mocking execution of pifmrds"
	for _, line := range receivedLinesCopy {
		assert.Contains(t, line, expectedPattern,
			"received line should contain expected mock content: %s", line)
	}
}
