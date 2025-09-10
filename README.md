# gorpitx

🚀 **Go wrapper that fucking executes rpitx modules without the bullshit.**

Tired of wrestling with raw C binaries like a goddamn caveman? This badass Go interface wraps rpitx so you can transmit radio signals without losing your shit. Singleton pattern because we're not animals, and proper process management because segfaults are for scrubs.

## 📡 What This Bastard Does

Executes rpitx modules through Go without the usual clusterfuck of manual process wrangling. Supports dev mode (fake transmission for testing) and production mode (actual RF carnage).

**Current Module:**
- **pifmrds**: FM broadcasting with RDS data (frequency in MHz) ✅ Done and dusted

**Architecture Highlights:**
- Singleton pattern with `GetInstance()` because global state done right
- Module interface for adding more transmission types without breaking shit  
- Process management with timeout and graceful stop (no zombie apocalypse)
- Dev mode with mock execution (test without frying your neighbors' electronics)
- Production mode requires root privileges (because RF transmission isn't a joke)

## ⚡ Quick Start (Stop Reading, Start Transmitting)

```bash
go get github.com/psyb0t/gorpitx
```

```go
package main

import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

func main() {
    // Get the singleton instance (there can be only one)
    rpitx := gorpitx.GetInstance()
    
    // Configure PIFMRDS module (FM with RDS, fancy shit)
    args := map[string]interface{}{
        "freq":  107.9,  // MHz - pick a frequency, any frequency
        "audio": "/path/to/audio.wav",  // Your audio masterpiece
        "pi":    "1234",  // Station ID (4 hex digits)
        "ps":    "BADASS",  // Station name (8 chars max)
        "rt":    "Broadcasting from Go like a boss!",
    }
    
    argsJSON, _ := json.Marshal(args)
    ctx := context.Background()
    
    // Execute with timeout (because infinite loops are evil)
    err := rpitx.Exec(ctx, gorpitx.ModuleNamePIFMRDS, argsJSON, 5*time.Minute)
    if err != nil {
        panic(err)  // Shit hit the fan
    }
}
```

## 🔧 Installation Requirements (Don't Skip This Shit)

**Hardware**: Raspberry Pi with GPIO access (Pi Zero, Pi Zero W, Pi A+, Pi B+, Pi 2B, Pi 3B, Pi 3B+)  
**OS**: Raspbian/Raspberry Pi OS (anything else is asking for trouble)  
**Dependencies**: rpitx (install this beast first or nothing works)  
**Privileges**: Must run as root in production (sudo your way to glory)

### Install rpitx (The Foundation of Everything)
```bash
# On your Pi, do this shit:
sudo apt update
git clone https://github.com/F5OEO/rpitx.git
cd rpitx
chmod +x install.sh
sudo ./install.sh  # This might take a hot minute
```

### Configure Path (Optional But Smart)
```bash
# Set rpitx binary path if you're not using defaults
export GORPITX_PATH="/home/pi/rpitx"
```

## 📋 PIFMRDS Module Configuration

```go
type PIFMRDS struct {
    Freq        float64  // Frequency in MHz (required, 0.005-1500 MHz)
    Audio       string   // Audio file path (required, must exist)
    PI          string   // PI code - 4 hex digits (optional)
    PS          string   // Station name - max 8 chars (optional)
    RT          string   // Radio text - max 64 chars (optional)
    PPM         *float64 // Clock correction ppm (optional)
    ControlPipe *string  // Named pipe for runtime control (optional)
}
```

**Validation Rules:**
- `Freq`: Required, positive, within RPiTX range (5kHz-1500MHz), 0.1MHz precision
- `Audio`: Required, file must exist (no stdin support yet)
- `PI`: Exactly 4 hexadecimal characters if specified
- `PS`: Max 8 characters, cannot be empty/whitespace if specified
- `RT`: Max 64 characters
- `ControlPipe`: Must exist if specified (create with `mkfifo`)

## 🎛️ Process Control

### Stream Output
```go
stdout := make(chan string, 100)
stderr := make(chan string, 100)

// Start streaming (only works during execution)
rpitx.StreamOutputs(stdout, stderr)

go func() {
    for line := range stdout {
        fmt.Println("STDOUT:", line)
    }
}()
```

### Graceful Stop
```go
ctx := context.Background()
err := rpitx.Stop(ctx, 3*time.Second)
if err != nil {
    // Handle stop error
}
```

### Execution State
- Only one module can execute at a time
- `Exec()` blocks until completion or timeout
- Automatic cleanup on context cancellation
- Process termination with SIGTERM then SIGKILL

## ⚙️ Environment Configuration

### Development Mode
Set `ENV=dev` to enable mock execution:
```bash
ENV=dev go run main.go
```

Mock execution runs infinite loop printing status every second instead of actual RF transmission.

### Production Mode
Default mode requiring root privileges:
```bash
sudo go run main.go  # or deploy as root
```

Executes actual rpitx binaries with proper RF transmission.

## 🧪 Error Handling

**Module Errors:**
- `ErrUnknownModule`: Requested module not registered
- `ErrExecuting`: Another command already running
- `ErrNotExecuting`: No active execution for stop/stream

**Validation Errors:**
- `ErrFreqNotSet`, `ErrFreqNegative`, `ErrFreqOutOfRange`, `ErrFreqPrecision`
- `ErrAudioRequired`, `ErrAudioNotFound`
- `ErrPIInvalidLength`, `ErrPIInvalidHex`
- `ErrPSTooLong`, `ErrPSEmpty`
- `ErrRTTooLong`
- `ErrControlPipeEmpty`, `ErrControlPipeNotFound`

## 🔗 Architecture

### Module Interface
```go
type Module interface {
    ParseArgs(json.RawMessage) ([]string, error)
}
```

New modules implement this interface with:
1. JSON unmarshaling of configuration
2. Parameter validation
3. Command-line argument building

### Frequency Utilities
- `kHzToMHz(kHz float64) float64` - Convert kHz to MHz
- `mHzToKHz(mHz float64) float64` - Convert MHz to kHz  
- `isValidFreqMHz(freqMHz float64) bool` - Validate MHz frequency
- `hasValidFreqPrecision(freqMHz float64) bool` - Check 0.1MHz precision

**Note**: pifmrds uses MHz, other planned modules use Hz.

## 📋 TODO: Remaining Modules Implementation (The Fun Stuff)

Based on the easytest modules from rpitx, here are the **10 badass modules** we need to implement (excluding that legacy rpitx garbage):

- **TUNE** - Simple Carrier Generation
  - **Command**: `tune [-f Frequency] [-e] [-p ppm]`
  - **Go struct**:
    ```go
    type TUNE struct {
        Frequency *float64 `json:"frequency,omitempty"` // Hz, required, 50kHz-1500MHz
        ExitImmediate *bool `json:"exitImmediate,omitempty"` // Optional, don't kill carrier
        PPM *float64 `json:"ppm,omitempty"` // Optional, clock correction > 0
    }
    ```
  - **Validation**: Frequency > 0, PPM > 0 if provided

- **PICHIRP** - Frequency Sweep Generator
  - **Command**: `pichirp Frequency(Hz) Bandwidth(Hz) Time(Seconds)`
  - **Go struct**:
    ```go
    type PICHIRP struct {
        Frequency float64 `json:"frequency"` // Hz, required, center frequency
        Bandwidth float64 `json:"bandwidth"` // Hz, required, sweep bandwidth  
        Time float64 `json:"time"` // Seconds, required, sweep duration
    }
    ```
  - **Validation**: All parameters required and > 0

- **SPECTRUMPAINT** - Spectrum Painting
  - **Command**: `spectrumpaint picture.rgb frequency(Hz) [Excursion(Hz)]`
  - **Go struct**:
    ```go
    type SPECTRUMPAINT struct {
        PictureFile string `json:"pictureFile"` // Required, .rgb file path
        Frequency float64 `json:"frequency"` // Hz, required
        Excursion *float64 `json:"excursion,omitempty"` // Hz, optional, default 100000
    }
    ```
  - **Validation**: File exists, frequency > 0, excursion > 0 if provided

- **MORSE** - Morse Code Transmission
  - **Command**: `morse frequency(Hz) rate(dits) "message"`
  - **Go struct**:
    ```go
    type MORSE struct {
        Frequency float64 `json:"frequency"` // Hz, required, carrier frequency
        Rate int `json:"rate"` // Required, rate in dits per minute
        Message string `json:"message"` // Required, message text to transmit
    }
    ```
  - **Validation**: All required, frequency > 0, rate > 0, message not empty

- **SENDIQ** - IQ Data Transmission
  - **Command**: `sendiq [-i File] [-s Samplerate] [-f Frequency] [-l] [-h Harmonic] [-m Token] [-d] [-p Power] [-t IQType]`
  - **Go struct**:
    ```go
    type SENDIQ struct {
        InputFile string `json:"inputFile"` // Required, input file path
        SampleRate *int `json:"sampleRate,omitempty"` // Optional, 10000-250000, default 48000
        Frequency *float64 `json:"frequency,omitempty"` // Hz, optional, 50kHz-1500MHz, default 434e6
        LoopMode *bool `json:"loopMode,omitempty"` // Optional, default false
        Harmonic *int `json:"harmonic,omitempty"` // Optional, >= 1, default 1
        SharedMemoryToken *int `json:"sharedMemoryToken,omitempty"` // Optional
        DDSMode *bool `json:"ddsMode,omitempty"` // Optional, default false
        PowerLevel *float64 `json:"powerLevel,omitempty"` // Optional, 0.0-7.0, default 0.1
        IQType *string `json:"iqType,omitempty"` // Optional, i16/u8/float/double, default "i16"
    }
    ```
  - **Validation**: File exists, sample rate 10000-250000, power 0.0-7.0, IQ type enum

- **FREEDV** - FreeDV Digital Voice
  - **Command**: `freedv vco.rf frequency(Hz) [samplerate(Hz)]`
  - **Go struct**:
    ```go
    type FREEDV struct {
        VCOFile string `json:"vcoFile"` // Required, .rf file path
        Frequency float64 `json:"frequency"` // Hz, required
        SampleRate *int `json:"sampleRate,omitempty"` // Hz, optional, default 400
    }
    ```
  - **Validation**: File exists, frequency > 0, sample rate > 0 if provided

- **PISSTV** - Slow Scan TV
  - **Command**: `pisstv picture.rgb frequency(Hz)`
  - **Go struct**:
    ```go
    type PISSTV struct {
        PictureFile string `json:"pictureFile"` // Required, 320x256 .rgb file
        Frequency float64 `json:"frequency"` // Hz, required
    }
    ```
  - **Validation**: File exists, frequency > 0

- **POCSAG** - Pager Protocol
  - **Command**: `pocsag [-f Frequency] [-r Rate] [-b FunctionBits] [-n] [-t RepeatCount] [-i] [-d]`
  - **Go struct**:
    ```go
    type POCSAG struct {
        Frequency *float64 `json:"frequency,omitempty"` // Hz, optional, 50kHz-1500MHz, default 466230000
        BaudRate *int `json:"baudRate,omitempty"` // Optional, 512/1200/2400, default 1200
        FunctionBits *int `json:"functionBits,omitempty"` // Optional, 0-3, default 3
        NumericMode *bool `json:"numericMode,omitempty"` // Optional, default false
        RepeatCount *int `json:"repeatCount,omitempty"` // Optional, default 4
        InvertPolarity *bool `json:"invertPolarity,omitempty"` // Optional, default false
        Debug *bool `json:"debug,omitempty"` // Optional, default false
        Messages []POCSAGMessage `json:"messages"` // Required, address:message pairs
    }

    type POCSAGMessage struct {
        Address int `json:"address"` // Required, pager address
        Message string `json:"message"` // Required, message text
        FunctionBits *int `json:"functionBits,omitempty"` // Optional override
    }
    ```
  - **Validation**: Baud rate enum, function bits 0-3, at least one message

- **PIOPERA** - OPERA Protocol
  - **Command**: `piopera CALLSIGN OperaMode[0.5,1,2,4,8] frequency(Hz)`
  - **Go struct**:
    ```go
    type PIOPERA struct {
        Callsign string `json:"callsign"` // Required, amateur radio callsign
        Mode float64 `json:"mode"` // Required, 0.5/1/2/4/8
        Frequency float64 `json:"frequency"` // Hz, required
    }
    ```
  - **Validation**: Callsign format (3rd char numeric), mode enum, frequency > 0

- **PIRTTY** - RTTY Protocol
  - **Command**: `pirtty frequency(Hz) SpaceFrequency(Hz) text`
  - **Go struct**:
    ```go
    type PIRTTY struct {
        Frequency float64 `json:"frequency"` // Hz, required, carrier frequency
        SpaceFrequency int `json:"spaceFrequency"` // Hz, required, space tone
        Text string `json:"text"` // Required, message text
    }
    ```
  - **Validation**: All required, frequency > 0, space frequency > 0, text not empty

### Common Validation Functions Needed

```go
func ValidateFrequency(freq float64, min, max float64) error
func ValidateFileExists(path string) error  
func ValidateEnum(value string, allowedValues []string) error
func ValidateRange(value, min, max float64) error
```

## ⚠️ Legal Notice (Read This or Get Fucked by the FCC)

**RF transmission is regulated as hell.** Don't be a dickhead - get proper licensing before broadcasting. This software is for:
- Licensed amateur radio operators (you know who you are)
- Low-power experimentation in permitted bands (don't fry your neighbor's radio gear)
- Educational/research purposes (learn responsibly, you beautiful bastards)

**Absolutely NOT for**: Commercial broadcasting without authorization (the FCC will skull-fuck your wallet).

## 📚 Dependencies

- [`github.com/psyb0t/commander`](https://github.com/psyb0t/commander) - Process execution
- [`github.com/psyb0t/common-go/env`](https://github.com/psyb0t/common-go) - Environment detection  
- [`github.com/psyb0t/ctxerrors`](https://github.com/psyb0t/ctxerrors) - Context-aware errors
- [`github.com/psyb0t/gonfiguration`](https://github.com/psyb0t/gonfiguration) - Configuration parsing
- [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus) - Logging

## 📄 License

MIT License. Use responsibly and don't be a twat.

---

*Go interface for rpitx that doesn't suck. Built for radio enthusiasts who want clean code without the usual C library nightmare fuel.*