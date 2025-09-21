package gorpitx

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSK_ParseArgs_Success(t *testing.T) {
	tests := []struct {
		name          string
		input         FSK
		expectedArgs  []string
		expectedStdin bool
	}{
		{
			name: "text input with default baud rate",
			input: FSK{
				InputType: InputTypeText,
				Text:      "HELLO WORLD",
				Frequency: 431000000.0,
			},
			expectedArgs:  []string{"50", "431000000"},
			expectedStdin: true,
		},
		{
			name: "text input with custom baud rate",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST MESSAGE",
				BaudRate:  intPtr(100),
				Frequency: 434000000.0,
			},
			expectedArgs:  []string{"100", "434000000"},
			expectedStdin: true,
		},
		{
			name: "file input with default baud rate",
			input: FSK{
				InputType: InputTypeFile,
				File:      ".fixtures/test.txt",
				Frequency: 144500000.0,
			},
			expectedArgs:  []string{"50", "144500000"},
			expectedStdin: true,
		},
		{
			name: "file input with custom baud rate",
			input: FSK{
				InputType: InputTypeFile,
				File:      ".fixtures/test.txt",
				BaudRate:  intPtr(300),
				Frequency: 28070000.0,
			},
			expectedArgs:  []string{"300", "28070000"},
			expectedStdin: true,
		},
	}

	// Create test file for file input tests
	testFile := ".fixtures/test.txt"
	err := os.MkdirAll(".fixtures", 0o750)
	require.NoError(t, err)
	err = os.WriteFile(testFile, []byte("test file content"), 0o600)
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(testFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, stdin, err := tt.input.ParseArgs(inputBytes)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedArgs, args)

			if tt.expectedStdin {
				assert.NotNil(t, stdin)
				// Test that we can read from stdin
				data, err := io.ReadAll(stdin)
				require.NoError(t, err)

				if tt.input.InputType == InputTypeText {
					assert.Equal(t, tt.input.Text, string(data))
				} else {
					assert.Equal(t, "test file content", string(data))
				}
			} else {
				assert.Nil(t, stdin)
			}
		})
	}
}

