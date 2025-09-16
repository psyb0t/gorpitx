package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTUNE_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid minimal args - frequency only",
			input: map[string]any{
				"frequency": 434000000.0, // 434 MHz in Hz
			},
			expectError: false,
			expectArgs:  []string{"-f", "434000000"},
		},
		{
			name: "valid complete args",
			input: map[string]any{
				"frequency":     434000000.0, // 434 MHz in Hz
				"exitImmediate": true,
				"ppm":           2.5,
			},
			expectError: false,
			expectArgs:  []string{"-f", "434000000", "-e", "-p", "2.5"},
		},
		{
			name: "valid args with exit immediate false",
			input: map[string]any{
				"frequency":     107900000.0, // 107.9 MHz in Hz
				"exitImmediate": false,
				"ppm":           1.0,
			},
			expectError: false,
			expectArgs:  []string{"-f", "107900000", "-p", "1"},
		},
		{
			name: "valid args with ppm only",
			input: map[string]any{
				"frequency": 144500000.0, // 144.5 MHz in Hz
				"ppm":       0.5,
			},
			expectError: false,
			expectArgs:  []string{"-f", "144500000", "-p", "0.5"},
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"exitImmediate": true,
				"ppm":           2.5,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"frequency": 0.0,
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"frequency": -434000000.0,
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"frequency": 1000.0, // 1 kHz - below minimum
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"frequency": 2000000000.0, // 2 GHz - above maximum
			},
			expectError: true,
		},
		{
			name: "negative PPM",
			input: map[string]any{
				"frequency": 434000000.0,
				"ppm":       -1.0,
			},
			expectError: true,
		},
		{
			name: "zero PPM",
			input: map[string]any{
				"frequency": 434000000.0,
				"ppm":       0.0,
			},
			expectError: true,
		},
		{
			name: "invalid json",
			input: map[string]any{
				"frequency": "not_a_number",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tune := &TUNE{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, err := tune.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestTUNE_BuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		tune       TUNE
		expectArgs []string
	}{
		{
			name: "frequency only",
			tune: TUNE{
				Frequency: floatPtr(434000000.0),
			},
			expectArgs: []string{"-f", "434000000"},
		},
		{
			name: "frequency with exit immediate true",
			tune: TUNE{
				Frequency:     floatPtr(434000000.0),
				ExitImmediate: boolPtr(true),
			},
			expectArgs: []string{"-f", "434000000", "-e"},
		},
		{
			name: "frequency with exit immediate false",
			tune: TUNE{
				Frequency:     floatPtr(434000000.0),
				ExitImmediate: boolPtr(false),
			},
			expectArgs: []string{"-f", "434000000"},
		},
		{
			name: "frequency with ppm",
			tune: TUNE{
				Frequency: floatPtr(434000000.0),
				PPM:       floatPtr(2.5),
			},
			expectArgs: []string{"-f", "434000000", "-p", "2.5"},
		},
		{
			name: "all parameters",
			tune: TUNE{
				Frequency:     floatPtr(434000000.0),
				ExitImmediate: boolPtr(true),
				PPM:           floatPtr(1.5),
			},
			expectArgs: []string{"-f", "434000000", "-e", "-p", "1.5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.tune.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestTUNE_ValidateFreq(t *testing.T) {
	tests := []struct {
		name        string
		frequency   *float64
		expectError bool
		errorType   error
	}{
		{
			name:        "nil frequency",
			frequency:   nil,
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "zero frequency",
			frequency:   floatPtr(0.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative frequency",
			frequency:   floatPtr(-100000.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "frequency too low",
			frequency:   floatPtr(1000.0), // 1 kHz
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
		{
			name:        "frequency too high",
			frequency:   floatPtr(2000000000.0), // 2 GHz
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
		{
			name:        "valid minimum frequency",
			frequency:   floatPtr(50000.0), // 50 kHz
			expectError: false,
		},
		{
			name:        "valid typical frequency",
			frequency:   floatPtr(434000000.0), // 434 MHz
			expectError: false,
		},
		{
			name:        "valid maximum frequency",
			frequency:   floatPtr(1500000000.0), // 1500 MHz
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tune := &TUNE{Frequency: tt.frequency}
			err := tune.validateFreq()

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

func TestTUNE_ValidatePPM(t *testing.T) {
	tests := []struct {
		name        string
		ppm         *float64
		expectError bool
		errorType   error
	}{
		{
			name:        "nil ppm",
			ppm:         nil,
			expectError: false,
		},
		{
			name:        "zero ppm",
			ppm:         floatPtr(0.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative ppm",
			ppm:         floatPtr(-1.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "valid positive ppm",
			ppm:         floatPtr(2.5),
			expectError: false,
		},
		{
			name:        "valid small positive ppm",
			ppm:         floatPtr(0.1),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tune := &TUNE{PPM: tt.ppm}
			err := tune.validatePPM()

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

func TestTUNE_Validate(t *testing.T) {
	tests := []struct {
		name        string
		tune        TUNE
		expectError bool
	}{
		{
			name: "valid minimal tune",
			tune: TUNE{
				Frequency: floatPtr(434000000.0),
			},
			expectError: false,
		},
		{
			name: "valid complete tune",
			tune: TUNE{
				Frequency:     floatPtr(434000000.0),
				ExitImmediate: boolPtr(true),
				PPM:           floatPtr(2.5),
			},
			expectError: false,
		},
		{
			name: "invalid - missing frequency",
			tune: TUNE{
				ExitImmediate: boolPtr(true),
				PPM:           floatPtr(2.5),
			},
			expectError: true,
		},
		{
			name: "invalid - negative frequency",
			tune: TUNE{
				Frequency: floatPtr(-434000000.0),
				PPM:       floatPtr(2.5),
			},
			expectError: true,
		},
		{
			name: "invalid - negative ppm",
			tune: TUNE{
				Frequency: floatPtr(434000000.0),
				PPM:       floatPtr(-1.0),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tune.validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for test pointer creation
func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}