package gorpitx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptExists(t *testing.T) {
	// Test with existing file
	tempFile := "/tmp/test_script_exists.sh"
	err := os.WriteFile(tempFile, []byte("test"), 0o600)
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile) }()

	assert.True(t, scriptExists(tempFile))

	// Test with non-existing file
	assert.False(t, scriptExists("/tmp/nonexistent_file.sh"))
}

func TestEnsureAudioSockPresets(t *testing.T) {
	tests := []struct {
		name       string
		moduleName ModuleName
		setupFunc  func()
		expectErr  bool
	}{
		{
			name:       "non-audiosock module",
			moduleName: ModuleNameFSK,
			setupFunc:  func() {},
			expectErr:  false,
		},
		{
			name:       "audiosock module with existing presets",
			moduleName: ModuleNameAudioSockBroadcast,
			setupFunc: func() {
				_ = os.WriteFile(csdrPresetsPath, []byte("test"), 0o600)
			},
			expectErr: false,
		},
		{
			name:       "audiosock module without presets",
			moduleName: ModuleNameAudioSockBroadcast,
			setupFunc: func() {
				_ = os.Remove(csdrPresetsPath)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			defer func() { _ = os.Remove(csdrPresetsPath) }()

			err := ensureAudioSockPresets(tt.moduleName)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetScriptContent(t *testing.T) {
	tests := []struct {
		name       string
		moduleName ModuleName
		expectErr  bool
	}{
		{
			name:       "FSK module",
			moduleName: ModuleNameFSK,
			expectErr:  false,
		},
		{
			name:       "AudioSock module",
			moduleName: ModuleNameAudioSockBroadcast,
			expectErr:  false,
		},
		{
			name:       "unknown module",
			moduleName: ModuleName("unknown"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := getScriptContent(tt.moduleName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

func TestCreateScriptDir(t *testing.T) {
	tempDir := "/tmp/test_script_dir"
	testPath := filepath.Join(tempDir, "subdir", "script.sh")

	// Clean up
	defer func() { _ = os.RemoveAll(tempDir) }()

	err := createScriptDir(testPath)
	assert.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(filepath.Dir(testPath))
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestWriteScriptFile(t *testing.T) {
	tempFile := "/tmp/test_write_script.sh"

	defer func() { _ = os.Remove(tempFile) }()

	content := "#!/bin/bash\necho 'test'\n"
	err := writeScriptFile(tempFile, content)
	assert.NoError(t, err)

	// Verify file was written with correct content
	written, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, content, string(written))
}

func TestMakeExecutable(t *testing.T) {
	tempFile := "/tmp/test_make_executable.sh"

	// Create file with non-executable permissions
	err := os.WriteFile(tempFile, []byte("test"), 0o600)
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile) }()

	err = makeExecutable(tempFile)
	assert.NoError(t, err)

	// Verify file is now executable
	info, err := os.Stat(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(execPerm), info.Mode().Perm())
}

func TestWriteScript(t *testing.T) {
	tempDir := "/tmp/test_write_script_dir"
	testPath := filepath.Join(tempDir, "test_script.sh")

	// Clean up
	defer func() { _ = os.RemoveAll(tempDir) }()

	err := writeScript(ModuleNameFSK, testPath)
	assert.NoError(t, err)

	// Verify script was written and is executable
	info, err := os.Stat(testPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(execPerm), info.Mode().Perm())

	// Verify content
	content, err := os.ReadFile(testPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "#!/bin/bash")
}
