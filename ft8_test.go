package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFT8_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid minimal args",
			input: map[string]any{
				"frequency": 14074000.0, // 14.074 MHz in Hz
				"message":   "CQ W1AW FN31",
			},
			expectError: false,
			expectArgs:  []string{"-f", "14074000", "-m", "CQ W1AW FN31"},
		},
		{
			name: "valid complete args",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "K0HAM W5XYZ",
				"ppm":       2.5,
				"offset":    1240.0,
				"slot":      1,
				"repeat":    true,
			},
			expectError: false,
			expectArgs: []string{
				"-f", "14074000", "-m", "K0HAM W5XYZ", "-p", "2.5",
				"-o", "1240", "-s", "1", "-r",
			},
		},
		{
			name: "valid with ppm only",
			input: map[string]any{
				"frequency": 7074000.0,
				"message":   "CQ K9ABC EM69",
				"ppm":       -1.5,
			},
			expectError: false,
			expectArgs:  []string{"-f", "7074000", "-m", "CQ K9ABC EM69", "-p", "-1.5"},
		},
		{
			name: "valid with offset only",
			input: map[string]any{
				"frequency": 21074000.0,
				"message":   "VE3XYZ K1AB",
				"offset":    2000.0,
			},
			expectError: false,
			expectArgs:  []string{"-f", "21074000", "-m", "VE3XYZ K1AB", "-o", "2000"},
		},
		{
			name: "valid with slot only",
			input: map[string]any{
				"frequency": 28074000.0,
				"message":   "W6QAR JA1XYZ",
				"slot":      0,
			},
			expectError: false,
			expectArgs:  []string{"-f", "28074000", "-m", "W6QAR JA1XYZ", "-s", "0"},
		},
		{
			name: "valid with repeat false",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "N5QAM 73",
				"repeat":    false,
			},
			expectError: false,
			expectArgs:  []string{"-f", "14074000", "-m", "N5QAM 73"},
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"message": "TEST",
			},
			expectError: true,
		},
		{
			name: "missing message",
			input: map[string]any{
				"frequency": 14074000.0,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"frequency": 0.0,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"frequency": -14074000.0,
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"frequency": 1000.0, // 1 kHz - below minimum
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"frequency": 2000000000.0, // 2 GHz - above maximum
				"message":   "TEST",
			},
			expectError: true,
		},
		{
			name: "empty message",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "",
			},
			expectError: true,
		},
		{
			name: "whitespace only message",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "   ",
			},
			expectError: true,
		},
		{
			name: "offset too low",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "TEST",
				"offset":    -100.0,
			},
			expectError: true,
		},
		{
			name: "offset too high",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "TEST",
				"offset":    3000.0,
			},
			expectError: true,
		},
		{
			name: "invalid slot negative",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "TEST",
				"slot":      -1,
			},
			expectError: true,
		},
		{
			name: "invalid slot too high",
			input: map[string]any{
				"frequency": 14074000.0,
				"message":   "TEST",
				"slot":      3,
			},
			expectError: true,
		},
		{
			name: "invalid json",
			input: map[string]any{
				"frequency": "not_a_number",
				"message":   "TEST",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := ft8.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestFT8_BuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		ft8        FT8
		expectArgs []string
	}{
		{
			name: "minimal ft8 transmission",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "CQ W1AW FN31",
			},
			expectArgs: []string{"-f", "14074000", "-m", "CQ W1AW FN31"},
		},
		{
			name: "complete ft8 transmission",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "K0HAM W5XYZ",
				PPM: func() *float64 {
					v := 2.5

					return &v
				}(),
				Offset: func() *float64 {
					v := 1240.0

					return &v
				}(),
				Slot: func() *int {
					v := 1

					return &v
				}(),
				Repeat: func() *bool {
					v := true

					return &v
				}(),
			},
			expectArgs: []string{
				"-f", "14074000", "-m", "K0HAM W5XYZ", "-p", "2.5",
				"-o", "1240", "-s", "1", "-r",
			},
		},
		{
			name: "with negative ppm",
			ft8: FT8{
				Frequency: 7074000.0,
				Message:   "CQ K9ABC EM69",
				PPM: func() *float64 {
					v := -1.5

					return &v
				}(),
			},
			expectArgs: []string{"-f", "7074000", "-m", "CQ K9ABC EM69", "-p", "-1.5"},
		},
		{
			name: "with custom offset",
			ft8: FT8{
				Frequency: 21074000.0,
				Message:   "VE3XYZ K1AB",
				Offset: func() *float64 {
					v := 2000.0

					return &v
				}(),
			},
			expectArgs: []string{"-f", "21074000", "-m", "VE3XYZ K1AB", "-o", "2000"},
		},
		{
			name: "with slot 0",
			ft8: FT8{
				Frequency: 28074000.0,
				Message:   "W6QAR JA1XYZ",
				Slot: func() *int {
					v := 0

					return &v
				}(),
			},
			expectArgs: []string{"-f", "28074000", "-m", "W6QAR JA1XYZ", "-s", "0"},
		},
		{
			name: "with slot 2 (always)",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "CQ DX KL7ABC",
				Slot: func() *int {
					v := 2

					return &v
				}(),
			},
			expectArgs: []string{"-f", "14074000", "-m", "CQ DX KL7ABC", "-s", "2"},
		},
		{
			name: "with repeat false (should not add -r flag)",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "N5QAM 73",
				Repeat: func() *bool {
					v := false

					return &v
				}(),
			},
			expectArgs: []string{"-f", "14074000", "-m", "N5QAM 73"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.ft8.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestFT8_ValidateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, []FrequencyValidationTest{
		{
			name:        "valid frequency 14.074 MHz",
			frequency:   14074000.0,
			expectError: false,
		},
		{
			name:        "valid frequency 7.074 MHz",
			frequency:   7074000.0,
			expectError: false,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, ft8.validateFrequency, tt)
		})
	}
}

