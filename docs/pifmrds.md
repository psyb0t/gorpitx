---
title: "pifmrds"
description: "The gorpitx pifmrds module: FM broadcast of an audio file with RDS station name, radiotext and PI code, with frequency in MHz."
nav_order: 1
parent: "Modules"
---

# pifmrds: FM with RDS

Broadcasts an audio file as FM, with RDS data a car radio can display: a station name, a scrolling radiotext and a PI code. It runs rpitx's port of [Pi-FM-RDS](https://github.com/ChristopheJacquet/PiFmRds). This is the one module that takes its frequency in MHz. Code: [`pifmrds.go`](https://github.com/psyb0t/gorpitx/blob/main/pifmrds.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `freq` | `Freq` | `float64` | yes | MHz, 0.1 to 1500, at most one decimal place (`107.9` passes, `107.95` fails with `ErrFreqPrecision`) |
| `audio` | `Audio` | `string` | yes | Path to an existing audio file |
| `pi` | `PI` | `string` | no | Exactly 4 hex digits, such as `1234` or `FFFF` |
| `ps` | `PS` | `string` | no | Station name, at most 8 bytes, not only spaces |
| `rt` | `RT` | `string` | no | Radiotext, at most 64 bytes |
| `ppm` | `PPM` | `*float64` | no | Clock correction in parts per million, any value |
| `controlPipe` | `ControlPipe` | `*string` | no | Path to an existing named pipe for run-time changes |

The audio file goes through libsndfile, so WAV and Ogg/Vorbis work and MP3 does not. A stereo file gives an FM stereo signal. Pi-FM-RDS can read audio from stdin with `-audio -`, but gorpitx checks that the path exists as a file, so stdin input is not available through this module.

Without `ps`, Pi-FM-RDS alternates the station name between `Pi-FmRds` and a counter.

## Example

```json
{
  "freq": 107.9,
  "audio": "/home/pi/sound.wav",
  "pi": "1234",
  "ps": "GORPITX",
  "rt": "Hello from Go"
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pifmrds" -freq 107.9 -audio /home/pi/sound.wav \
  -pi 1234 -ps GORPITX -rt "Hello from Go"
```

The frequency is always printed with one decimal place.

## Changing PS and RT while it runs

Create a named pipe before the call and pass its path as `controlPipe`:

```bash
mkfifo /tmp/rds_ctl
```

While the module runs, write one command per line into the pipe:

```bash
echo "PS NEWNAME" > /tmp/rds_ctl
echo "RT Now playing something else" > /tmp/rds_ctl
echo "TA ON" > /tmp/rds_ctl
```

`PS` and `RT` change the station name and radiotext, and `TA` sets the traffic announcement flag. These commands belong to Pi-FM-RDS; gorpitx only passes the pipe path with `-ctl`.
