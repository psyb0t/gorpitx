package gorpitx

const (
	kHzToMHzDivisor  = 1000.0 // conversion factor from kHz to MHz
	roundingOffset   = 0.5    // rounding offset for precision check
	decimalPrecision = 10.0   // for 1 decimal place precision check
)

// kHzToMHz converts frequency from kilohertz to megahertz.
func kHzToMHz(kHz float64) float64 {
	return kHz / kHzToMHzDivisor
}

// mHzToKHz converts frequency from megahertz to kilohertz.
func mHzToKHz(mHz float64) float64 {
	return mHz * kHzToMHzDivisor
}

// getMinFreqMHz returns the minimum supported frequency in MHz.
func getMinFreqMHz() float64 {
	return kHzToMHz(float64(minFreqKHz))
}

// getMaxFreqMHz returns the maximum supported frequency in MHz.
func getMaxFreqMHz() float64 {
	return kHzToMHz(float64(maxFreqKHz))
}

// isValidFreqMHz checks if a frequency in MHz is within RPiTX hardware limits.
func isValidFreqMHz(freqMHz float64) bool {
	return freqMHz >= getMinFreqMHz() && freqMHz <= getMaxFreqMHz()
}

// hasValidFreqPrecision checks if frequency has acceptable precision.
// pifmrds works best with 1 decimal place (0.1 MHz precision).
func hasValidFreqPrecision(freqMHz float64) bool {
	// Round to 1 decimal place and compare
	rounded := float64(int(freqMHz*decimalPrecision+roundingOffset)) / decimalPrecision

	return freqMHz == rounded
}
