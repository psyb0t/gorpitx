---
title: "pirtty"
description: "The gorpitx pirtty module: send a text message as RTTY at 45.45 baud in Baudot code, with a configurable space tone."
nav_order: 9
parent: "Modules"
---

# pirtty: RTTY

Sends a text message as RTTY: Baudot code at 45.45 baud, with two tones 170 Hz apart. Receive it in USB mode with any RTTY decoder. Code: [`pirtty.go`](https://github.com/psyb0t/gorpitx/blob/main/pirtty.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `spaceFrequency` | `SpaceFrequency` | `*int` | no | Space tone in Hz, greater than 0; gorpitx sends 170 when it is not set |
| `message` | `Message` | `string` | yes | Not empty or only spaces |

The program puts the mark tone at `spaceFrequency + 170`, so the shift is always the standard 170 Hz. The tones sit above the carrier frequency, like audio fed to an SSB transmitter. With the default space tone of 170 Hz the mark is at 340 Hz; decoders usually expect something like 2125/2295 Hz or 1275/1445 Hz, so set `spaceFrequency` to match your decoder.

Baudot has letters and figures shifts; the program inserts them as the text switches between letters and digits or punctuation. Lowercase letters go out as uppercase.

## Example

```json
{
  "frequency": 14080000,
  "spaceFrequency": 2125,
  "message": "RYRYRY DE N0CALL"
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pirtty" 14080000 2125 "RYRYRY DE N0CALL"
```

The arguments are positional: frequency, space tone, message. Without `spaceFrequency` the second argument is `170`.