func TestFT8_ValidateMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid message",
			message:     "CQ W1AW FN31",
			expectError: false,
		},
		{
			name:        "valid message with call",
			message:     "K0HAM W5XYZ",
			expectError: false,
		},
		{
			name:        "valid message at limit",
			message:     "VE7ABC R-08",
			expectError: false,
		},
		{
			name:        "valid single character",
			message:     "K",
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
			name:        "longer message allowed",
			message:     "CQ W9DEF EM59",
			expectError: false,
		},
		{
			name:        "very long message allowed",
			message:     "W8ABC DX EM79 CALLING CQ FROM GRID SQUARE",
			expectError: false,
		},
		{
			name: "any length message allowed",
			message: "VK2ABC G0XYZ 73 THIS IS A VERY LONG MESSAGE FOR TESTING " +
				"PURPOSES",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{Message: tt.message}
			err := ft8.validateMessage()

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

func TestFT8_ValidatePPM(t *testing.T) {
	tests := []struct {
		name        string
		ppm         *float64
		expectError bool
	}{
		{
			name:        "nil ppm (not set)",
			ppm:         nil,
			expectError: false,
		},
		{
			name: "positive ppm",
			ppm: func() *float64 {
				v := 2.5

				return &v
			}(),
			expectError: false,
		},
		{
			name: "negative ppm",
			ppm: func() *float64 {
				v := -1.5

				return &v
			}(),
			expectError: false,
		},
		{
			name: "zero ppm",
			ppm: func() *float64 {
				v := 0.0

				return &v
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{PPM: tt.ppm}
			err := ft8.validatePPM()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFT8_ValidateOffset(t *testing.T) {
	tests := []struct {
		name        string
		offset      *float64
		expectError bool
		errorType   error
	}{
		{
			name:        "nil offset (not set)",
			offset:      nil,
			expectError: false,
		},
		{
			name: "valid offset default",
			offset: func() *float64 {
				v := 1240.0

				return &v
			}(),
			expectError: false,
		},
		{
			name: "valid offset minimum",
			offset: func() *float64 {
				v := 0.0

				return &v
			}(),
			expectError: false,
		},
		{
			name: "valid offset maximum",
			offset: func() *float64 {
				v := 2500.0

				return &v
			}(),
			expectError: false,
		},
		{
			name: "valid offset mid-range",
			offset: func() *float64 {
				v := 1500.0

				return &v
			}(),
			expectError: false,
		},
		{
			name: "offset too low",
			offset: func() *float64 {
				v := -100.0

				return &v
			}(),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name: "offset too high",
			offset: func() *float64 {
				v := 3000.0

				return &v
			}(),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{Offset: tt.offset}
			err := ft8.validateOffset()

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

func TestFT8_ValidateSlot(t *testing.T) {
	tests := []struct {
		name        string
		slot        *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil slot (not set)",
			slot:        nil,
			expectError: false,
		},
		{
			name: "valid slot 0",
			slot: func() *int {
				v := 0

				return &v
			}(),
			expectError: false,
		},
		{
			name: "valid slot 1",
			slot: func() *int {
				v := 1

				return &v
			}(),
			expectError: false,
		},
		{
			name: "valid slot 2 (always)",
			slot: func() *int {
				v := 2

				return &v
			}(),
			expectError: false,
		},
		{
			name: "invalid slot negative",
			slot: func() *int {
				v := -1

				return &v
			}(),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name: "invalid slot too high",
			slot: func() *int {
				v := 3

				return &v
			}(),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft8 := &FT8{Slot: tt.slot}
			err := ft8.validateSlot()

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

func TestFT8_Validate(t *testing.T) {
	tests := []struct {
		name        string
		ft8         FT8
		expectError bool
	}{
		{
			name: "valid minimal ft8",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "CQ W1AW FN31",
			},
			expectError: false,
		},
		{
			name: "valid complete ft8",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "K0HAM W5XYZ",
				PPM: func() *float64 {
					v := 2.5

					return &v
				}(),
				Offset: func() *float64 {
					v := 1240.0

					return &v
				}(),
				Slot: func() *int {
					v := 1

					return &v
				}(),
				Repeat: func() *bool {
					v := true

					return &v
				}(),
			},
			expectError: false,
		},
		{
			name: "invalid - zero frequency",
			ft8: FT8{
				Frequency: 0.0,
				Message:   "TEST",
			},
			expectError: true,
		},
		{
			name: "invalid - empty message",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "",
			},
			expectError: true,
		},
		{
			name: "invalid - offset too high",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "TEST",
				Offset: func() *float64 {
					v := 3000.0

					return &v
				}(),
			},
			expectError: true,
		},
		{
			name: "invalid - slot too high",
			ft8: FT8{
				Frequency: 14074000.0,
				Message:   "TEST",
				Slot: func() *int {
					v := 3

					return &v
				}(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ft8.validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
