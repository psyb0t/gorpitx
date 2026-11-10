---
title: "pocsag"
description: "The gorpitx pocsag module: send POCSAG pager messages at 512, 1200 or 2400 baud, with messages fed to the rpitx program on stdin."
nav_order: 5
parent: "Modules"
---

# pocsag: pager messages

Sends one or more POCSAG pager messages. The rpitx `pocsag` program reads its messages from stdin as `address:message` lines and takes only flags on the command line, and gorpitx does the same split for you. Code: [`pocsag.go`](https://github.com/psyb0t/gorpitx/blob/main/pocsag.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `baudRate` | `BaudRate` | `*int` | no | `512`, `1200` or `2400`; the program defaults to 1200 |
| `functionBits` | `FunctionBits` | `*int` | no | `0` to `3`; the program defaults to 3 |
| `numericMode` | `NumericMode` | `*bool` | no | `true` adds `-n`, numeric messages |
| `repeatCount` | `RepeatCount` | `*int` | no | Greater than 0; the program defaults to 4 |
| `invertPolarity` | `InvertPolarity` | `*bool` | no | `true` adds `-i` |
| `debug` | `Debug` | `*bool` | no | `true` adds `-d` |
| `messages` | `Messages` | `[]POCSAGMessage` | yes | At least one |

Each entry in `messages`:

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `address` | `Address` | `int` | yes | Pager address, 0 or more |
| `message` | `Message` | `string` | yes | Not empty or only spaces |
| `functionBits` | `FunctionBits` | `*int` | no | `0` to `3`, validated but not sent (see below) |

The defaults in the table are the `pocsag` program's own. gorpitx leaves a flag out when you leave the field out.

## Example

```json
{
  "frequency": 466230000,
  "baudRate": 1200,
  "functionBits": 3,
  "repeatCount": 4,
  "messages": [
    {
      "address": 123456,
      "message": "Test page"
    },
    {
      "address": 789012,
      "message": "Second page"
    }
  ]
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pocsag" -f 466230000 -r 1200 -b 3 -t 4
```

with this on stdin, lines joined by newlines and no newline after the last one:

```text
123456:Test page
789012:Second page
```

## Per-message function bits are ignored

`POCSAGMessage.FunctionBits` is checked for the 0 to 3 range and then dropped: the stdin line is always `address:message`. The `pocsag` program does support a per-message override, as a letter `a` to `d` right after the address (`123456b:text` means function bits 1), but gorpitx does not write it. Use the top-level `functionBits` for now; it applies to every message.
