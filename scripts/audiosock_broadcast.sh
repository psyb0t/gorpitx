#!/bin/bash

# AudioSock Broadcast Script
# Reads audio from unix socket and transmits via rpitx with CSDR presets
# Usage: ./audiosock_broadcast.sh <frequency_hz> <unix_socket_path> <sample_rate> <csdr_preset> <gain>

# Configuration
FREQUENCY="${1:-144500000}"  # Default 144.5 MHz
SOCKET_PATH="${2:-/tmp/audio_socket}"
SAMPLE_RATE="${3:-48000}"
CSDR_PRESET="${4:-FM_NARROW}"  # Default narrow FM preset
GAIN="${5:-1.0}"  # Default gain
LOG_FILE="/tmp/audiosock_broadcast.log"

# Function to log events
log_event() {
    echo "$(date '+%Y-%m-%d %H:%M:%S'): $1" | tee -a "$LOG_FILE"
}

# Cleanup function
cleanup() {
    log_event "Cleaning up AudioSock broadcast..."
    pkill -f "sendiq"
    pkill -f "csdr"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Check if socket exists
if [ ! -S "$SOCKET_PATH" ]; then
    log_event "ERROR: Unix socket $SOCKET_PATH does not exist"
    exit 1
fi

# Check if rpitx sendiq exists
SENDIQ_PATH="./sendiq"
if [ -n "$RPITX_PATH" ]; then
    SENDIQ_PATH="$RPITX_PATH/sendiq"
fi

if [ ! -f "$SENDIQ_PATH" ]; then
    log_event "ERROR: sendiq not found at $SENDIQ_PATH"
    exit 1
fi

log_event "Starting AudioSock broadcast on $FREQUENCY Hz from socket $SOCKET_PATH"
log_event "Sample rate: $SAMPLE_RATE Hz"
log_event "CSDR preset: $CSDR_PRESET"
log_event "Gain: $GAIN"
log_event "Using sendiq path: $SENDIQ_PATH"

# Main AudioSock transmission pipeline using csdr presets
log_event "Using CSDR preset: $CSDR_PRESET with gain $GAIN"
log_event "Full command: socat UNIX-CONNECT:$SOCKET_PATH STDOUT | csdr_presets.sh $CSDR_PRESET $GAIN | $SENDIQ_PATH -i /dev/stdin -s $SAMPLE_RATE -f $FREQUENCY -t float"

# Use csdr_presets.sh from same tmp directory
CSDR_PRESETS_PATH="/tmp/csdr_presets.sh"

socat UNIX-CONNECT:"$SOCKET_PATH" STDOUT | \
"$CSDR_PRESETS_PATH" "$CSDR_PRESET" "$GAIN" | \
"$SENDIQ_PATH" -i /dev/stdin -s "$SAMPLE_RATE" -f "$FREQUENCY" -t float

# Filter params explanation for bandpass_fir_fft_cc:
# 0.004 = low cutoff (0.4% of 48k = ~192Hz) - removes carrier and below
# 0.12 = high cutoff (12% of 48k = ~5.76kHz) - voice bandwidth limit
# 0.02 = transition bandwidth (2% of 48k = ~960Hz) - filter rolloff steepness

# Default pipeline explanation:
# Raw audio -> Complex signal -> DSB -> Filter to USB -> AGC -> RF out
# Result: Upper Side Band transmission (configurable via DSPPipeline parameter)

log_event "AudioSock broadcast ended"
