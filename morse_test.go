package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMORSE_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid complete args",
			input: map[string]any{
				"frequency": 14070000.0, // 14.070 MHz in Hz
				"rate":      20,         // 20 dits per minute
				"message":   "CQ DE N0CALL",
			},
			expectError: false,
			expectArgs:  []string{"14070000", "20", "CQ DE N0CALL"},
		},
		{
			name: "valid args with different frequency",
			input: map[string]any{
				"frequency": 7040000.0, // 7.040 MHz in Hz
				"rate":      15,
				"message":   "HELLO WORLD",
			},
			expectError: false,
			expectArgs:  []string{"7040000", "15", "HELLO WORLD"},
		},
		{
			name: "valid args with high rate",
			input: map[string]any{
				"frequency": 28070000.0, // 28.070 MHz in Hz
				"rate":      30,
				"message":   "TEST",
			},
			expectError: false,
			expectArgs:  []string{"28070000", "30", "TEST"},
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"rate":    20,
				"message": "TEST",
			},
			expectError: true,
		},
		{
			name: "missing rate",
			input: map[string]any{
				"frequency": 14070000.0,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "missing message",
			input: map[string]any{
				"frequency": 14070000.0,
				"rate":      20,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"frequency": 0.0,
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"frequency": -14070000.0,
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"frequency": 1000.0, // 1 kHz - below minimum
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"frequency": 2000000000.0, // 2 GHz - above maximum
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "zero rate",
			input: map[string]any{
				"frequency": 14070000.0,
				"rate":      0,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "negative rate",
			input: map[string]any{
				"frequency": 14070000.0,
				"rate":      -20,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "empty message",
			input: map[string]any{
				"frequency": 14070000.0,
				"rate":      20,
				"message":   "",
			},
			expectError: true,
		},
		{
			name: "whitespace only message",
			input: map[string]any{
				"frequency": 14070000.0,
				"rate":      20,
				"message":   "   ",
			},
			expectError: true,
		},
		{
			name: "invalid json",
			input: map[string]any{
				"frequency": "not_a_number",
				"rate":      20,
				"message":   "TEST",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			morse := &MORSE{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := morse.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestMORSE_BuildArgs(t *testing.T) {
	tests := []BuildArgsTest{
		{expectArgs: []string{"14070000", "20", "CQ DE N0CALL"}},
		{expectArgs: []string{"7040000", "15", "HELLO WORLD"}},
		{expectArgs: []string{"28070000", "30", "TEST MSG"}},
		{expectArgs: []string{"14070000", "20", "CQ DE N0CALL/P"}},
	}

	testNames := []string{
		"basic CW transmission",
		"low frequency CW",
		"high speed CW",
		"CW with portable callsign",
	}

	morseConfigs := []MORSE{
		{Frequency: 14070000.0, Rate: 20, Message: "CQ DE N0CALL"},
		{Frequency: 7040000.0, Rate: 15, Message: "HELLO WORLD"},
		{Frequency: 28070000.0, Rate: 30, Message: "TEST MSG"},
		{Frequency: 14070000.0, Rate: 20, Message: "CQ DE N0CALL/P"},
	}

	for i, tt := range tests {
		t.Run(testNames[i], func(t *testing.T) {
			morse := morseConfigs[i]
			RunBuildArgsTest(t, morse.buildArgs, tt)
		})
	}
}

func TestMORSE_ValidateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, []FrequencyValidationTest{
		{
			name:        "valid frequency 14.070 MHz",
			frequency:   14070000.0,
			expectError: false,
		},
		{
			name:        "valid frequency 7.040 MHz",
			frequency:   7040000.0,
			expectError: false,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			morse := &MORSE{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, morse.validateFrequency, tt)
		})
	}
}

func TestMORSE_ValidateRate(t *testing.T) {
	tests := []struct {
		name        string
		rate        int
		expectError bool
		errorType   error
	}{
		{
			name:        "valid rate 20",
			rate:        20,
			expectError: false,
		},
		{
			name:        "valid rate 15",
			rate:        15,
			expectError: false,
		},
		{
			name:        "valid rate 30",
			rate:        30,
			expectError: false,
		},
		{
			name:        "valid rate 1",
			rate:        1,
			expectError: false,
		},
		{
			name:        "zero rate",
			rate:        0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative rate",
			rate:        -20,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			morse := &MORSE{Rate: tt.rate}
			err := morse.validateRate()

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMORSE_ValidateMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid message",
			message:     "CQ DE N0CALL",
			expectError: false,
		},
		{
			name:        "valid single character",
			message:     "K",
			expectError: false,
		},
		{
			name:        "valid message with numbers",
			message:     "TEST 123",
			expectError: false,
		},
		{
			name:        "valid message with special chars",
			message:     "CQ DE N0CALL/P",
			expectError: false,
		},
		{
			name:        "empty message",
			message:     "",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "whitespace only message",
			message:     "   ",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "tab only message",
			message:     "\t",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "newline only message",
			message:     "\n",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			morse := &MORSE{Message: tt.message}
			err := morse.validateMessage()

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMORSE_Validate(t *testing.T) {
	tests := []struct {
		name        string
		morse       MORSE
		expectError bool
	}{
		{
			name: "valid complete morse",
			morse: MORSE{
				Frequency: 14070000.0,
				Rate:      20,
				Message:   "CQ DE N0CALL",
			},
			expectError: false,
		},
		{
			name: "valid different params",
			morse: MORSE{
				Frequency: 7040000.0,
				Rate:      15,
				Message:   "TEST",
			},
			expectError: false,
		},
		{
			name: "invalid - zero frequency",
			morse: MORSE{
				Frequency: 0.0,
				Rate:      20,
				Message:   "TEST",
			},
			expectError: true,
		},
		{
			name: "invalid - negative rate",
			morse: MORSE{
				Frequency: 14070000.0,
				Rate:      -20,
				Message:   "TEST",
			},
			expectError: true,
		},
		{
			name: "invalid - empty message",
			morse: MORSE{
				Frequency: 14070000.0,
				Rate:      20,
				Message:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.morse.validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
