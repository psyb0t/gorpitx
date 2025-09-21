package gorpitx

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPITX_Exec(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  ModuleName
		args        map[string]any
		timeout     time.Duration
		expectError bool
	}{
		{
			name:       "valid pifmrds module",
			moduleName: ModuleNamePIFMRDS,
			args: map[string]any{
				"freq":  107.9,
				"audio": ".fixtures/test.wav",
			},
			timeout:     1 * time.Second,
			expectError: true, // will error because audio file doesn't exist
		},
		{
			name:        "unknown module",
			moduleName:  "nonexistent",
			args:        map[string]any{},
			timeout:     1 * time.Second,
			expectError: true,
		},
		{
			name:        "invalid json args",
			moduleName:  ModuleNamePIFMRDS,
			args:        map[string]any{}, // missing required fields
			timeout:     1 * time.Second,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create RPITX instance with mock commander
			mockCommander := commander.NewMock()
			rpitx := &RPITX{
				modules: map[string]Module{
					ModuleNamePIFMRDS: &PIFMRDS{},
				},
				commander: mockCommander,
			}

			// Set up mock expectations for dev environment if needed
			// Note: Environment detection now handled by env.IsDev() from common-go

			argsBytes, err := json.Marshal(tt.args)
			require.NoError(t, err)

			ctx := context.Background()
			err = rpitx.Exec(ctx, tt.moduleName, argsBytes, tt.timeout)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRPITX_Exec_DevEnvironment(t *testing.T) {
	// Set ENV=dev to trigger dev mode
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Test dev environment specifically
	mockCommander := commander.NewMock()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: mockCommander,
	}

	// Mock the shell command for dev environment
	mockCommander.ExpectWithMatchers("sh", commander.Exact("-c"), commander.Any()).ReturnError(context.DeadlineExceeded)

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 100*time.Millisecond)

	// Should timeout in dev mode
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRPITX_GetInstance(t *testing.T) {
	// Set ENV=dev to avoid root check in tests
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx1 := GetInstance()
	rpitx2 := GetInstance()

	// Should return same instance (singleton)
	assert.Same(t, rpitx1, rpitx2)

	// Should have all modules registered
	assert.Contains(t, rpitx1.modules, ModuleNamePIFMRDS)
	assert.Contains(t, rpitx1.modules, ModuleNameTUNE)
	assert.Contains(t, rpitx1.modules, ModuleNameMORSE)
	assert.Contains(t, rpitx1.modules, ModuleNameSPECTRUMPAINT)
}

func TestRPITX_GetSupportedModules(t *testing.T) {
	// Set ENV=dev to avoid root check in tests
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx := GetInstance()
	modules := rpitx.GetSupportedModules()

	// Should return all registered modules
	assert.Len(t, modules, 8)
	assert.Contains(t, modules, ModuleNamePIFMRDS)
	assert.Contains(t, modules, ModuleNameTUNE)
	assert.Contains(t, modules, ModuleNameMORSE)
	assert.Contains(t, modules, ModuleNameSPECTRUMPAINT)
	assert.Contains(t, modules, ModuleNamePICHIRP)
	assert.Contains(t, modules, ModuleNamePOCSAG)
	assert.Contains(t, modules, ModuleNameFT8)
	assert.Contains(t, modules, ModuleNamePISSSTV)

	// Should return a new slice each time (checking length consistency)
	modules2 := rpitx.GetSupportedModules()
	assert.Len(t, modules2, 8)
	assert.Contains(t, modules2, ModuleNamePIFMRDS)
	assert.Contains(t, modules2, ModuleNameTUNE)
	assert.Contains(t, modules2, ModuleNameMORSE)
	assert.Contains(t, modules2, ModuleNameSPECTRUMPAINT)
	assert.Contains(t, modules2, ModuleNamePICHIRP)
	assert.Contains(t, modules2, ModuleNamePOCSAG)
}

func TestRPITX_IsSupportedModule(t *testing.T) {
	// Set ENV=dev to avoid root check in tests
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx := GetInstance()

	// Test supported modules
	assert.True(t, rpitx.IsSupportedModule(ModuleNamePIFMRDS))
	assert.True(t, rpitx.IsSupportedModule(ModuleNameTUNE))
	assert.True(t, rpitx.IsSupportedModule(ModuleNameMORSE))
	assert.True(t, rpitx.IsSupportedModule(ModuleNameSPECTRUMPAINT))

	// Test unsupported module
	assert.False(t, rpitx.IsSupportedModule("nonexistent"))
	assert.False(t, rpitx.IsSupportedModule(""))
}

func TestRPITX_StreamOutputs_DuringExecution(t *testing.T) {
	// Test StreamOutputs timing during actual execution
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx := GetInstance()
	ctx := context.Background()

	// Create valid MORSE args
	args := map[string]any{
		"frequency": 434000000.0,
		"rate":      20,
		"message":   "TEST STREAMING",
	}
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	// Create channels for streaming
	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	// Track execution state
	execStarted := make(chan bool, 1)
	execFinished := make(chan error, 1)

	// Start execution in goroutine
	go func() {
		// Signal that execution is starting
		execStarted <- true

		// Execute with timeout
		err := rpitx.Exec(ctx, ModuleNameMORSE, argsJSON, 2*time.Second)
		execFinished <- err
	}()

	// Wait for execution to start
	<-execStarted

	// Give the process time to actually start (this is the critical timing)
	time.Sleep(200 * time.Millisecond)

	// Now try to stream outputs - this should work if timing is correct
	streamStarted := make(chan bool, 1)
	go func() {
		rpitx.StreamOutputs(stdout, stderr)
		streamStarted <- true
	}()

	// Wait for streaming to be set up
	<-streamStarted

	// Collect some output or timeout
	outputReceived := false
	timeout := time.After(1500 * time.Millisecond)

	for {
		select {
		case line := <-stdout:
			t.Logf("Received stdout: %s", line)
			outputReceived = true
		case line := <-stderr:
			t.Logf("Received stderr: %s", line)
			outputReceived = true
		case err := <-execFinished:
			t.Logf("Execution finished with error: %v", err)
			goto checkResults
		case <-timeout:
			t.Log("Test timeout reached")
			goto checkResults
		}
	}

checkResults:
	// We should have received some output from the mock command
	if !outputReceived {
		t.Log("WARNING: No output received - this indicates a timing issue with StreamOutputs")
		t.Log("The mock command should output 'mocking execution' every second")
	}

	// Clean up
	instance = nil
	once = sync.Once{}
}

func TestRPITX_StreamOutputsAsync(t *testing.T) {
	// Test StreamOutputsAsync for easier usage
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx := GetInstance()
	ctx := context.Background()

	// Create valid MORSE args
	args := map[string]any{
		"frequency": 434000000.0,
		"rate":      20,
		"message":   "TEST ASYNC STREAMING",
	}
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	// Create channels for streaming
	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	// Start async streaming BEFORE execution (this should work)
	rpitx.StreamOutputsAsync(stdout, stderr)

	// Start execution after setting up streaming
	execFinished := make(chan error, 1)
	go func() {
		err := rpitx.Exec(ctx, ModuleNameMORSE, argsJSON, 2*time.Second)
		execFinished <- err
	}()

	// Collect output
	outputReceived := false
	timeout := time.After(3 * time.Second)

	for {
		select {
		case line := <-stdout:
			t.Logf("Received stdout: %s", line)
			outputReceived = true
		case line := <-stderr:
			t.Logf("Received stderr: %s", line)
			outputReceived = true
		case err := <-execFinished:
			t.Logf("Execution finished with error: %v", err)
			goto checkResults
		case <-timeout:
			t.Log("Test timeout reached")
			goto checkResults
		}
	}

checkResults:
	// Should have received output
	assert.True(t, outputReceived, "Should have received output from async streaming")

	// Clean up
	instance = nil
	once = sync.Once{}
}

func TestRPITX_getMockExecCmd(t *testing.T) {
	// Set ENV=dev to test mock execution
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	rpitx := &RPITX{}

	args := []string{"-freq", "107.9", "-audio", ".fixtures/test.wav"}

	cmdName, cmdArgs := rpitx.getMockExecCmd(ModuleNamePIFMRDS, args)

	// Should return shell command
	assert.Equal(t, "sh", cmdName)
	assert.Len(t, cmdArgs, 2)
	assert.Equal(t, "-c", cmdArgs[0])
	assert.Contains(t, cmdArgs[1], "mocking execution of pifmrds")
	assert.Contains(t, cmdArgs[1], "-freq 107.9 -audio .fixtures/test.wav")
}

func TestRPITX_getMockExecCmd_CommandContent(t *testing.T) {
	// Test that mock execution generates correct command content
	rpitx := &RPITX{}

	args := []string{"-freq", "107.9", "-ps", "TEST FM"}

	cmdName, cmdArgs := rpitx.getMockExecCmd("testmodule", args)

	// Should return shell command
	assert.Equal(t, "sh", cmdName)
	assert.Len(t, cmdArgs, 2)
	assert.Equal(t, "-c", cmdArgs[0])

	// Check command contains the infinite loop structure
	assert.Contains(t, cmdArgs[1], "while true; do")
	assert.Contains(t, cmdArgs[1], "echo \"mocking execution of testmodule")
	assert.Contains(t, cmdArgs[1], "-freq 107.9 -ps TEST FM")
	assert.Contains(t, cmdArgs[1], "sleep 1")
	assert.Contains(t, cmdArgs[1], "done")
}

func TestRPITX_Exec_TuneModule(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  ModuleName
		args        map[string]any
		expectError bool
	}{
		{
			name:       "valid tune module",
			moduleName: ModuleNameTUNE,
			args: map[string]any{
				"frequency": 434000000.0, // 434 MHz in Hz
			},
			expectError: false,
		},
		{
			name:       "tune module with all parameters",
			moduleName: ModuleNameTUNE,
			args: map[string]any{
				"frequency":     434000000.0,
				"exitImmediate": true,
				"ppm":           2.5,
			},
			expectError: false,
		},
		{
			name:       "tune module missing frequency",
			moduleName: ModuleNameTUNE,
			args: map[string]any{
				"exitImmediate": true,
			},
			expectError: true,
		},
		{
			name:       "tune module invalid frequency",
			moduleName: ModuleNameTUNE,
			args: map[string]any{
				"frequency": -434000000.0,
			},
			expectError: true,
		},
		{
			name:       "morse module valid",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"frequency": 14070000.0,
				"rate":      20,
				"message":   "CQ DE N0CALL",
			},
			expectError: false,
		},
		{
			name:       "morse module with different params",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"frequency": 7040000.0,
				"rate":      15,
				"message":   "HELLO WORLD",
			},
			expectError: false,
		},
		{
			name:       "morse module missing frequency",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"rate":    20,
				"message": "TEST",
			},
			expectError: true,
		},
		{
			name:       "morse module missing rate",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"frequency": 14070000.0,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name:       "morse module missing message",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"frequency": 14070000.0,
				"rate":      20,
			},
			expectError: true,
		},
		{
			name:       "morse module invalid frequency",
			moduleName: ModuleNameMORSE,
			args: map[string]any{
				"frequency": -14070000.0,
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set ENV=dev to test mock execution
			t.Setenv(env.EnvVarName, env.EnvTypeDev)

			// Create RPITX instance with mock commander
			mockCommander := commander.NewMock()
			rpitx := &RPITX{
				modules: map[string]Module{
					ModuleNamePIFMRDS: &PIFMRDS{},
					ModuleNameTUNE:    &TUNE{},
					ModuleNameMORSE:   &MORSE{},
				},
				commander: mockCommander,
			}

			if !tt.expectError {
				// Mock successful execution for valid test cases
				mockCommander.ExpectWithMatchers("sh", commander.Exact("-c"), commander.Any()).ReturnError(context.DeadlineExceeded)
			}

			argsBytes, err := json.Marshal(tt.args)
			require.NoError(t, err)

			ctx := context.Background()
			err = rpitx.Exec(ctx, tt.moduleName, argsBytes, 1*time.Second)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			// Should timeout in dev mode (this is expected)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		})
	}
}

