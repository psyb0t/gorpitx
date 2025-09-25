package gorpitx

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/psyb0t/ctxerrors"
)

const (
	fskScriptPath          = "/tmp/fsk.sh"
	audioSockBroadcastPath = "/tmp/audiosock_broadcast.sh"
	csdrPresetsPath        = "/tmp/csdr_presets.sh"

	dirPerm    = 0o750
	scriptPerm = 0o600
	execPerm   = 0o700
)

// fskScript contains the embedded FSK script content
//
//go:embed scripts/fsk.sh
var fskScript string

// audioSockBroadcastScript contains the embedded AudioSock script
//
//go:embed scripts/audiosock_broadcast.sh
var audioSockBroadcastScript string

// csdrPresetsScript contains the embedded CSDR presets script
//
//go:embed scripts/csdr_presets.sh
var csdrPresetsScript string

// ModuleNameToScriptName returns the script path for script-based modules.
func ModuleNameToScriptName(moduleName ModuleName) (string, bool) {
	switch moduleName {
	case ModuleNameFSK:
		return fskScriptPath, true
	case ModuleNameAudioSockBroadcast:
		return audioSockBroadcastPath, true
	default:
		return "", false
	}
}

// EnsureScriptExists writes the embedded script to filesystem if it doesn't
// exist.
func EnsureScriptExists(moduleName ModuleName) error {
	scriptPath, isScript := ModuleNameToScriptName(moduleName)
	if !isScript {
		return nil // Not a script-based module
	}

	// Always overwrite script (no existence check)

	var scriptContent string

	switch moduleName {
	case ModuleNameFSK:
		scriptContent = fskScript
	case ModuleNameAudioSockBroadcast:
		scriptContent = audioSockBroadcastScript
	default:
		return ctxerrors.Wrapf(
			ErrUnknownModule,
			"no script content for module: %s",
			moduleName,
		)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(scriptPath), dirPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to create script directory: %s",
			filepath.Dir(scriptPath),
		)
	}

	// Write script to filesystem
	if err := os.WriteFile(
		scriptPath, []byte(scriptContent), scriptPerm,
	); err != nil {
		return ctxerrors.Wrapf(err, "failed to write script: %s", scriptPath)
	}

	// Make script executable
	if err := os.Chmod(scriptPath, execPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to make script executable: %s",
			scriptPath,
		)
	}

	// Also write csdr_presets.sh for audiosock module
	if moduleName == ModuleNameAudioSockBroadcast {
		if err := ensureCSRDPresetsScript(scriptPerm, execPerm); err != nil {
			return err
		}
	}

	return nil
}

// ensureCSRDPresetsScript writes the csdr_presets.sh script to filesystem.
func ensureCSRDPresetsScript(scriptPerm, execPerm os.FileMode) error {
	if err := os.WriteFile(
		csdrPresetsPath, []byte(csdrPresetsScript), scriptPerm,
	); err != nil {
		return ctxerrors.Wrapf(err,
			"failed to write csdr_presets.sh: %s", csdrPresetsPath)
	}

	// Make csdr_presets.sh executable
	if err := os.Chmod(csdrPresetsPath, execPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to make csdr_presets.sh executable: %s",
			csdrPresetsPath,
		)
	}

	return nil
}

// IsScriptModule returns true if the module uses an embedded script.
func IsScriptModule(moduleName ModuleName) bool {
	_, isScript := ModuleNameToScriptName(moduleName)

	return isScript
}
