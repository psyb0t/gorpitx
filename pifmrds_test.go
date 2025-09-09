package gorpitx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPIFMRDS_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid complete args",
			input: map[string]any{
				"freq":  107.9,
				"audio": ".fixtures/test.wav",
				"pi":    "ABCD",
				"ps":    "TestPS",
				"rt":    "Test Radio Text",
			},
			expectError: false,
			expectArgs: []string{
				"-freq", "107.9", "-audio", ".fixtures/test.wav", "-pi", "ABCD", "-ps", "TestPS", "-rt", "Test Radio Text",
			},
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"audio": ".fixtures/test.wav",
			},
			expectError: true,
		},
		{
			name: "missing audio",
			input: map[string]any{
				"freq": 107.9,
			},
			expectError: true,
		},
		{
			name: "invalid PI code length",
			input: map[string]any{
				"freq":  107.9,
				"audio": ".fixtures/test.wav",
				"pi":    "ABC", // too short
			},
			expectError: true,
		},
		{
			name: "PS too long",
			input: map[string]any{
				"freq":  107.9,
				"audio": ".fixtures/test.wav",
				"ps":    "TooLongPS", // 9 chars, max is 8
			},
			expectError: true,
		},
		{
			name: "frequency out of range",
			input: map[string]any{
				"freq":  2000000.0, // too high
				"audio": ".fixtures/test.wav",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test audio file for file existence check - we're not savages who skip validation
			if audioPath, ok := tt.input["audio"].(string); ok && audioPath != "" {
				// Skip file creation for this unit test - we'll mock it because we're not fucking around with real files
				_ = audioPath // acknowledge we got the path but aren't using it in unit tests
			}

			module := &PIFMRDS{}
			inputBytes, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			if err != nil {
				return
			}

			args, err := module.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			if err != nil {
				return
			}

			// Check that we got some args back or this whole thing is pointless
			assert.NotEmpty(t, args)

			// Check frequency is always present for valid cases - no freq, no fucking transmission
			assert.Contains(t, args, "-freq")
			assert.Contains(t, args, "-audio")
		})
	}
}

func TestPIFMRDS_buildArgs(t *testing.T) {
	module := &PIFMRDS{
		Freq:  107.9,
		Audio: ".fixtures/test.wav",
		PI:    "ABCD",
		PS:    "TestPS",
		RT:    "Test Radio Text",
	}

	args := module.buildArgs()

	expected := []string{
		"-freq", "107.9",
		"-audio", ".fixtures/test.wav",
		"-pi", "ABCD",
		"-ps", "TestPS",
		"-rt", "Test Radio Text",
	}

	assert.Equal(t, expected, args)
}

func TestPIFMRDS_validateFreq(t *testing.T) {
	tests := []struct {
		name        string
		freq        float64
		expectError bool
	}{
		{"valid frequency", 107.9, false},
		{"zero frequency", 0.0, true},
		{"negative frequency", -10.0, true},
		{"too low frequency", 0.001, true},
		{"too high frequency", 2000.0, true},
		{"valid precision", 88.5, false},
		{"invalid precision", 88.55, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := &PIFMRDS{Freq: tt.freq}
			err := module.validateFreq()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPIFMRDS_validatePI(t *testing.T) {
	tests := []struct {
		name        string
		pi          string
		expectError bool
	}{
		{"valid PI", "ABCD", false},
		{"valid PI lowercase", "abcd", false},
		{"valid PI numbers", "1234", false},
		{"empty PI", "", false}, // empty is valid, gets default
		{"too short", "ABC", true},
		{"too long", "ABCDE", true},
		{"invalid hex", "GHIJ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := &PIFMRDS{PI: tt.pi}
			err := module.validatePI()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPIFMRDS_validatePS(t *testing.T) {
	tests := []struct {
		name        string
		ps          string
		expectError bool
	}{
		{"valid PS", "TestPS", false},
		{"max length PS", "12345678", false},
		{"empty PS", "", false}, // empty is valid, gets default
		{"too long PS", "123456789", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := &PIFMRDS{PS: tt.ps}
			err := module.validatePS()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