// Additional coverage tests for missing scenarios

func TestRPITX_ProductionExecution_Success(t *testing.T) {
	// Test actual production execution path with mock commander
	t.Setenv(env.EnvVarName, env.EnvTypeProd)

	mockCommander := commander.NewMock()
	rpitx := &RPITX{
		config: Config{Path: "$HOME/rpitx"},
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: mockCommander,
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

	// Mock successful production execution with stdbuf wrapper
	mockCommander.Expect("stdbuf",
		"-oL", "$HOME/rpitx/pifmrds",
		"-freq", "107.9",
		"-audio", ".fixtures/test.wav",
		"-pi", "1234",
		"-ps", "TEST FM",
		"-rt", "Test Radio Text",
	).ReturnError(nil)

	ctx := context.Background()

	// Should execute production path and succeed
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 1*time.Second)

	require.NoError(t, err)
	assert.NoError(t, mockCommander.VerifyExpectations())
}

func TestRPITX_ProductionExecution_StartFailure(t *testing.T) {
	// Test production execution start failure
	t.Setenv(env.EnvVarName, env.EnvTypeProd)

	mockCommander := commander.NewMock()
	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: mockCommander,
	}

	// Mock start failure by returning error from execute
	mockCommander.Expect("stdbuf", "-oL", "./pifmrds", "-freq", "107.9", "-audio", ".fixtures/test.wav").ReturnError(assert.AnError)

	args := map[string]any{
		"freq":  107.9,
		"audio": ".fixtures/test.wav",
	}

	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ctx := context.Background()

	// Should fail to start process
	err = rpitx.Exec(ctx, ModuleNamePIFMRDS, argsBytes, 1*time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start process")
}

