package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPICHIRP_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid complete args",
			input: map[string]any{
				"frequency": 434000000.0, // 434 MHz in Hz
				"bandwidth": 100000.0,    // 100 kHz bandwidth
				"time":      5.0,         // 5 seconds
			},
			expectError: false,
			expectArgs:  []string{"434000000", "100000", "5"},
		},
		{
			name: "valid args with different values",
			input: map[string]any{
				"frequency": 144500000.0, // 144.5 MHz in Hz
				"bandwidth": 50000.0,     // 50 kHz bandwidth
				"time":      10.5,        // 10.5 seconds
			},
			expectError: false,
			expectArgs:  []string{"144500000", "50000", "10.5"},
		},
		{
			name: "valid args with small bandwidth",
			input: map[string]any{
				"frequency": 28070000.0, // 28.070 MHz in Hz
				"bandwidth": 1000.0,     // 1 kHz bandwidth
				"time":      1.0,        // 1 second
			},
			expectError: false,
			expectArgs:  []string{"28070000", "1000", "1"},
		},
		{
			name: "valid args with large bandwidth",
			input: map[string]any{
				"frequency": 1296000000.0, // 1296 MHz in Hz
				"bandwidth": 1000000.0,    // 1 MHz bandwidth
				"time":      0.5,          // 0.5 seconds
			},
			expectError: false,
			expectArgs:  []string{"1296000000", "1000000", "0.5"},
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "missing bandwidth",
			input: map[string]any{
				"frequency": 434000000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "missing time",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": 100000.0,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"frequency": 0.0,
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"frequency": -434000000.0,
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"frequency": 1000.0, // 1 kHz - below minimum
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"frequency": 2000000000.0, // 2 GHz - above maximum
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "zero bandwidth",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": 0.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "negative bandwidth",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": -100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "zero time",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": 100000.0,
				"time":      0.0,
			},
			expectError: true,
		},
		{
			name: "negative time",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": 100000.0,
				"time":      -5.0,
			},
			expectError: true,
		},
		{
			name: "invalid json - frequency as string",
			input: map[string]any{
				"frequency": "not_a_number",
				"bandwidth": 100000.0,
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "invalid json - bandwidth as string",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": "not_a_number",
				"time":      5.0,
			},
			expectError: true,
		},
		{
			name: "invalid json - time as string",
			input: map[string]any{
				"frequency": 434000000.0,
				"bandwidth": 100000.0,
				"time":      "not_a_number",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := pichirp.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPICHIRP_BuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		pichirp    PICHIRP
		expectArgs []string
	}{
		{
			name: "basic chirp transmission",
			pichirp: PICHIRP{
				Frequency: 434000000.0,
				Bandwidth: 100000.0,
				Time:      5.0,
			},
			expectArgs: []string{"434000000", "100000", "5"},
		},
		{
			name: "different frequency and bandwidth",
			pichirp: PICHIRP{
				Frequency: 144500000.0,
				Bandwidth: 50000.0,
				Time:      10.5,
			},
			expectArgs: []string{"144500000", "50000", "10.5"},
		},
		{
			name: "high frequency with large bandwidth",
			pichirp: PICHIRP{
				Frequency: 1296000000.0,
				Bandwidth: 1000000.0,
				Time:      0.5,
			},
			expectArgs: []string{"1296000000", "1000000", "0.5"},
		},
		{
			name: "low frequency with small bandwidth",
			pichirp: PICHIRP{
				Frequency: 28070000.0,
				Bandwidth: 1000.0,
				Time:      1.0,
			},
			expectArgs: []string{"28070000", "1000", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.pichirp.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPICHIRP_ValidateFrequency(t *testing.T) {
	tests := []struct {
		name        string
		frequency   float64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid frequency - 434 MHz",
			frequency:   434000000.0,
			expectError: false,
		},
		{
			name:        "valid frequency - minimum (5 kHz)",
			frequency:   5000.0,
			expectError: false,
		},
		{
			name:        "valid frequency - maximum (1500 MHz)",
			frequency:   1500000000.0,
			expectError: false,
		},
		{
			name:        "zero frequency",
			frequency:   0.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative frequency",
			frequency:   -434000000.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "frequency too low (1 kHz)",
			frequency:   1000.0,
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
		{
			name:        "frequency too high (2 GHz)",
			frequency:   2000000000.0,
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Frequency: tt.frequency}
			err := pichirp.validateFrequency()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestPICHIRP_ValidateBandwidth(t *testing.T) {
	tests := []struct {
		name        string
		bandwidth   float64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid bandwidth - 100 kHz",
			bandwidth:   100000.0,
			expectError: false,
		},
		{
			name:        "valid bandwidth - 1 Hz",
			bandwidth:   1.0,
			expectError: false,
		},
		{
			name:        "valid bandwidth - 1 MHz",
			bandwidth:   1000000.0,
			expectError: false,
		},
		{
			name:        "zero bandwidth",
			bandwidth:   0.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative bandwidth",
			bandwidth:   -100000.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Bandwidth: tt.bandwidth}
			err := pichirp.validateBandwidth()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestPICHIRP_ValidateTime(t *testing.T) {
	tests := []struct {
		name        string
		time        float64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid time - 5 seconds",
			time:        5.0,
			expectError: false,
		},
		{
			name:        "valid time - 0.1 seconds",
			time:        0.1,
			expectError: false,
		},
		{
			name:        "valid time - 60 seconds",
			time:        60.0,
			expectError: false,
		},
		{
			name:        "zero time",
			time:        0.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative time",
			time:        -5.0,
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Time: tt.time}
			err := pichirp.validateTime()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestPICHIRP_Validate(t *testing.T) {
	tests := []struct {
		name        string
		pichirp     PICHIRP
		expectError bool
	}{
		{
			name: "valid complete configuration",
			pichirp: PICHIRP{
				Frequency: 434000000.0,
				Bandwidth: 100000.0,
				Time:      5.0,
			},
			expectError: false,
		},
		{
			name: "invalid frequency",
			pichirp: PICHIRP{
				Frequency: 0.0, // Invalid
				Bandwidth: 100000.0,
				Time:      5.0,
			},
			expectError: true,
		},
		{
			name: "invalid bandwidth",
			pichirp: PICHIRP{
				Frequency: 434000000.0,
				Bandwidth: 0.0, // Invalid
				Time:      5.0,
			},
			expectError: true,
		},
		{
			name: "invalid time",
			pichirp: PICHIRP{
				Frequency: 434000000.0,
				Bandwidth: 100000.0,
				Time:      0.0, // Invalid
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pichirp.validate()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}