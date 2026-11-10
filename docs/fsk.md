---
title: "fsk"
description: "The gorpitx fsk module: encode text or a file as audio FSK with minimodem and transmit it with sendiq, at a default 50 baud."
nav_order: 10
parent: "Modules"
---

# fsk: FSK text with minimodem

Encodes text, or the contents of a file, as audio FSK with [minimodem](http://www.whence.com/minimodem/), and transmits that audio with rpitx's `sendiq`. On the receiving side, `minimodem --rx <baud>` on the receiver's audio output decodes it. Code: [`fsk.go`](https://github.com/psyb0t/gorpitx/blob/main/fsk.go) and [`scripts/fsk.sh`](https://github.com/psyb0t/gorpitx/blob/main/scripts/fsk.sh).

This is a script module. It needs `minimodem` and `sox` on the Pi, and `sendiq` in [`GORPITX_PATH`](configuration.md#gorpitx_path).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `inputType` | `InputType` | `string` | yes | `"file"` or `"text"` (`gorpitx.InputTypeFile`, `gorpitx.InputTypeText`) |
| `text` | `Text` | `string` | when `inputType` is `text` | Not empty or only spaces |
| `file` | `File` | `string` | when `inputType` is `file` | Path to an existing file |
| `baudRate` | `BaudRate` | `*int` | no | Greater than 0; gorpitx sends 50 when it is not set |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |

`inputType` decides which of `text` and `file` is used. If you send both, the other one is ignored, not rejected. The module appends a newline to the input before sending it.

50 baud is the default because it decoded cleanest in testing. Higher rates such as 110 or 300 work too and need a stronger signal.

## Example

```json
{
  "inputType": "text",
  "text": "HELLO DE N0CALL",
  "frequency": 144390000
}
```

Runs:

```bash
stdbuf -oL /tmp/fsk.sh 50 144390000
```

with `HELLO DE N0CALL` and a newline on stdin, and `RPITX_PATH` set to `GORPITX_PATH`.

## What the script does

1. Reads all of stdin and encodes it with `minimodem --tx <baud>` into `/tmp/fsk_<pid>.wav`.
2. Converts the WAV with `sox` to raw 16-bit signed samples, 48 kHz, two channels.
3. Pipes that into `$RPITX_PATH/sendiq -i /dev/stdin -s 48000 -f <frequency> -t i16`.
4. Deletes the WAV file when it exits.

The whole input is encoded before anything goes on the air, so a long file takes a moment before the transmission starts. If `minimodem` or the `sox` and `sendiq` pipeline fails, the script prints an error to stderr and exits with status 1.
