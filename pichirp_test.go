package gorpitx

import (
	"encoding/json"
	"testing"

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
	tests := []BuildArgsTest{
		{expectArgs: []string{"434000000", "100000", "5"}},
		{expectArgs: []string{"144500000", "50000", "10.5"}},
		{expectArgs: []string{"1296000000", "1000000", "0.5"}},
		{expectArgs: []string{"28070000", "1000", "1"}},
	}

	testNames := []string{
		"UHF chirp sweep",
		"VHF chirp with decimal time",
		"microwave wideband chirp",
		"HF narrowband chirp",
	}

	chirpConfigs := []PICHIRP{
		{Frequency: 434000000.0, Bandwidth: 100000.0, Time: 5.0},
		{Frequency: 144500000.0, Bandwidth: 50000.0, Time: 10.5},
		{Frequency: 1296000000.0, Bandwidth: 1000000.0, Time: 0.5},
		{Frequency: 28070000.0, Bandwidth: 1000.0, Time: 1.0},
	}

	for i, tt := range tests {
		t.Run(testNames[i], func(t *testing.T) {
			pichirp := chirpConfigs[i]
			RunBuildArgsTest(t, pichirp.buildArgs, tt)
		})
	}
}

func TestPICHIRP_ValidateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, FrequencyValidationTest{
		name:        "valid frequency - 434 MHz",
		frequency:   434000000.0,
		expectError: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, pichirp.validateFrequency, tt)
		})
	}
}

func TestPICHIRP_ValidateBandwidth(t *testing.T) {
	tests := GetStandardPositiveValidationTests()
	tests = append(tests, []PositiveValidationTest{
		{
			name:        "valid bandwidth - 100 kHz",
			value:       100000.0,
			expectError: false,
		},
		{
			name:        "valid bandwidth - 1 MHz",
			value:       1000000.0,
			expectError: false,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Bandwidth: tt.value}
			RunPositiveValidationTest(t, pichirp.validateBandwidth, tt)
		})
	}
}

func TestPICHIRP_ValidateTime(t *testing.T) {
	tests := GetStandardPositiveValidationTests()
	tests = append(tests, []PositiveValidationTest{
		{
			name:        "valid time - 5 seconds",
			value:       5.0,
			expectError: false,
		},
		{
			name:        "valid time - 60 seconds",
			value:       60.0,
			expectError: false,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pichirp := &PICHIRP{Time: tt.value}
			RunPositiveValidationTest(t, pichirp.validateTime, tt)
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
