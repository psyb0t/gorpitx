package gorpitx

import (
	"encoding/json"
	"io"
	"testing"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPOCSAG_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectArgs  []string
	}{
		{
			name: "valid minimal args - single message",
			input: map[string]any{
				"frequency": 466230000,
				"messages": []map[string]any{
					{
						"address": 123456,
						"message": "Test message",
					},
				},
			},
			expectError: false,
			expectArgs:  []string{"-f", "466230000"},
		},
		{
			name: "valid complete args - all options",
			input: map[string]any{
				"frequency":      466230000.0,
				"baudRate":       1200,
				"functionBits":   3,
				"numericMode":    true,
				"repeatCount":    4,
				"invertPolarity": true,
				"debug":          true,
				"messages": []map[string]any{
					{
						"address": 123456,
						"message": "Alert message",
					},
				},
			},
			expectError: false,
			expectArgs: []string{
				"-f", "466230000", "-r", "1200", "-b", "3", "-n", "-t", "4", "-i", "-d",
			},
		},
		{
			name: "valid multiple messages",
			input: map[string]any{
				"frequency": 433920000.0,
				"baudRate":  512,
				"messages": []map[string]any{
					{
						"address": 100,
						"message": "First message",
					},
					{
						"address": 200,
						"message": "Second message",
					},
				},
			},
			expectError: false,
			expectArgs:  []string{"-f", "433920000", "-r", "512"},
		},
		{
			name: "valid with different baud rates",
			input: map[string]any{
				"frequency": 466230000,
				"baudRate":  2400,
				"messages": []map[string]any{
					{
						"address": 789,
						"message": "High speed message",
					},
				},
			},
			expectError: false,
			expectArgs:  []string{"-f", "466230000", "-r", "2400"},
		},
		{
			name: "valid with function bits variations",
			input: map[string]any{
				"frequency":    466230000,
				"functionBits": 0,
				"messages": []map[string]any{
					{
						"address": 555,
						"message": "Function 0 message",
					},
				},
			},
			expectError: false,
			expectArgs:  []string{"-f", "466230000", "-b", "0"},
		},
		{
			name: "valid with false flags",
			input: map[string]any{
				"frequency":      466230000,
				"numericMode":    false,
				"invertPolarity": false,
				"debug":          false,
				"messages": []map[string]any{
					{
						"address": 777,
						"message": "Normal message",
					},
				},
			},
			expectError: false,
			expectArgs:  []string{"-f", "466230000"},
		},
		{
			name: "missing messages",
			input: map[string]any{
				"frequency": 466230000.0,
			},
			expectError: true,
		},
		{
			name: "empty messages array",
			input: map[string]any{
				"messages": []map[string]any{},
			},
			expectError: true,
		},
		{
			name: "invalid frequency - negative",
			input: map[string]any{
				"frequency": -466230000.0,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid frequency - too low",
			input: map[string]any{
				"frequency": 1000.0, // 1 kHz - below minimum
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid frequency - too high",
			input: map[string]any{
				"frequency": 2000000000.0, // 2 GHz - above maximum
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid baud rate",
			input: map[string]any{
				"baudRate": 9600, // Not 512, 1200, or 2400
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid function bits - too low",
			input: map[string]any{
				"functionBits": -1,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid function bits - too high",
			input: map[string]any{
				"functionBits": 4,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid repeat count - zero",
			input: map[string]any{
				"repeatCount": 0,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid repeat count - negative",
			input: map[string]any{
				"repeatCount": -5,
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid message - negative address",
			input: map[string]any{
				"messages": []map[string]any{
					{
						"address": -123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid message - empty text",
			input: map[string]any{
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid message - whitespace only text",
			input: map[string]any{
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "   ",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid per-message function bits",
			input: map[string]any{
				"messages": []map[string]any{
					{
						"address":      123,
						"message":      "Test",
						"functionBits": 5,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid json - frequency as string",
			input: map[string]any{
				"frequency": "not_a_number",
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid json - baud rate as string",
			input: map[string]any{
				"baudRate": "not_a_number",
				"messages": []map[string]any{
					{
						"address": 123,
						"message": "Test",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{}
			inputBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			args, _, err := pocsag.ParseArgs(inputBytes)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPOCSAG_BuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		pocsag     POCSAG
		expectArgs []string
	}{
		{
			name: "minimal configuration",
			pocsag: POCSAG{
				Frequency: 466230000.0,
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectArgs: []string{"-f", "466230000"},
		},
		{
			name: "complete configuration",
			pocsag: POCSAG{
				Frequency:      466230000.0,
				BaudRate:       intPtr(1200),
				FunctionBits:   intPtr(3),
				NumericMode:    boolPtr(true),
				RepeatCount:    intPtr(4),
				InvertPolarity: boolPtr(true),
				Debug:          boolPtr(true),
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Alert message",
					},
				},
			},
			expectArgs: []string{
				"-f", "466230000", "-r", "1200", "-b", "3", "-n", "-t", "4", "-i", "-d",
			},
		},
		{
			name: "multiple messages",
			pocsag: POCSAG{
				Frequency: 433920000.0,
				Messages: []POCSAGMessage{
					{
						Address: 100,
						Message: "First message",
					},
					{
						Address: 200,
						Message: "Second message",
					},
					{
						Address: 300,
						Message: "Third message",
					},
				},
			},
			expectArgs: []string{"-f", "433920000"},
		},
		{
			name: "false flags not included",
			pocsag: POCSAG{
				Frequency:      466230000.0,
				NumericMode:    boolPtr(false),
				InvertPolarity: boolPtr(false),
				Debug:          boolPtr(false),
				Messages: []POCSAGMessage{
					{
						Address: 777,
						Message: "Normal message",
					},
				},
			},
			expectArgs: []string{"-f", "466230000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.pocsag.buildArgs()
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestPOCSAG_Stdin(t *testing.T) {
	tests := []struct {
		name          string
		pocsag        POCSAG
		expectedStdin string
	}{
		{
			name: "single message",
			pocsag: POCSAG{
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectedStdin: "123456:Test message",
		},
		{
			name: "multiple messages",
			pocsag: POCSAG{
				Messages: []POCSAGMessage{
					{
						Address: 100,
						Message: "First message",
					},
					{
						Address: 200,
						Message: "Second message",
					},
					{
						Address: 300,
						Message: "Third message",
					},
				},
			},
			expectedStdin: "100:First message\n200:Second message\n300:Third message",
		},
		{
			name: "message with special characters",
			pocsag: POCSAG{
				Messages: []POCSAGMessage{
					{
						Address: 777,
						Message: "Hello! @#$% World 123",
					},
				},
			},
			expectedStdin: "777:Hello! @#$% World 123",
		},
		{
			name: "zero address",
			pocsag: POCSAG{
				Messages: []POCSAGMessage{
					{
						Address: 0,
						Message: "Zero address message",
					},
				},
			},
			expectedStdin: "0:Zero address message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := tt.pocsag.buildStdin()

			// Read stdin content
			stdinBytes, err := io.ReadAll(stdin)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStdin, string(stdinBytes))
		})
	}
}

func TestPOCSAG_ParseArgs_Stdin(t *testing.T) {
	// Test that ParseArgs returns proper stdin content
	input := map[string]any{
		"frequency": 466230000.0,
		"messages": []map[string]any{
			{
				"address": 123,
				"message": "Hello POCSAG",
			},
			{
				"address": 456,
				"message": "Second message",
			},
		},
	}

	pocsag := &POCSAG{}
	inputBytes, err := json.Marshal(input)
	require.NoError(t, err)

	args, stdin, err := pocsag.ParseArgs(inputBytes)

	require.NoError(t, err)
	require.NotNil(t, stdin)
	// Frequency should be in args
	assert.Equal(t, []string{"-f", "466230000"}, args)

	// Verify stdin content
	stdinContent, err := io.ReadAll(stdin)
	require.NoError(t, err)
	assert.Equal(t, "123:Hello POCSAG\n456:Second message", string(stdinContent))
}

func TestPOCSAG_ValidateFrequency(t *testing.T) {
	tests := GetStandardFrequencyValidationTests()
	tests = append(tests, FrequencyValidationTest{
		name:        "valid frequency - 466.230 MHz",
		frequency:   466230000.0,
		expectError: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{Frequency: tt.frequency}
			RunFrequencyValidationTest(t, pocsag.validateFrequency, tt)
		})
	}
}

func TestPOCSAG_ValidateBaudRate(t *testing.T) {
	tests := []struct {
		name        string
		baudRate    *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil baud rate (optional)",
			baudRate:    nil,
			expectError: false,
		},
		{
			name:        "valid baud rate - 512",
			baudRate:    intPtr(512),
			expectError: false,
		},
		{
			name:        "valid baud rate - 1200",
			baudRate:    intPtr(1200),
			expectError: false,
		},
		{
			name:        "valid baud rate - 2400",
			baudRate:    intPtr(2400),
			expectError: false,
		},
		{
			name:        "invalid baud rate - 9600",
			baudRate:    intPtr(9600),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid baud rate - 300",
			baudRate:    intPtr(300),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid baud rate - 0",
			baudRate:    intPtr(0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid baud rate - negative",
			baudRate:    intPtr(-1200),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{BaudRate: tt.baudRate}
			err := pocsag.validateBaudRate()

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

func TestPOCSAG_ValidateFunctionBits(t *testing.T) {
	tests := []struct {
		name         string
		functionBits *int
		expectError  bool
		errorType    error
	}{
		{
			name:         "nil function bits (optional)",
			functionBits: nil,
			expectError:  false,
		},
		{
			name:         "valid function bits - 0",
			functionBits: intPtr(0),
			expectError:  false,
		},
		{
			name:         "valid function bits - 1",
			functionBits: intPtr(1),
			expectError:  false,
		},
		{
			name:         "valid function bits - 2",
			functionBits: intPtr(2),
			expectError:  false,
		},
		{
			name:         "valid function bits - 3",
			functionBits: intPtr(3),
			expectError:  false,
		},
		{
			name:         "invalid function bits - negative",
			functionBits: intPtr(-1),
			expectError:  true,
			errorType:    commonerrors.ErrInvalidValue,
		},
		{
			name:         "invalid function bits - too high",
			functionBits: intPtr(4),
			expectError:  true,
			errorType:    commonerrors.ErrInvalidValue,
		},
		{
			name:         "invalid function bits - way too high",
			functionBits: intPtr(10),
			expectError:  true,
			errorType:    commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{FunctionBits: tt.functionBits}
			err := pocsag.validateFunctionBits()

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

func TestPOCSAG_ValidateRepeatCount(t *testing.T) {
	tests := []struct {
		name        string
		repeatCount *int
		expectError bool
		errorType   error
	}{
		{
			name:        "nil repeat count (optional)",
			repeatCount: nil,
			expectError: false,
		},
		{
			name:        "valid repeat count - 1",
			repeatCount: intPtr(1),
			expectError: false,
		},
		{
			name:        "valid repeat count - 4",
			repeatCount: intPtr(4),
			expectError: false,
		},
		{
			name:        "valid repeat count - 10",
			repeatCount: intPtr(10),
			expectError: false,
		},
		{
			name:        "invalid repeat count - zero",
			repeatCount: intPtr(0),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name:        "invalid repeat count - negative",
			repeatCount: intPtr(-5),
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{RepeatCount: tt.repeatCount}
			err := pocsag.validateRepeatCount()

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

func TestPOCSAG_ValidateMessages(t *testing.T) {
	tests := []struct {
		name        string
		messages    []POCSAGMessage
		expectError bool
		errorType   error
	}{
		{
			name: "valid single message",
			messages: []POCSAGMessage{
				{
					Address: 123456,
					Message: "Test message",
				},
			},
			expectError: false,
		},
		{
			name: "valid multiple messages",
			messages: []POCSAGMessage{
				{
					Address: 100,
					Message: "First message",
				},
				{
					Address: 200,
					Message: "Second message",
				},
			},
			expectError: false,
		},
		{
			name: "valid message with per-message function bits",
			messages: []POCSAGMessage{
				{
					Address:      123456,
					Message:      "Test message",
					FunctionBits: intPtr(2),
				},
			},
			expectError: false,
		},
		{
			name:        "empty messages array",
			messages:    []POCSAGMessage{},
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name:        "nil messages array",
			messages:    nil,
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name: "invalid message - negative address",
			messages: []POCSAGMessage{
				{
					Address: -123,
					Message: "Test message",
				},
			},
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name: "invalid message - empty text",
			messages: []POCSAGMessage{
				{
					Address: 123,
					Message: "",
				},
			},
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name: "invalid message - whitespace only text",
			messages: []POCSAGMessage{
				{
					Address: 123,
					Message: "   ",
				},
			},
			expectError: true,
			errorType:   commonerrors.ErrRequiredFieldNotSet,
		},
		{
			name: "invalid per-message function bits - too high",
			messages: []POCSAGMessage{
				{
					Address:      123,
					Message:      "Test",
					FunctionBits: intPtr(4),
				},
			},
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
		{
			name: "invalid per-message function bits - negative",
			messages: []POCSAGMessage{
				{
					Address:      123,
					Message:      "Test",
					FunctionBits: intPtr(-1),
				},
			},
			expectError: true,
			errorType:   commonerrors.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pocsag := &POCSAG{Messages: tt.messages}
			err := pocsag.validateMessages()

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

func TestPOCSAG_Validate(t *testing.T) {
	tests := []struct {
		name        string
		pocsag      POCSAG
		expectError bool
	}{
		{
			name: "valid minimal configuration",
			pocsag: POCSAG{
				Frequency: 466230000.0,
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid complete configuration",
			pocsag: POCSAG{
				Frequency:      466230000.0,
				BaudRate:       intPtr(1200),
				FunctionBits:   intPtr(3),
				NumericMode:    boolPtr(true),
				RepeatCount:    intPtr(4),
				InvertPolarity: boolPtr(true),
				Debug:          boolPtr(true),
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Alert message",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid frequency",
			pocsag: POCSAG{
				Frequency: 0.0, // Invalid
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid baud rate",
			pocsag: POCSAG{
				Frequency: 466230000.0,
				BaudRate:  intPtr(9600), // Invalid
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid function bits",
			pocsag: POCSAG{
				Frequency:    466230000.0,
				FunctionBits: intPtr(5), // Invalid
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid repeat count",
			pocsag: POCSAG{
				Frequency:   466230000.0,
				RepeatCount: intPtr(0), // Invalid
				Messages: []POCSAGMessage{
					{
						Address: 123456,
						Message: "Test message",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid messages - empty array",
			pocsag: POCSAG{
				Frequency: 466230000.0,
				Messages:  []POCSAGMessage{}, // Invalid
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pocsag.validate()

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
