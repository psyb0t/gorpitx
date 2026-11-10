---
title: "sendiq"
description: "The gorpitx sendiq module: transmit raw I/Q samples from a file, with sample rate, I/Q format, power, loop mode and run-time control over shared memory."
nav_order: 12
parent: "Modules"
---

# sendiq: raw I/Q samples

Transmits I/Q samples from a file. Anything you can produce as I/Q, from GNU Radio, csdr or your own code, goes on the air this way. With a shared memory token, another program can change the frequency and power while it runs. Code: [`sendiq.go`](https://github.com/psyb0t/gorpitx/blob/main/sendiq.go).

## Fields

The frequency field is `freq` here, not `frequency`, and it is in Hz.

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `inputFile` | `InputFile` | `string` | yes | Path to an existing file, or `-` (see below) |
| `freq` | `Freq` | `float64` | yes | Hz, 5000 to 1500000000 |
| `sampleRate` | `SampleRate` | `*int` | no | 10000 to 2000000; the program defaults to 48000 |
| `harmonic` | `Harmonic` | `*int` | no | 1 or more; the program defaults to 1 |
| `iqType` | `IQType` | `*string` | no | `i16`, `u8`, `float` or `double` (`gorpitx.IQType*` constants); the program defaults to `i16` |
| `power` | `Power` | `*float64` | no | Drive level 0 to 7, clamped into that range rather than refused; the program defaults to 0.1 |
| `sharedMemToken` | `SharedMemToken` | `*int` | no | Non-zero shared memory key for run-time control |
| `loopMode` | `LoopMode` | `bool` | no | `true` adds `-l` and restarts the file at its end |

The package exports the program's defaults as `DefaultSampleRate`, `DefaultHarmonic`, `DefaultIQType` and `DefaultPower`. gorpitx does not send them; it leaves a flag out when you leave its field out.

`sampleRate` above 200000 is reduced by `sendiq` itself: it divides by the smallest factor from 2 to 11 that brings the rate under 200000 and decimates the samples by that factor.

`power` is printed with two decimals, so `9` becomes `-p 7.00`.

`inputFile: "-"` passes `-i -`, which makes `sendiq` read its own stdin. This module gives the process no stdin data, so in practice `-` reads nothing. Write the samples to a file or a named pipe and pass that path instead.

## Example

```json
{
  "inputFile": "/home/pi/signal.iq",
  "freq": 434000000,
  "sampleRate": 96000,
  "power": 2.5,
  "sharedMemToken": 12345,
  "loopMode": true
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/sendiq" -i /home/pi/signal.iq -f 434000000 \
  -s 96000 -p 2.50 -m 12345 -l
```

With `loopMode`, the program never exits by itself. Give `Exec` a timeout or call [`Stop`](running-modules.md#stopping-a-transmission).

## Run-time control over shared memory

With `sharedMemToken` set, `sendiq` creates a System V shared memory segment with that key and checks it while it transmits. It also forces the I/Q type to `float`, whatever `iqType` says.

The segment holds this C struct from rpitx's `sendiq.cpp`:

```c
struct shared_memory_struct {
    bool  updated;
    int   command;
    float data;
    char  common_data[1024];
};
```

Write a command and set `updated` to true. The commands:

| Command | Effect | `data` |
| --- | --- | --- |
| `1111` | Switch to I/Q mode | unused |
| `2222` | Switch to frequency and amplitude mode | unused |
| `3333` | Set the drive level | level, 0 to 7 |
| `4444` | Change the frequency | frequency in Hz |

From Python with the `sysv_ipc` package, for example:

```python
import struct

import sysv_ipc

shm = sysv_ipc.SharedMemory(12345)

# updated=True, command=4444, data=144.5 MHz
shm.write(struct.pack("?if", True, 4444, 144500000.0))
```

`?if` uses native alignment, which matches the C struct on the Pi. A 32-bit float holds about seven significant digits, so a frequency sent this way lands on the nearest value a float can hold, within a few tens of Hz at 144 MHz.
