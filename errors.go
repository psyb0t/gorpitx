package gorpitx

import (
	"errors"
)

// Module execution errors.
var (
	ErrUnknownModule = errors.New("unknown module")
	ErrExecuting     = errors.New("RPITX is busy executing another command")
	ErrNotExecuting  = errors.New("RPITX is not executing a command")
)

// Frequency validation errors.
var (
	ErrFreqNotSet     = errors.New("frequency is required")
	ErrFreqNegative   = errors.New("frequency must be positive")
	ErrFreqOutOfRange = errors.New("frequency out of RPiTX range")
	ErrFreqPrecision  = errors.New("frequency precision too high")
)

// Audio validation errors.
var (
	ErrAudioRequired = errors.New("audio file is required")
	ErrAudioNotFound = errors.New("audio file does not exist")
)

// PI code validation errors.
var (
	ErrPIInvalidLength = errors.New("PI code must be exactly 4 characters")
	ErrPIInvalidHex    = errors.New("PI code must be valid hex")
)

// PS validation errors.
var (
	ErrPSTooLong = errors.New("PS text must be 8 characters or less")
	ErrPSEmpty   = errors.New("PS text cannot be empty when specified")
)

// RT validation errors.
var (
	ErrRTTooLong = errors.New("RT text must be 64 characters or less")
)

// Control pipe validation errors.
var (
	ErrControlPipeEmpty    = errors.New("control pipe path cannot be empty when specified")
	ErrControlPipeNotFound = errors.New("control pipe does not exist")
)
