---
title: "audiosock-broadcast"
description: "The gorpitx audiosock-broadcast module: read raw audio from a unix socket and transmit it as FM, AM, DSB, USB, LSB or raw through csdr and sendiq."
nav_order: 11
parent: "Modules"
---

# audiosock-broadcast: audio from a unix socket

Reads raw audio from a unix socket and puts it on the air, modulated through [csdr](https://github.com/F5OEO/csdr) and sent with rpitx's `sendiq`. Point any audio source at the socket: a file decoded by ffmpeg, a microphone, a program that generates tones. Code: [`audiosock_broadcast.go`](https://github.com/psyb0t/gorpitx/blob/main/audiosock_broadcast.go), [`scripts/audiosock_broadcast.sh`](https://github.com/psyb0t/gorpitx/blob/main/scripts/audiosock_broadcast.sh) and [`scripts/modulation.sh`](https://github.com/psyb0t/gorpitx/blob/main/scripts/modulation.sh).

This is a script module. It needs `socat` and `csdr` on the Pi, and `sendiq` in [`GORPITX_PATH`](configuration.md#gorpitx_path).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `socketPath` | `SocketPath` | `string` | yes | Path of the unix socket to read from |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `sampleRate` | `SampleRate` | `*int` | no | Audio sample rate in Hz, greater than 0; gorpitx sends 48000 when it is not set |
| `modulation` | `Modulation` | `*string` | no | `AM`, `DSB`, `USB`, `LSB`, `FM` or `RAW` (the `gorpitx.Modulation*` constants); gorpitx sends `FM` when it is not set |
| `gain` | `Gain` | `*float64` | no | Multiplier applied to the audio, 0 or more; gorpitx sends 1 when it is not set |

gorpitx does not check that the socket exists. The script does, when it starts, and exits with status 1 and `ERROR: Unix socket <path> does not exist` if it does not.

## The socket and the audio format

The script connects to the socket as a client (`socat UNIX-CONNECT:<socketPath> STDOUT`), so some other process has to be listening on it and writing audio. A named pipe made with `mkfifo` does not work: the script checks for a socket and refuses anything else.

The audio must be raw PCM with no header: signed 16-bit little-endian, mono, at `sampleRate`.

Start the source first, then call `Exec`. Stream a file with ffmpeg and socat:

```bash
ffmpeg -i music.mp3 -f s16le -ar 48000 -ac 1 - \
  | socat -u - UNIX-LISTEN:/tmp/audio.sock
```

Or a microphone with ALSA:

```bash
arecord -f S16_LE -r 48000 -c 1 -t raw \
  | socat -u - UNIX-LISTEN:/tmp/audio.sock
```

The transmission ends when the source closes the connection, or when `Exec` times out or is stopped.

## Modulations

`modulation.sh` turns the audio into I/Q samples for `sendiq` with one csdr chain per mode:

| Mode | csdr chain after `convert_s16_f` and `gain_ff` |
| --- | --- |
| `FM` | `fmmod_fc` |
| `AM` | `dsb_fc`, `add_dcoffset_cc`, `agc_ff` |
| `DSB` | `dsb_fc`, `agc_ff` |
| `USB` | `dsb_fc`, `bandpass_fir_fft_cc 0.002 0.06 0.01`, `agc_ff` |
| `LSB` | `dsb_fc`, `bandpass_fir_fft_cc -0.06 -0.002 0.01`, `agc_ff` |
| `RAW` | nothing more: conversion and gain only, no modulation, no AGC |

USB and LSB are slow on a Pi Zero. The FFT bandpass filter adds latency, artifacts and dropouts there. Use DSB instead: it sends both sidebands, so a receiver in USB or LSB mode hears it.

## Example

```json
{
  "socketPath": "/tmp/audio.sock",
  "frequency": 14200000,
  "modulation": "DSB",
  "gain": 2
}
```

Runs:

```bash
stdbuf -oL /tmp/audiosock_broadcast.sh 14200000 /tmp/audio.sock 48000 DSB 2
```

with `RPITX_PATH` set to `GORPITX_PATH`. The script then runs:

```bash
socat UNIX-CONNECT:/tmp/audio.sock STDOUT \
  | /tmp/modulation.sh DSB 2 \
  | "$RPITX_PATH/sendiq" -i /dev/stdin -s 48000 -f 14200000 -t float
```

## Side effects

- The script appends every run to `/tmp/audiosock_broadcast.log` as well as printing it.
- On SIGINT or SIGTERM it runs `pkill -f sendiq` and `pkill -f csdr`. That kills every process on the Pi whose command line contains those words, not only its own.