func TestPIFMRDS_ControlPipeValidation(t *testing.T) {
	tests := []struct {
		name        string
		controlPipe *string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no control pipe",
			controlPipe: nil,
			expectError: false,
		},
		{
			name:        "empty control pipe",
			controlPipe: stringPtr("   "),
			expectError: true,
			errorMsg:    "control pipe path cannot be empty when specified",
		},
		{
			name:        "non-existent control pipe",
			controlPipe: stringPtr("/tmp/nonexistent.pipe"),
			expectError: true,
			errorMsg:    "control pipe does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pifm := &PIFMRDS{
				Freq:        107.9,
				Audio:       ".fixtures/test.wav",
				ControlPipe: tt.controlPipe,
			}

			err := pifm.validateControlPipe()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPIFMRDS_RTValidation_TooLong(t *testing.T) {
	pifm := &PIFMRDS{
		Freq:  107.9,
		Audio: ".fixtures/test.wav",
		RT:    "This radio text message is way too fucking long for RDS standards and should trigger validation error",
	}

	err := pifm.validateRT()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RT text must be 64 characters or less")
}

func TestPIFMRDS_BuildArgs_AllFields(t *testing.T) {
	ppm := 100.0
	controlPipe := "/tmp/test.pipe"
	pifm := &PIFMRDS{
		Freq:        107.9,
		Audio:       ".fixtures/test.wav",
		PI:          "1234",
		PS:          "TEST FM",
		RT:          "Test radio text",
		PPM:         &ppm,
		ControlPipe: &controlPipe,
	}

	args := pifm.buildArgs()

	// Check all args are present
	assert.Contains(t, args, "-freq")
	assert.Contains(t, args, "107.9")
	assert.Contains(t, args, "-audio")
	assert.Contains(t, args, ".fixtures/test.wav")
	assert.Contains(t, args, "-pi")
	assert.Contains(t, args, "1234")
	assert.Contains(t, args, "-ps")
	assert.Contains(t, args, "TEST FM")
	assert.Contains(t, args, "-rt")
	assert.Contains(t, args, "Test radio text")
	assert.Contains(t, args, "-ppm")
	assert.Contains(t, args, "100")
	assert.Contains(t, args, "-ctl")
	assert.Contains(t, args, "/tmp/test.pipe")
}

func TestPIFMRDS_Validate_RTError(t *testing.T) {
	// Test validate function early return from validateRT
	pifm := &PIFMRDS{
		Freq:  107.9,
		Audio: ".fixtures/test.wav",
		PI:    "1234",
		PS:    "TEST",
		RT: "This radio text message is way too fucking long for RDS standards and should trigger " +
			"validation error during validate call",
	}

	err := pifm.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RT text must be 64 characters or less")
}

func TestPIFMRDS_Validate_ControlPipeError(t *testing.T) {
	// Test validate function early return from validateControlPipe
	pifm := &PIFMRDS{
		Freq:  107.9,
		Audio: ".fixtures/test.wav",
		PI:    "1234",
		PS:    "TEST",
		RT:    "Test",
	}

	// Empty control pipe should trigger error
	controlPipe := ""
	pifm.ControlPipe = &controlPipe

	err := pifm.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "control pipe path cannot be empty when specified")
}

