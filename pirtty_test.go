package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestPIRTTY_ParseArgs_Success(t *testing.T) {
	tests := []struct {
		name           string
		input          PIRTTY
		expectedArgs   []string
		expectedStdin  bool
	}{
		{
			name: "basic PIRTTY parameters",
			input: PIRTTY{
				Frequency:      14070000.0, // 14.070 MHz
				SpaceFrequency: intPtr(170),
				Message:        "CQ DE N0CALL",
			},
			expectedArgs:  []string{"14070000", "170", "CQ DE N0CALL"},
			expectedStdin: false,
		},
		{
			name: "PIRTTY with different frequency",
			input: PIRTTY{
				Frequency:      28070000.0, // 28.070 MHz
				SpaceFrequency: intPtr(200),
				Message:        "RTTY TEST MESSAGE",
			},
			expectedArgs:  []string{"28070000", "200", "RTTY TEST MESSAGE"},
			expectedStdin: false,
		},
		{
			name: "PIRTTY with long message",
			input: PIRTTY{
				Frequency:      7040000.0, // 7.040 MHz
				SpaceFrequency: intPtr(85),
				Message:        "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG 1234567890",
			},
			expectedArgs:  []string{"7040000", "85", "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG 1234567890"},
			expectedStdin: false,
		},
		{
			name: "PIRTTY with default space frequency",
			input: PIRTTY{
				Frequency: 14070000.0, // 14.070 MHz
				Message:   "DEFAULT SPACE TEST",
			},
			expectedArgs:  []string{"14070000", "170", "DEFAULT SPACE TEST"},
			expectedStdin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, stdin, err := tt.input.ParseArgs(inputBytes)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedArgs, args)

			if tt.expectedStdin {
				assert.NotNil(t, stdin)
			} else {
				assert.Nil(t, stdin)
			}
		})
	}
}

func TestPIRTTY_ParseArgs_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         PIRTTY
		expectedError string
	}{
		{
			name: "missing frequency",
			input: PIRTTY{
				Message: "TEST",
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "negative frequency",
			input: PIRTTY{
				Frequency: -14070000.0,
				Message:   "TEST",
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "frequency too high",
			input: PIRTTY{
				Frequency: 2000000000.0, // 2 GHz - exceeds 1500 MHz limit
				Message:   "TEST",
			},
			expectedError: "frequency out of RPiTX range",
		},
		{
			name: "frequency too low",
			input: PIRTTY{
				Frequency: 1000.0, // 1 kHz - below 5 kHz limit
				Message:   "TEST",
			},
			expectedError: "frequency out of RPiTX range",
		},
		{
			name: "negative space frequency",
			input: PIRTTY{
				Frequency:      14070000.0,
				SpaceFrequency: intPtr(-170),
				Message:        "TEST",
			},
			expectedError: "space frequency must be positive",
		},
		{
			name: "missing message",
			input: PIRTTY{
				Frequency: 14070000.0,
			},
			expectedError: "message",
		},
		{
			name: "empty message",
			input: PIRTTY{
				Frequency: 14070000.0,
				Message:   "",
			},
			expectedError: "message",
		},
		{
			name: "whitespace only message",
			input: PIRTTY{
				Frequency: 14070000.0,
				Message:   "   \t\n   ",
			},
			expectedError: "message",
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

func TestPIRTTY_ParseArgs_JSONUnmarshalError(t *testing.T) {
	pirtty := &PIRTTY{}
	invalidJSON := []byte(`{"frequency": "invalid"}`)

	_, _, err := pirtty.ParseArgs(invalidJSON)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal args")
}

func TestPIRTTY_validateFrequency(t *testing.T) {
	tests := []struct {
		name        string
		frequency   float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid frequency - 14.070 MHz",
			frequency:   14070000.0,
			expectError: false,
		},
		{
			name:        "valid frequency - 50 kHz (minimum)",
			frequency:   50000.0,
			expectError: false,
		},
		{
			name:        "valid frequency - 1500 MHz (maximum)",
			frequency:   1500000000.0,
			expectError: false,
		},
		{
			name:        "zero frequency",
			frequency:   0.0,
			expectError: true,
			errorMsg:    "frequency must be positive",
		},
		{
			name:        "negative frequency",
			frequency:   -100000.0,
			expectError: true,
			errorMsg:    "frequency must be positive",
		},
		{
			name:        "frequency too low",
			frequency:   1000.0, // 1 kHz
			expectError: true,
			errorMsg:    "frequency out of RPiTX range",
		},
		{
			name:        "frequency too high",
			frequency:   2000000000.0, // 2 GHz
			expectError: true,
			errorMsg:    "frequency out of RPiTX range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pirtty := &PIRTTY{Frequency: tt.frequency}
			err := pirtty.validateFrequency()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPIRTTY_validateSpaceFrequency(t *testing.T) {
	tests := []struct {
		name           string
		spaceFrequency int
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "valid space frequency - 170",
			spaceFrequency: 170,
			expectError:    false,
		},
		{
			name:           "valid space frequency - 200",
			spaceFrequency: 200,
			expectError:    false,
		},
		{
			name:           "valid space frequency - 1",
			spaceFrequency: 1,
			expectError:    false,
		},
		{
			name:           "zero space frequency",
			spaceFrequency: 0,
			expectError:    true,
			errorMsg:       "space frequency must be positive",
		},
		{
			name:           "negative space frequency",
			spaceFrequency: -170,
			expectError:    true,
			errorMsg:       "space frequency must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pirtty := &PIRTTY{SpaceFrequency: &tt.spaceFrequency}
			err := pirtty.validateSpaceFrequency()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPIRTTY_validateMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
	}{
		{
			name:        "valid message",
			message:     "CQ DE N0CALL",
			expectError: false,
		},
		{
			name:        "valid long message",
			message:     "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG 1234567890",
			expectError: false,
		},
		{
			name:        "valid message with special characters",
			message:     "HELLO WORLD! TEST 123",
			expectError: false,
		},
		{
			name:        "empty message",
			message:     "",
			expectError: true,
		},
		{
			name:        "whitespace only message",
			message:     "   \t\n   ",
			expectError: true,
		},
		{
			name:        "single space",
			message:     " ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pirtty := &PIRTTY{Message: tt.message}
			err := pirtty.validateMessage()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, commonerrors.ErrRequiredFieldNotSet)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPIRTTY_buildArgs(t *testing.T) {
	tests := []struct {
		name         string
		pirtty       PIRTTY
		expectedArgs []string
	}{
		{
			name: "basic parameters",
			pirtty: PIRTTY{
				Frequency:      14070000.0,
				SpaceFrequency: intPtr(170),
				Message:        "TEST MESSAGE",
			},
			expectedArgs: []string{"14070000", "170", "TEST MESSAGE"},
		},
		{
			name: "different frequency and space",
			pirtty: PIRTTY{
				Frequency:      28070000.0,
				SpaceFrequency: intPtr(200),
				Message:        "ANOTHER TEST",
			},
			expectedArgs: []string{"28070000", "200", "ANOTHER TEST"},
		},
		{
			name: "message with spaces and special chars",
			pirtty: PIRTTY{
				Frequency:      7040000.0,
				SpaceFrequency: intPtr(85),
				Message:        "HELLO WORLD! 123",
			},
			expectedArgs: []string{"7040000", "85", "HELLO WORLD! 123"},
		},
		{
			name: "default space frequency",
			pirtty: PIRTTY{
				Frequency: 14070000.0,
				Message:   "DEFAULT TEST",
			},
			expectedArgs: []string{"14070000", "170", "DEFAULT TEST"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.pirtty.buildArgs()
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}