package gorpitx

import (
	"encoding/json"
	"os"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPECTRUMPAINT_ParseArgs(t *testing.T) {
	// Use fixture file for testing
	testFile := ".fixtures/test_spectrum_320x100.Y"

	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid complete args",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   434000000.0, // 434 MHz in Hz
				"excursion":   100000.0,    // 100 kHz
			},
			expectError: false,
			expectArgs:  []string{testFile, "434000000", "100000"},
		},
		{
			name: "valid args without excursion",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   144000000.0, // 144 MHz in Hz
			},
			expectError: false,
			expectArgs:  []string{testFile, "144000000"},
		},
		{
			name: "valid args different frequency",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   28000000.0, // 28 MHz in Hz
				"excursion":   50000.0,    // 50 kHz
			},
			expectError: false,
			expectArgs:  []string{testFile, "28000000", "50000"},
		},
		{
			name: "missing picture file",
			input: map[string]any{
				"frequency": 434000000.0,
				"excursion": 100000.0,
			},
			expectError: true,
		},
		{
			name: "empty picture file",
			input: map[string]any{
				"pictureFile": "",
				"frequency":   434000000.0,
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "nonexistent picture file",
			input: map[string]any{
				"pictureFile": "/nonexistent/file.rgb",
				"frequency":   434000000.0,
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"pictureFile": testFile,
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   0.0,
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   -434000000.0,
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   1000.0, // 1 kHz - below minimum
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   2000000000.0, // 2 GHz - above maximum
				"excursion":   100000.0,
			},
			expectError: true,
		},
		{
			name: "zero excursion",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   434000000.0,
				"excursion":   0.0,
			},
			expectError: true,
		},
		{
			name: "negative excursion",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   434000000.0,
				"excursion":   -100000.0,
			},
			expectError: true,
		},
		{
			name: "invalid json",
			input: map[string]any{
				"pictureFile": testFile,
				"frequency":   "not_a_number",
				"excursion":   100000.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectrum := &SPECTRUMPAINT{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := spectrum.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestSPECTRUMPAINT_BuildArgs(t *testing.T) {
	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_spectrum_*.rgb")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())
	testFile.Close()

	tests := []struct {
		name       string
		spectrum   SPECTRUMPAINT
		expectArgs []string
	}{
		{
			name: "complete spectrum paint",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   434000000.0,
				Excursion:   floatPtr(100000.0),
			},
			expectArgs: []string{testFile.Name(), "434000000", "100000"},
		},
		{
			name: "without excursion",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   144000000.0,
			},
			expectArgs: []string{testFile.Name(), "144000000"},
		},
		{
			name: "different frequency and excursion",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   28000000.0,
				Excursion:   floatPtr(50000.0),
			},
			expectArgs: []string{testFile.Name(), "28000000", "50000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.spectrum.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestSPECTRUMPAINT_ValidatePictureFile(t *testing.T) {
	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_spectrum_*.rgb")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())
	testFile.Close()

	tests := []struct {
		name        string
		pictureFile string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid existing file",
			pictureFile: testFile.Name(),
			expectError: false,
		},
		{
			name:        "empty picture file",
			pictureFile: "",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "nonexistent file",
			pictureFile: "/nonexistent/file.rgb",
			expectError: true,
			errorType:   commonerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectrum := &SPECTRUMPAINT{PictureFile: tt.pictureFile}
			err := spectrum.validatePictureFile()

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

func TestSPECTRUMPAINT_ValidateFrequency(t *testing.T) {
	tests := []struct {
		name        string
		frequency   float64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid frequency 434 MHz",
			frequency:   434000000.0,
			expectError: false,
		},
		{
			name:        "valid frequency 144 MHz",
			frequency:   144000000.0,
			expectError: false,
		},
		{
			name:        "valid minimum frequency",
			frequency:   50000.0, // 50 kHz
			expectError: false,
		},
		{
			name:        "valid maximum frequency",
			frequency:   1500000000.0, // 1500 MHz
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
			name:        "frequency too low",
			frequency:   1000.0, // 1 kHz
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
		{
			name:        "frequency too high",
			frequency:   2000000000.0, // 2 GHz
			expectError: true,
			errorType:   ErrFreqOutOfRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectrum := &SPECTRUMPAINT{Frequency: tt.frequency}
			err := spectrum.validateFrequency()

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

func TestSPECTRUMPAINT_ValidateExcursion(t *testing.T) {
	tests := []struct {
		name        string
		excursion   *float64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid excursion",
			excursion:   floatPtr(100000.0),
			expectError: false,
		},
		{
			name:        "nil excursion (optional)",
			excursion:   nil,
			expectError: false,
		},
		{
			name:        "zero excursion",
			excursion:   floatPtr(0.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "negative excursion",
			excursion:   floatPtr(-100000.0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectrum := &SPECTRUMPAINT{Excursion: tt.excursion}
			err := spectrum.validateExcursion()

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

func TestSPECTRUMPAINT_Validate(t *testing.T) {
	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_spectrum_*.rgb")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())
	testFile.Close()

	tests := []struct {
		name        string
		spectrum    SPECTRUMPAINT
		expectError bool
	}{
		{
			name: "valid complete spectrum",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   434000000.0,
				Excursion:   floatPtr(100000.0),
			},
			expectError: false,
		},
		{
			name: "valid without excursion",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   144000000.0,
			},
			expectError: false,
		},
		{
			name: "invalid - missing picture file",
			spectrum: SPECTRUMPAINT{
				PictureFile: "",
				Frequency:   434000000.0,
				Excursion:   floatPtr(100000.0),
			},
			expectError: true,
		},
		{
			name: "invalid - zero frequency",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   0.0,
				Excursion:   floatPtr(100000.0),
			},
			expectError: true,
		},
		{
			name: "invalid - negative excursion",
			spectrum: SPECTRUMPAINT{
				PictureFile: testFile.Name(),
				Frequency:   434000000.0,
				Excursion:   floatPtr(-100000.0),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spectrum.validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