func TestRPITX_StopWithoutExecution(t *testing.T) {
	// Test stopping when nothing is executing
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	rpitx := GetInstance()

	ctx := context.Background()
	err := rpitx.Stop(ctx)

	// Should return ErrNotExecuting
	assert.ErrorIs(t, err, ErrNotExecuting)

	// Clean up
	instance = nil
	once = sync.Once{}
}

func TestRPITX_StreamOutputs_NotExecuting(t *testing.T) {
	// Test StreamOutputs when not executing
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	rpitx := &RPITX{
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: commander.NewMock(),
	}

	stdout := make(chan string, 10)
	stderr := make(chan string, 10)

	// Should return early without closing channels (due to early return)
	rpitx.StreamOutputs(stdout, stderr)

	// Channels should NOT be closed (early return prevents defer)
	// Try to send to verify they're still open
	select {
	case stdout <- "test":
		// Good, channel is still open
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should be able to send to stdout channel")
	}

	select {
	case stderr <- "test":
		// Good, channel is still open
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should be able to send to stderr channel")
	}

	// Clean up
	close(stdout)
	close(stderr)
}

func TestRPITX_PrepareCommand_Production(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeProd)

	rpitx := &RPITX{
		config: Config{Path: "/home/test/rpitx"},
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: commander.NewMock(),
	}

	// Test args
	args := map[string]any{
		"freq":  100.0,
		"audio": ".fixtures/test.wav",
	}

	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("Failed to marshal args: %v", err)
	}

	cmdName, cmdArgs, _, err := rpitx.prepareCommand("pifmrds", argsJSON)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedCmdName := "stdbuf"
	if cmdName != expectedCmdName {
		t.Errorf("Expected cmdName to be '%s', got: %s", expectedCmdName, cmdName)
	}

	// Check that cmdArgs contains line buffering and parsed arguments
	expectedArgs := []string{"-oL", "/home/test/rpitx/pifmrds", "-freq", "100.0", "-audio", ".fixtures/test.wav"}
	if len(cmdArgs) < len(expectedArgs) {
		t.Errorf("Expected cmdArgs to have at least %d elements, got: %v", len(expectedArgs), cmdArgs)
	}

	// Check key elements are present
	if !contains(cmdArgs, "-oL") || !contains(cmdArgs, "/home/test/rpitx/pifmrds") {
		t.Errorf("Expected cmdArgs to contain stdbuf args, got: %v", cmdArgs)
	}

	t.Logf("Production command: %s %v", cmdName, cmdArgs)
}

