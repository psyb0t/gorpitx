package gorpitx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHzToMHz(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"convert 434000000 Hz to MHz", 434000000, 434.0},
		{"convert 107900000 Hz to MHz", 107900000, 107.9},
		{"convert 88500000 Hz to MHz", 88500000, 88.5},
		{"convert 1000000 Hz to MHz", 1000000, 1.0},
		{"convert 0 Hz to MHz", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hzToMHz(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKHzToMHz(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"convert 107900 kHz to MHz", 107900, 107.9},
		{"convert 88500 kHz to MHz", 88500, 88.5},
		{"convert 1000 kHz to MHz", 1000, 1.0},
		{"convert 0 kHz to MHz", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kHzToMHz(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMHzToHz(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"convert 107.9 MHz to Hz", 107.9, 107900000},
		{"convert 88.5 MHz to Hz", 88.5, 88500000},
		{"convert 434.0 MHz to Hz", 434.0, 434000000},
		{"convert 1.0 MHz to Hz", 1.0, 1000000},
		{"convert 0 MHz to Hz", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mHzToHz(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMHzToKHz(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"convert 107.9 MHz to kHz", 107.9, 107900},
		{"convert 88.5 MHz to kHz", 88.5, 88500},
		{"convert 1.0 MHz to kHz", 1.0, 1000},
		{"convert 0 MHz to kHz", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mHzToKHz(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMinFreqHz(t *testing.T) {
	result := getMinFreqHz()
	expected := float64(minFreqKHz * 1000) // Convert kHz to Hz
	assert.Equal(t, expected, result)
	assert.Equal(t, 5000.0, result) // 5 kHz = 5000 Hz
}

func TestGetMaxFreqHz(t *testing.T) {
	result := getMaxFreqHz()
	expected := float64(maxFreqKHz * 1000) // Convert kHz to Hz
	assert.Equal(t, expected, result)
	assert.Equal(t, 1500000000.0, result) // 1500000 kHz = 1500000000 Hz (1.5 GHz)
}

func TestGetMinFreqMHzDisplay(t *testing.T) {
	result := getMinFreqMHzDisplay()
	expected := kHzToMHz(float64(minFreqKHz))
	assert.Equal(t, expected, result)
	assert.Equal(t, 0.005, result) // 5 kHz = 0.005 MHz
}

func TestGetMaxFreqMHzDisplay(t *testing.T) {
	result := getMaxFreqMHzDisplay()
	expected := kHzToMHz(float64(maxFreqKHz))
	assert.Equal(t, expected, result)
	assert.Equal(t, 1500.0, result) // 1500000 kHz = 1500 MHz
}

func TestIsValidFreqHz(t *testing.T) {
	tests := []struct {
		name     string
		freq     float64
		expected bool
	}{
		{"valid frequency 107.9 MHz in Hz", 107900000, true},
		{"valid frequency at min", 5000, true},
		{"valid frequency at max", 1500000000, true},
		{"frequency too low", 1000, false},
		{"frequency too high", 2000000000, false},
		{"zero frequency", 0, false},
		{"negative frequency", -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFreqHz(tt.freq)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasValidFreqPrecision(t *testing.T) {
	tests := []struct {
		name     string
		freq     float64
		expected bool
	}{
		{"valid precision 107.9", 107.9, true},
		{"valid precision 88.5", 88.5, true},
		{"valid precision 100.0", 100.0, true},
		{"valid precision 99.1", 99.1, true},
		{"invalid precision 107.95", 107.95, false},
		{"invalid precision 88.55", 88.55, false},
		{"invalid precision 100.01", 100.01, false},
		{"invalid precision 99.99", 99.99, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasValidFreqPrecision(tt.freq)
			assert.Equal(t, tt.expected, result, "freq: %f", tt.freq)
		})
	}
}

func TestFrequencyConversionRoundTrip(t *testing.T) {
	// Test that converting kHz -> MHz -> kHz returns original value - math better fucking work
	originalKHz := 107900.0

	mhz := kHzToMHz(originalKHz)
	backToKHz := mHzToKHz(mhz)

	assert.Equal(t, originalKHz, backToKHz)

	// Test Hz -> MHz conversion for common frequencies
	originalHz := 434000000.0 // 434 MHz
	mhzFromHz := hzToMHz(originalHz)
	assert.Equal(t, 434.0, mhzFromHz)

	// Test MHz -> Hz -> MHz round trip
	originalMHz := 107.9
	hzFromMHz := mHzToHz(originalMHz)
	backToMHz := hzToMHz(hzFromMHz)
	assert.Equal(t, originalMHz, backToMHz)
}

func TestFrequencyConstants(t *testing.T) {
	// Test that our constants make sense - if these are wrong we're all fucked
	assert.Equal(t, 1000000.0, hzToMhzDivisor)
	assert.Equal(t, 1000.0, kHzToMHzDivisor)
	assert.Equal(t, 1000.0, khzToHzMultiplier)
	assert.Equal(t, 0.5, roundingOffset)
	assert.Equal(t, 10.0, decimalPrecision)

	// Test frequency range constants - stay within legal limits, dipshits
	assert.Equal(t, 5, minFreqKHz)
	assert.Equal(t, 1500000, maxFreqKHz)

	// Test that max > min - basic fucking logic
	assert.Greater(t, maxFreqKHz, minFreqKHz)
}
