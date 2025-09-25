package gorpitx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudioSockBroadcast_ParseArgs_Success(t *testing.T) {
	tests := []struct {
		name         string
		input        AudioSockBroadcast
		expectedArgs []string
	}{
		{
			name: "basic configuration with default sample rate",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
			},
			expectedArgs: []string{
				"144500000", "/tmp/audio_socket", "48000", "FM", "1",
			},
		},
		{
			name: "custom sample rate",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/custom_socket",
				Frequency:  434000000.0,
				SampleRate: intPtr(96000),
			},
			expectedArgs: []string{
				"434000000", "/tmp/custom_socket", "96000", "FM", "1",
			},
		},
		{
			name: "high frequency configuration",
			input: AudioSockBroadcast{
				SocketPath: "/var/tmp/voice_socket",
				Frequency:  1296000000.0,
				SampleRate: intPtr(22050),
			},
			expectedArgs: []string{
				"1296000000", "/var/tmp/voice_socket", "22050", "FM", "1",
			},
		},
		{
			name: "custom modulation",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				Modulation: stringPtr("FM"),
			},
			expectedArgs: []string{
				"144500000", "/tmp/audio_socket", "48000", "FM", "1",
			},
		},
		{
			name: "custom gain",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				Gain:       floatPtr(2.5),
			},
			expectedArgs: []string{
				"144500000", "/tmp/audio_socket", "48000", "FM", "2.5",
			},
		},
		{
			name: "all custom parameters",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/custom_socket",
				Frequency:  434000000.0,
				SampleRate: intPtr(96000),
				Modulation: stringPtr("USB"),
				Gain:       floatPtr(3.0),
			},
			expectedArgs: []string{
				"434000000", "/tmp/custom_socket", "96000", "USB", "3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, stdin, err := tt.input.ParseArgs(inputBytes)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedArgs, args)
			assert.Nil(t, stdin) // USB AudioSock doesn't use stdin
		})
	}
}

func TestAudioSockBroadcast_ParseArgs_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         AudioSockBroadcast
		expectedError string
	}{
		{
			name: "missing socket path",
			input: AudioSockBroadcast{
				Frequency: 144500000.0,
			},
			expectedError: "socketPath",
		},
		{
			name: "empty socket path",
			input: AudioSockBroadcast{
				SocketPath: "",
				Frequency:  144500000.0,
			},
			expectedError: "socketPath",
		},
		{
			name: "missing frequency",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "negative frequency",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  -144500000.0,
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "frequency too low",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  1000.0, // 1 kHz - below 5 kHz limit
			},
			expectedError: "frequency out of RPiTX range",
		},
		{
			name: "frequency too high",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  2000000000.0, // 2 GHz - exceeds 1500 MHz limit
			},
			expectedError: "frequency out of RPiTX range",
		},
		{
			name: "negative sample rate",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				SampleRate: intPtr(-48000),
			},
			expectedError: "sample rate must be positive",
		},
		{
			name: "zero sample rate",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				SampleRate: intPtr(0),
			},
			expectedError: "sample rate must be positive",
		},
		{
			name: "invalid modulation",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				Modulation: stringPtr("INVALID"),
			},
			expectedError: "invalid modulation",
		},
		{
			name: "negative gain",
			input: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
				Gain:       floatPtr(-1.0),
			},
			expectedError: "gain must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			_, _, err = tt.input.ParseArgs(inputBytes)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestAudioSockBroadcast_ParseArgs_JSONUnmarshalError(t *testing.T) {
	usb := &AudioSockBroadcast{}
	invalidJSON := []byte(`{"frequency": "invalid"}`)

	_, _, err := usb.ParseArgs(invalidJSON)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal args")
}

func TestAudioSockBroadcast_validateSocketPath(t *testing.T) {
	tests := []struct {
		name        string
		socketPath  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid socket path",
			socketPath:  "/tmp/audio_socket",
			expectError: false,
		},
		{
			name:        "valid socket path with subdirectory",
			socketPath:  "/var/tmp/audio/voice_socket",
			expectError: false,
		},
		{
			name:        "empty socket path",
			socketPath:  "",
			expectError: true,
			errorMsg:    "socketPath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usb := &AudioSockBroadcast{SocketPath: tt.socketPath}
			err := usb.validateSocketPath()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAudioSockBroadcast_validateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, FrequencyValidationTest{
		name:        "valid frequency - 144.5 MHz",
		frequency:   144500000.0,
		expectError: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usb := &AudioSockBroadcast{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, usb.validateFrequency, tt)
		})
	}
}

