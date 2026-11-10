---
title: "morse"
description: "The gorpitx morse module: send a text message as Morse code on a frequency in Hz at a rate given in dits per minute."
nav_order: 4
parent: "Modules"
---

# morse: Morse code

Keys a text message as Morse code (CW). Code: [`morse.go`](https://github.com/psyb0t/gorpitx/blob/main/morse.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `rate` | `Rate` | `int` | yes | Dits per minute, greater than 0 |
| `message` | `Message` | `string` | yes | Text to send, not empty or only spaces |

## Example

```json
{
  "frequency": 14070000,
  "rate": 20,
  "message": "CQ CQ DE N0CALL"
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/morse" 14070000 20 "CQ CQ DE N0CALL"
```

The message travels as one argument, so spaces in it are safe. The program exits when the message is done.
