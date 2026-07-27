package gorpitx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSENDIQ_ParseArgs(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.iq")
	err := os.WriteFile(testFile, []byte("test data"), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid minimal args - required only",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000"},
		},
		{
			name: "valid with stdin input",
			input: map[string]any{
				"inputFile": "-",
				"freq":      434000000.0,
			},
			expectError: false,
			expectArgs:  []string{"-i", "-", "-f", "434000000"},
		},
		{
			name: "valid complete args",
			input: map[string]any{
				"inputFile":      testFile,
				"freq":           434000000.0,
				"sampleRate":     96000,
				"harmonic":       2,
				"iqType":         "float",
				"power":          2.5,
				"sharedMemToken": 12345,
				"loopMode":       true,
			},
			expectError: false,
			expectArgs: []string{
				"-i", testFile,
				"-f", "434000000",
				"-s", "96000",
				"-h", "2",
				"-t", "float",
				"-p", "2.50",
				"-m", "12345",
				"-l",
			},
		},
		{
			name: "valid with i16 type",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"iqType":    "i16",
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-t", "i16"},
		},
		{
			name: "valid with u8 type",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"iqType":    "u8",
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-t", "u8"},
		},
		{
			name: "valid with double type",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"iqType":    "double",
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-t", "double"},
		},
		{
			name: "valid with minimum sample rate",
			input: map[string]any{
				"inputFile":  testFile,
				"freq":       434000000.0,
				"sampleRate": 10000,
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-s", "10000"},
		},
		{
			name: "valid with maximum sample rate",
			input: map[string]any{
				"inputFile":  testFile,
				"freq":       434000000.0,
				"sampleRate": 2000000,
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-s", "2000000"},
		},
		{
			name: "valid with power clamping to minimum",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"power":     -1.0, // Should be clamped to 0.0
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-p", "0.00"},
		},
		{
			name: "valid with power clamping to maximum",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"power":     10.0, // Should be clamped to 7.0
			},
			expectError: false,
			expectArgs:  []string{"-i", testFile, "-f", "434000000", "-p", "7.00"},
		},
		{
			name: "missing inputFile",
			input: map[string]any{
				"freq": 434000000.0,
			},
			expectError: true,
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"inputFile": testFile,
			},
			expectError: true,
		},
		{
			name: "non-existent file",
			input: map[string]any{
				"inputFile": "/nonexistent/file.iq",
				"freq":      434000000.0,
			},
			expectError: true,
		},
		{
			name: "invalid IQ type",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"iqType":    "invalid",
			},
			expectError: true,
		},
		{
			name: "sample rate too low",
			input: map[string]any{
				"inputFile":  testFile,
				"freq":       434000000.0,
				"sampleRate": 5000,
			},
			expectError: true,
		},
		{
			name: "sample rate too high",
			input: map[string]any{
				"inputFile":  testFile,
				"freq":       434000000.0,
				"sampleRate": 3000000,
			},
			expectError: true,
		},
		{
			name: "zero harmonic",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"harmonic":  0,
			},
			expectError: true,
		},
		{
			name: "negative harmonic",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      434000000.0,
				"harmonic":  -1,
			},
			expectError: true,
		},
		{
			name: "zero shared memory token",
			input: map[string]any{
				"inputFile":      testFile,
				"freq":           434000000.0,
				"sharedMemToken": 0,
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      1000.0,
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      2000000000.0,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      0.0,
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"inputFile": testFile,
				"freq":      -434000000.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := sendiq.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestSENDIQ_BuildArgs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.iq")

	tests := []struct {
		name       string
		sendiq     SENDIQ
		expectArgs []string
	}{
		{
			name: "minimal args",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000"},
		},
		{
			name: "with stdin",
			sendiq: SENDIQ{
				InputFile: "-",
				Freq:      434000000.0,
			},
			expectArgs: []string{"-i", "-", "-f", "434000000"},
		},
		{
			name: "with sample rate",
			sendiq: SENDIQ{
				InputFile:  testFile,
				Freq:       434000000.0,
				SampleRate: new(96000),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-s", "96000"},
		},
		{
			name: "with harmonic",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				Harmonic:  new(3),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-h", "3"},
		},
		{
			name: "with IQ type i16",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				IQType:    new("i16"),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-t", "i16"},
		},
		{
			name: "with IQ type u8",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				IQType:    new("u8"),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-t", "u8"},
		},
		{
			name: "with IQ type float",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				IQType:    new("float"),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-t", "float"},
		},
		{
			name: "with IQ type double",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				IQType:    new("double"),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-t", "double"},
		},
		{
			name: "with power",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				Power:     new(3.5),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-p", "3.50"},
		},
		{
			name: "with shared memory token",
			sendiq: SENDIQ{
				InputFile:      testFile,
				Freq:           434000000.0,
				SharedMemToken: new(12345),
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-m", "12345"},
		},
		{
			name: "with loop mode",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				LoopMode:  true,
			},
			expectArgs: []string{"-i", testFile, "-f", "434000000", "-l"},
		},
		{
			name: "complete args",
			sendiq: SENDIQ{
				InputFile:      testFile,
				Freq:           434000000.0,
				SampleRate:     new(96000),
				Harmonic:       new(2),
				IQType:         new("float"),
				Power:          new(2.5),
				SharedMemToken: new(12345),
				LoopMode:       true,
			},
			expectArgs: []string{
				"-i", testFile,
				"-f", "434000000",
				"-s", "96000",
				"-h", "2",
				"-t", "float",
				"-p", "2.50",
				"-m", "12345",
				"-l",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.sendiq.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestSENDIQ_ValidateInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.iq")
	err := os.WriteFile(existingFile, []byte("test"), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		inputFile   string
		expectError bool
		errorType   error
	}{
		{
			name:        "empty input file",
			inputFile:   "",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "stdin input",
			inputFile:   "-",
			expectError: false,
		},
		{
			name:        "existing file",
			inputFile:   existingFile,
			expectError: false,
		},
		{
			name:        "non-existent file",
			inputFile:   "/nonexistent/file.iq",
			expectError: true,
			errorType:   commonerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{InputFile: tt.inputFile}
			err := sendiq.validateInputFile()

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

func TestSENDIQ_ValidateFreq(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{Freq: tt.frequency}
			RunFrequencyValidationTest(t, sendiq.validateFreq, tt)
		})
	}
}

func TestSENDIQ_ValidateSampleRate(t *testing.T) {
	tests := []struct {
		name        string
		sampleRate  *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil sample rate",
			sampleRate:  nil,
			expectError: false,
		},
		{
			name:        "minimum valid sample rate",
			sampleRate:  new(10000),
			expectError: false,
		},
		{
			name:        "typical sample rate",
			sampleRate:  new(48000),
			expectError: false,
		},
		{
			name:        "high sample rate",
			sampleRate:  new(250000),
			expectError: false,
		},
		{
			name:        "maximum valid sample rate",
			sampleRate:  new(2000000),
			expectError: false,
		},
		{
			name:        "sample rate too low",
			sampleRate:  new(5000),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "sample rate too high",
			sampleRate:  new(3000000),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{SampleRate: tt.sampleRate}
			err := sendiq.validateSampleRate()

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

func TestSENDIQ_ValidateHarmonic(t *testing.T) {
	tests := []struct {
		name        string
		harmonic    *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil harmonic",
			harmonic:    nil,
			expectError: false,
		},
		{
			name:        "harmonic 1",
			harmonic:    new(1),
			expectError: false,
		},
		{
			name:        "harmonic 2",
			harmonic:    new(2),
			expectError: false,
		},
		{
			name:        "harmonic 10",
			harmonic:    new(10),
			expectError: false,
		},
		{
			name:        "zero harmonic",
			harmonic:    new(0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative harmonic",
			harmonic:    new(-1),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{Harmonic: tt.harmonic}
			err := sendiq.validateHarmonic()

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

func TestSENDIQ_ValidateIQType(t *testing.T) {
	tests := []struct {
		name        string
		iqType      *string
		expectError bool
		errorType   error
	}{
		{
			name:        "nil IQ type",
			iqType:      nil,
			expectError: false,
		},
		{
			name:        "valid i16 type",
			iqType:      new("i16"),
			expectError: false,
		},
		{
			name:        "valid u8 type",
			iqType:      new("u8"),
			expectError: false,
		},
		{
			name:        "valid float type",
			iqType:      new("float"),
			expectError: false,
		},
		{
			name:        "valid double type",
			iqType:      new("double"),
			expectError: false,
		},
		{
			name:        "invalid type",
			iqType:      new("invalid"),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid type i32",
			iqType:      new("i32"),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid type uppercase",
			iqType:      new("I16"),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{IQType: tt.iqType}
			err := sendiq.validateIQType()

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

func TestSENDIQ_ValidatePower(t *testing.T) {
	tests := []struct {
		name          string
		power         *float64
		expectError   bool
		expectedValue *float64
	}{
		{
			name:        "nil power",
			power:       nil,
			expectError: false,
		},
		{
			name:          "valid power within range",
			power:         new(3.5),
			expectError:   false,
			expectedValue: new(3.5),
		},
		{
			name:          "minimum power",
			power:         new(0.0),
			expectError:   false,
			expectedValue: new(0.0),
		},
		{
			name:          "maximum power",
			power:         new(7.0),
			expectError:   false,
			expectedValue: new(7.0),
		},
		{
			name:          "power below minimum - should clamp",
			power:         new(-1.0),
			expectError:   false,
			expectedValue: new(0.0),
		},
		{
			name:          "power above maximum - should clamp",
			power:         new(10.0),
			expectError:   false,
			expectedValue: new(7.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{Power: tt.power}
			err := sendiq.validatePower()

			assert.NoError(t, err)

			if tt.expectedValue != nil {
				require.NotNil(t, sendiq.Power)
				assert.Equal(t, *tt.expectedValue, *sendiq.Power)
			}
		})
	}
}

func TestSENDIQ_ValidateSharedMemToken(t *testing.T) {
	tests := []struct {
		name        string
		token       *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil token",
			token:       nil,
			expectError: false,
		},
		{
			name:        "valid positive token",
			token:       new(12345),
			expectError: false,
		},
		{
			name:        "valid negative token",
			token:       new(-1),
			expectError: false,
		},
		{
			name:        "zero token",
			token:       new(0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendiq := &SENDIQ{SharedMemToken: tt.token}
			err := sendiq.validateSharedMemToken()

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

func TestSENDIQ_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.iq")
	err := os.WriteFile(testFile, []byte("test"), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		sendiq      SENDIQ
		expectError bool
	}{
		{
			name: "valid minimal sendiq",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
			},
			expectError: false,
		},
		{
			name: "valid complete sendiq",
			sendiq: SENDIQ{
				InputFile:      testFile,
				Freq:           434000000.0,
				SampleRate:     new(96000),
				Harmonic:       new(2),
				IQType:         new("float"),
				Power:          new(2.5),
				SharedMemToken: new(12345),
				LoopMode:       true,
			},
			expectError: false,
		},
		{
			name: "invalid - missing input file",
			sendiq: SENDIQ{
				Freq: 434000000.0,
			},
			expectError: true,
		},
		{
			name: "invalid - missing frequency",
			sendiq: SENDIQ{
				InputFile: testFile,
			},
			expectError: true,
		},
		{
			name: "invalid - non-existent file",
			sendiq: SENDIQ{
				InputFile: "/nonexistent/file.iq",
				Freq:      434000000.0,
			},
			expectError: true,
		},
		{
			name: "invalid - invalid IQ type",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				IQType:    new("invalid"),
			},
			expectError: true,
		},
		{
			name: "invalid - sample rate too low",
			sendiq: SENDIQ{
				InputFile:  testFile,
				Freq:       434000000.0,
				SampleRate: new(5000),
			},
			expectError: true,
		},
		{
			name: "invalid - negative harmonic",
			sendiq: SENDIQ{
				InputFile: testFile,
				Freq:      434000000.0,
				Harmonic:  new(-1),
			},
			expectError: true,
		},
		{
			name: "invalid - zero shared memory token",
			sendiq: SENDIQ{
				InputFile:      testFile,
				Freq:           434000000.0,
				SharedMemToken: new(0),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sendiq.validate()

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