func TestAudioSockBroadcast_validateSampleRate(t *testing.T) {
	tests := []struct {
		name        string
		sampleRate  *int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid sample rate",
			sampleRate:  intPtr(48000),
			expectError: false,
		},
		{
			name:        "nil sample rate (default)",
			sampleRate:  nil,
			expectError: false,
		},
		{
			name:        "high sample rate",
			sampleRate:  intPtr(96000),
			expectError: false,
		},
		{
			name:        "low sample rate",
			sampleRate:  intPtr(8000),
			expectError: false,
		},
		{
			name:        "zero sample rate",
			sampleRate:  intPtr(0),
			expectError: true,
			errorMsg:    "sample rate must be positive",
		},
		{
			name:        "negative sample rate",
			sampleRate:  intPtr(-48000),
			expectError: true,
			errorMsg:    "sample rate must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usb := &AudioSockBroadcast{SampleRate: tt.sampleRate}
			err := usb.validateSampleRate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAudioSockBroadcast_buildArgs(t *testing.T) {
	tests := []struct {
		name         string
		usb          AudioSockBroadcast
		expectedArgs []string
	}{
		{
			name: "default sample rate",
			usb: AudioSockBroadcast{
				SocketPath: "/tmp/audio_socket",
				Frequency:  144500000.0,
			},
			expectedArgs: []string{
				"144500000", "/tmp/audio_socket", "48000", "FM", "1",
			},
		},
		{
			name: "custom sample rate",
			usb: AudioSockBroadcast{
				SocketPath: "/var/tmp/voice_socket",
				Frequency:  434000000.0,
				SampleRate: intPtr(96000),
			},
			expectedArgs: []string{
				"434000000", "/var/tmp/voice_socket", "96000", "FM", "1",
			},
		},
		{
			name: "high frequency with low sample rate",
			usb: AudioSockBroadcast{
				SocketPath: "/tmp/narrowband_socket",
				Frequency:  1296000000.0,
				SampleRate: intPtr(16000),
			},
			expectedArgs: []string{
				"1296000000", "/tmp/narrowband_socket", "16000", "FM", "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.usb.buildArgs()
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestAudioSockBroadcast_validateModulation(t *testing.T) {
	tests := []struct {
		name        string
		modulation  *string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil modulation (default)",
			modulation:  nil,
			expectError: false,
		},
		{
			name:        "valid AM modulation",
			modulation:  stringPtr("AM"),
			expectError: false,
		},
		{
			name:        "valid DSB modulation",
			modulation:  stringPtr("DSB"),
			expectError: false,
		},
		{
			name:        "valid USB modulation",
			modulation:  stringPtr("USB"),
			expectError: false,
		},
		{
			name:        "valid LSB modulation",
			modulation:  stringPtr("LSB"),
			expectError: false,
		},
		{
			name:        "valid FM modulation",
			modulation:  stringPtr("FM"),
			expectError: false,
		},
		{
			name:        "invalid modulation - old WFM",
			modulation:  stringPtr("WFM"),
			expectError: true,
			errorMsg:    "invalid modulation",
		},
		{
			name:        "invalid modulation - old NFM",
			modulation:  stringPtr("NFM"),
			expectError: true,
			errorMsg:    "invalid modulation",
		},
		{
			name:        "valid RAW modulation",
			modulation:  stringPtr("RAW"),
			expectError: false,
		},
		{
			name:        "invalid modulation",
			modulation:  stringPtr("INVALID"),
			expectError: true,
			errorMsg:    "invalid modulation",
		},
		{
			name:        "case sensitive - lowercase",
			modulation:  stringPtr("fm"),
			expectError: true,
			errorMsg:    "invalid modulation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asb := &AudioSockBroadcast{Modulation: tt.modulation}
			err := asb.validateModulation()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAudioSockBroadcast_validateGain(t *testing.T) {
	tests := []struct {
		name        string
		gain        *float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil gain (default)",
			gain:        nil,
			expectError: false,
		},
		{
			name:        "valid gain - 1.0",
			gain:        floatPtr(1.0),
			expectError: false,
		},
		{
			name:        "valid gain - 2.5",
			gain:        floatPtr(2.5),
			expectError: false,
		},
		{
			name:        "valid gain - zero",
			gain:        floatPtr(0.0),
			expectError: false,
		},
		{
			name:        "valid gain - high value",
			gain:        floatPtr(10.0),
			expectError: false,
		},
		{
			name:        "invalid gain - negative",
			gain:        floatPtr(-1.0),
			expectError: true,
			errorMsg:    "gain must be non-negative",
		},
		{
			name:        "invalid gain - very negative",
			gain:        floatPtr(-5.5),
			expectError: true,
			errorMsg:    "gain must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asb := &AudioSockBroadcast{Gain: tt.gain}
			err := asb.validateGain()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
