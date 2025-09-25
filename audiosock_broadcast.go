package gorpitx

import (
	"encoding/json"
	"io"
	"slices"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameAudioSockBroadcast ModuleName = "audiosock-broadcast"
)

type CSDRPresetType = string

const (
	CSDRPresetAM  CSDRPresetType = "AM"
	CSDRPresetDSB CSDRPresetType = "DSB"
	CSDRPresetUSB CSDRPresetType = "USB"
	CSDRPresetLSB CSDRPresetType = "LSB"
	CSDRPresetFM  CSDRPresetType = "FM"
	CSDRPresetRAW CSDRPresetType = "RAW"
)

const (
	defaultAudioSockBroadcastSampleRate = 48000
)

type AudioSockBroadcast struct {
	// SocketPath specifies the Unix socket path for audio input. Required.
	SocketPath string `json:"socketPath"`

	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// SampleRate specifies the audio sample rate. Optional parameter.
	// Default: 48000 Hz
	SampleRate *int `json:"sampleRate,omitempty"`

	// CSDRPreset specifies the CSDR processing preset mode. Optional parameter.
	// If not specified, uses default "FM".
	// Available: AM, DSB, USB, LSB, FM, RAW
	CSDRPreset *string `json:"csdrPreset,omitempty"`

	// Gain specifies the gain multiplier for the audio signal. Optional parameter.
	// Default: 1.0
	Gain *float64 `json:"gain,omitempty"`
}

func (m *AudioSockBroadcast) ParseArgs(
	args json.RawMessage,
) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for
// AudioSock script.
func (m *AudioSockBroadcast) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add socket path argument (required)
	args = append(args, m.SocketPath)

	// Add sample rate argument (default if not specified)
	sampleRate := defaultAudioSockBroadcastSampleRate
	if m.SampleRate != nil {
		sampleRate = *m.SampleRate
	}

	args = append(args, strconv.Itoa(sampleRate))

	// Add CSDR preset argument (default if not specified)
	csdrPreset := CSDRPresetFM
	if m.CSDRPreset != nil {
		csdrPreset = *m.CSDRPreset
	}

	args = append(args, csdrPreset)

	// Add gain argument (default if not specified)
	gain := 1.0
	if m.Gain != nil {
		gain = *m.Gain
	}

	args = append(args, strconv.FormatFloat(gain, 'f', -1, 64))

	return args
}

// validate validates all AudioSock parameters.
func (m *AudioSockBroadcast) validate() error {
	if err := m.validateSocketPath(); err != nil {
		return err
	}

	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateSampleRate(); err != nil {
		return err
	}

	if err := m.validateCSDRPreset(); err != nil {
		return err
	}

	if err := m.validateGain(); err != nil {
		return err
	}

	return nil
}

// validateSocketPath validates the socket path parameter.
func (m *AudioSockBroadcast) validateSocketPath() error {
	if m.SocketPath == "" {
		return ctxerrors.Wrap(
			commonerrors.ErrRequiredFieldNotSet, "socketPath")
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *AudioSockBroadcast) validateFrequency() error {
	if m.Frequency <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			m.Frequency,
		)
	}

	// Validate frequency range using Hz-based validation
	if !isValidFreqHz(m.Frequency) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f Hz",
			minFreqKHz, getMaxFreqMHzDisplay(), m.Frequency,
		)
	}

	return nil
}

// validateSampleRate validates the sample rate parameter.
func (m *AudioSockBroadcast) validateSampleRate() error {
	if m.SampleRate != nil && *m.SampleRate <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"sample rate must be positive, got: %d",
			*m.SampleRate,
		)
	}

	return nil
}

// validateCSDRPreset validates the CSDR preset parameter.
func (m *AudioSockBroadcast) validateCSDRPreset() error {
	if m.CSDRPreset == nil {
		return nil // Optional parameter
	}

	validPresets := []CSDRPresetType{
		CSDRPresetAM,
		CSDRPresetDSB,
		CSDRPresetUSB,
		CSDRPresetLSB,
		CSDRPresetFM,
		CSDRPresetRAW,
	}

	preset := *m.CSDRPreset
	if slices.Contains(validPresets, preset) {
		return nil
	}

	return ctxerrors.Wrapf(
		commonerrors.ErrInvalidValue,
		"invalid CSDR preset: %s, valid presets: %v",
		preset, validPresets,
	)
}

// validateGain validates the gain parameter.
func (m *AudioSockBroadcast) validateGain() error {
	if m.Gain != nil && *m.Gain < 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"gain must be non-negative, got: %f",
			*m.Gain,
		)
	}

	return nil
}