func TestFSK_ParseArgs_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         FSK
		expectedError string
	}{
		{
			name: "missing input type",
			input: FSK{
				Text:      "TEST",
				Frequency: 431000000.0,
			},
			expectedError: "inputType",
		},
		{
			name: "invalid input type",
			input: FSK{
				InputType: "invalid",
				Text:      "TEST",
				Frequency: 431000000.0,
			},
			expectedError: "inputType must be 'file' or 'text'",
		},
		{
			name: "missing text for text input",
			input: FSK{
				InputType: InputTypeText,
				Frequency: 431000000.0,
			},
			expectedError: "text",
		},
		{
			name: "empty text for text input",
			input: FSK{
				InputType: InputTypeText,
				Text:      "",
				Frequency: 431000000.0,
			},
			expectedError: "text",
		},
		{
			name: "missing file for file input",
			input: FSK{
				InputType: InputTypeFile,
				Frequency: 431000000.0,
			},
			expectedError: "file",
		},
		{
			name: "empty file for file input",
			input: FSK{
				InputType: InputTypeFile,
				File:      "",
				Frequency: 431000000.0,
			},
			expectedError: "file",
		},
		{
			name: "non-existent file",
			input: FSK{
				InputType: InputTypeFile,
				File:      "/non/existent/file.txt",
				Frequency: 431000000.0,
			},
			expectedError: "file not found",
		},
		{
			name: "negative baud rate",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
				BaudRate:  intPtr(-50),
				Frequency: 431000000.0,
			},
			expectedError: "baud rate must be positive",
		},
		{
			name: "zero baud rate",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
				BaudRate:  intPtr(0),
				Frequency: 431000000.0,
			},
			expectedError: "baud rate must be positive",
		},
		{
			name: "missing frequency",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "negative frequency",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
				Frequency: -431000000.0,
			},
			expectedError: "frequency must be positive",
		},
		{
			name: "frequency too low",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
				Frequency: 1000.0, // 1 kHz - below 5 kHz limit
			},
			expectedError: "frequency out of RPiTX range",
		},
		{
			name: "frequency too high",
			input: FSK{
				InputType: InputTypeText,
				Text:      "TEST",
				Frequency: 2000000000.0, // 2 GHz - exceeds 1500 MHz limit
			},
			expectedError: "frequency out of RPiTX range",
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

func TestFSK_ParseArgs_JSONUnmarshalError(t *testing.T) {
	fsk := &FSK{}
	invalidJSON := []byte(`{"frequency": "invalid"}`)

	_, _, err := fsk.ParseArgs(invalidJSON)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal args")
}

func TestFSK_validateInputType(t *testing.T) {
	tests := []struct {
		name        string
		inputType   InputType
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid text input type",
			inputType:   InputTypeText,
			expectError: false,
		},
		{
			name:        "valid file input type",
			inputType:   InputTypeFile,
			expectError: false,
		},
		{
			name:        "empty input type",
			inputType:   "",
			expectError: true,
			errorMsg:    "inputType",
		},
		{
			name:        "invalid input type",
			inputType:   "invalid",
			expectError: true,
			errorMsg:    "inputType must be 'file' or 'text'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsk := &FSK{InputType: tt.inputType}
			err := fsk.validateInputType()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFSK_validateBaudRate(t *testing.T) {
	tests := []struct {
		name        string
		baudRate    *int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid baud rate",
			baudRate:    intPtr(50),
			expectError: false,
		},
		{
			name:        "nil baud rate (default)",
			baudRate:    nil,
			expectError: false,
		},
		{
			name:        "high baud rate",
			baudRate:    intPtr(9600),
			expectError: false,
		},
		{
			name:        "zero baud rate",
			baudRate:    intPtr(0),
			expectError: true,
			errorMsg:    "baud rate must be positive",
		},
		{
			name:        "negative baud rate",
			baudRate:    intPtr(-50),
			expectError: true,
			errorMsg:    "baud rate must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsk := &FSK{BaudRate: tt.baudRate}
			err := fsk.validateBaudRate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFSK_validateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, FrequencyValidationTest{
		name:        "valid frequency - 431 MHz",
		frequency:   431000000.0,
		expectError: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsk := &FSK{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, fsk.validateFrequency, tt)
		})
	}
}

func TestFSK_buildArgs(t *testing.T) {
	tests := []struct {
		name         string
		fsk          FSK
		expectedArgs []string
	}{
		{
			name: "default baud rate",
			fsk: FSK{
				Frequency: 431000000.0,
			},
			expectedArgs: []string{"50", "431000000"},
		},
		{
			name: "custom baud rate",
			fsk: FSK{
				BaudRate:  intPtr(300),
				Frequency: 144500000.0,
			},
			expectedArgs: []string{"300", "144500000"},
		},
		{
			name: "high frequency",
			fsk: FSK{
				BaudRate:  intPtr(1200),
				Frequency: 1296000000.0,
			},
			expectedArgs: []string{"1200", "1296000000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.fsk.buildArgs()
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestFSK_prepareStdin(t *testing.T) {
	// Create test file
	testFile := ".fixtures/stdin_test.txt"
	testContent := "test stdin content"
	err := os.MkdirAll(".fixtures", 0o750)
	require.NoError(t, err)
	err = os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(testFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	tests := []struct {
		name        string
		fsk         FSK
		expectError bool
		errorMsg    string
	}{
		{
			name: "text input",
			fsk: FSK{
				InputType: InputTypeText,
				Text:      "hello world",
			},
			expectError: false,
		},
		{
			name: "file input",
			fsk: FSK{
				InputType: InputTypeFile,
				File:      testFile,
			},
			expectError: false,
		},
		{
			name: "invalid input type",
			fsk: FSK{
				InputType: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid input type",
		},
		{
			name: "non-existent file",
			fsk: FSK{
				InputType: InputTypeFile,
				File:      "/non/existent/file.txt",
			},
			expectError: true,
			errorMsg:    "failed to open file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin, err := tt.fsk.prepareStdin()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, stdin)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, stdin)

				// Test reading from stdin
				data, err := io.ReadAll(stdin)
				require.NoError(t, err)

				if tt.fsk.InputType == InputTypeText {
					assert.Equal(t, tt.fsk.Text, string(data))
				} else {
					assert.Equal(t, testContent, string(data))
				}
			}
		})
	}
}