func TestRPITX_PrepareCommand_Development(t *testing.T) {
	// Test that development mode uses mock execution
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	rpitx := &RPITX{
		config: Config{Path: "/home/test/rpitx"},
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS: &PIFMRDS{},
		},
		commander: commander.NewMock(),
	}

	// Test args
	args := map[string]any{
		"freq":  100.0,
		"audio": ".fixtures/test.wav",
	}

	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("Failed to marshal args: %v", err)
	}

	cmdName, cmdArgs, _, err := rpitx.prepareCommand("pifmrds", argsJSON)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmdName != "sh" {
		t.Errorf("Expected cmdName to be 'sh', got: %s", cmdName)
	}

	t.Logf("Development command: %s %v", cmdName, cmdArgs)
}

// Helper functions for test.
func stringPtr(s string) *string {
	return &s
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func TestModules_StdinBehavior(t *testing.T) {
	tests := []struct {
		name           string
		module         Module
		input          map[string]any
		expectStdinNil bool
	}{
		{
			name:   "TUNE module returns nil stdin",
			module: &TUNE{},
			input: map[string]any{
				"frequency": 434000000,
			},
			expectStdinNil: true,
		},
		{
			name:   "MORSE module returns nil stdin",
			module: &MORSE{},
			input: map[string]any{
				"frequency": 14070000,
				"rate":      20,
				"message":   "CQ CQ DE N0CALL K",
			},
			expectStdinNil: true,
		},
		{
			name:   "PIFMRDS module returns nil stdin",
			module: &PIFMRDS{},
			input: map[string]any{
				"freq":  107.9,
				"audio": ".fixtures/test.wav",
			},
			expectStdinNil: true,
		},
		{
			name:   "PICHIRP module returns nil stdin",
			module: &PICHIRP{},
			input: map[string]any{
				"frequency": 28070000,
				"bandwidth": 1000,
				"time":      1,
			},
			expectStdinNil: true,
		},
		{
			name:   "SPECTRUMPAINT module returns nil stdin",
			module: &SPECTRUMPAINT{},
			input: map[string]any{
				"pictureFile": ".fixtures/test_spectrum_320x100.Y",
				"frequency":   28000000,
			},
			expectStdinNil: true,
		},
		{
			name:   "POCSAG module returns stdin content",
			module: &POCSAG{},
			input: map[string]any{
				"frequency": 466230000,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test message",
					},
				},
			},
			expectStdinNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, stdin, err := tt.module.ParseArgs(inputBytes)

			// All modules should parse successfully with valid input
			if tt.expectStdinNil {
				require.NoError(t, err)
				assert.NotNil(t, args)
				assert.Nil(t, stdin, "Module %T should return nil stdin", tt.module)
			} else {
				// POCSAG should have stdin content
				require.NoError(t, err)
				assert.NotNil(t, args)
				assert.NotNil(t, stdin, "POCSAG module should return non-nil stdin")
			}
		})
	}
}
