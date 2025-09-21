package gorpitx

import (
	"encoding/json"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPISSTVModule_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid minimal args",
			input: map[string]any{
				"pictureFile": ".fixtures/test_320x100.rgb",
				"frequency":   144500000.0, // 144.5 MHz in Hz
			},
			expectError: false,
			expectArgs:  []string{".fixtures/test_320x100.rgb", "144500000"},
		},
		{
			name: "valid different frequency",
			input: map[string]any{
				"pictureFile": ".fixtures/sstv_image.rgb",
				"frequency":   434000000.0, // 434 MHz in Hz
			},
			expectError: false,
			expectArgs:  []string{".fixtures/sstv_image.rgb", "434000000"},
		},
		{
			name: "valid high frequency",
			input: map[string]any{
				"pictureFile": ".fixtures/martin1.rgb",
				"frequency":   1296000000.0, // 1296 MHz in Hz
			},
			expectError: false,
			expectArgs:  []string{".fixtures/martin1.rgb", "1296000000"},
		},
		{
			name: "missing picture file",
			input: map[string]any{
				"frequency": 144500000.0,
			},
			expectError: true,
		},
		{
			name: "missing frequency",
			input: map[string]any{
				"pictureFile": ".fixtures/test.rgb",
			},
			expectError: true,
		},
		{
			name: "empty picture file",
			input: map[string]any{
				"pictureFile": "",
				"frequency":   144500000.0,
			},
			expectError: true,
		},
		{
			name: "zero frequency",
			input: map[string]any{
				"pictureFile": ".fixtures/test.rgb",
				"frequency":   0.0,
			},
			expectError: true,
		},
		{
			name: "negative frequency",
			input: map[string]any{
				"pictureFile": ".fixtures/test.rgb",
				"frequency":   -144500000.0,
			},
			expectError: true,
		},
		{
			name: "frequency too low",
			input: map[string]any{
				"pictureFile": ".fixtures/test.rgb",
				"frequency":   1000.0, // 1 kHz - below minimum
			},
			expectError: true,
		},
		{
			name: "frequency too high",
			input: map[string]any{
				"pictureFile": ".fixtures/test.rgb",
				"frequency":   2000000000.0, // 2 GHz - above maximum
			},
			expectError: true,
		},
		{
			name: "non-existent picture file",
			input: map[string]any{
				"pictureFile": "./nonexistent.rgb",
				"frequency":   144500000.0,
			},
			expectError: true,
		},
		{
			name: "invalid json",
			input: map[string]any{
				"pictureFile": 12345, // should be string
				"frequency":   144500000.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pisstv := &PISSTV{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := pisstv.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPISSTVModule_BuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		pisstv     PISSTV
		expectArgs []string
	}{
		{
			name: "basic SSTV transmission",
			pisstv: PISSTV{
				PictureFile: ".fixtures/test_320x100.rgb",
				Frequency:   144500000.0,
			},
			expectArgs: []string{".fixtures/test_320x100.rgb", "144500000"},
		},
		{
			name: "different image and frequency",
			pisstv: PISSTV{
				PictureFile: ".fixtures/sstv_martin1.rgb",
				Frequency:   434000000.0,
			},
			expectArgs: []string{".fixtures/sstv_martin1.rgb", "434000000"},
		},
		{
			name: "high frequency transmission",
			pisstv: PISSTV{
				PictureFile: ".fixtures/big_image_320x256.rgb",
				Frequency:   1296000000.0,
			},
			expectArgs: []string{".fixtures/big_image_320x256.rgb", "1296000000"},
		},
		{
			name: "absolute path",
			pisstv: PISSTV{
				PictureFile: "/tmp/sstv_test.rgb",
				Frequency:   28074000.0,
			},
			expectArgs: []string{"/tmp/sstv_test.rgb", "28074000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.pisstv.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPISSTVModule_ValidatePictureFile(t *testing.T) {
	tests := []struct {
		name        string
		pictureFile string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid picture file",
			pictureFile: ".fixtures/test_320x100.rgb",
			expectError: false,
		},
		{
			name:        "another valid file",
			pictureFile: ".fixtures/sstv_image.rgb",
			expectError: false,
		},
		{
			name:        "empty picture file",
			pictureFile: "",
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "non-existent file",
			pictureFile: "./does_not_exist.rgb",
			expectError: true,
			errorType:   commonerrors.ErrFileNotFound,
		},
		{
			name:        "absolute path non-existent",
			pictureFile: "/tmp/missing_file.rgb",
			expectError: true,
			errorType:   commonerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pisstv := &PISSTV{PictureFile: tt.pictureFile}
			err := pisstv.validatePictureFile()

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

func TestPISSTVModule_ValidateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, []FrequencyValidationTest{
		{
			name:        "valid frequency 144.5 MHz",
			frequency:   144500000.0,
			expectError: false,
		},
		{
			name:        "valid frequency 434 MHz",
			frequency:   434000000.0,
			expectError: false,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pisstv := &PISSTV{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, pisstv.validateFrequency, tt)
		})
	}
}

func TestPISSTVModule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		pisstv      PISSTV
		expectError bool
	}{
		{
			name: "valid complete PISSTV",
			pisstv: PISSTV{
				PictureFile: ".fixtures/test_320x100.rgb",
				Frequency:   144500000.0,
			},
			expectError: false,
		},
		{
			name: "valid different params",
			pisstv: PISSTV{
				PictureFile: ".fixtures/sstv_image.rgb",
				Frequency:   434000000.0,
			},
			expectError: false,
		},
		{
			name: "invalid - empty picture file",
			pisstv: PISSTV{
				PictureFile: "",
				Frequency:   144500000.0,
			},
			expectError: true,
		},
		{
			name: "invalid - zero frequency",
			pisstv: PISSTV{
				PictureFile: ".fixtures/test.rgb",
				Frequency:   0.0,
			},
			expectError: true,
		},
		{
			name: "invalid - non-existent file",
			pisstv: PISSTV{
				PictureFile: "./missing.rgb",
				Frequency:   144500000.0,
			},
			expectError: true,
		},
		{
			name: "invalid - frequency too high",
			pisstv: PISSTV{
				PictureFile: ".fixtures/test.rgb",
				Frequency:   2000000000.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pisstv.validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
