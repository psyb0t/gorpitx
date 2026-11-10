---
title: "tune"
description: "The gorpitx tune module: transmit a plain unmodulated carrier on one frequency in Hz, with optional PPM clock correction."
nav_order: 2
parent: "Modules"
---

# tune: a plain carrier

Puts an unmodulated carrier on one frequency. Useful for finding your signal on a receiver, checking a filter, or measuring the Pi's frequency error. Code: [`tune.go`](https://github.com/psyb0t/gorpitx/blob/main/tune.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `exitImmediate` | `ExitImmediate` | `*bool` | no | `true` adds `-e`: the program exits at once and leaves the carrier running |
| `ppm` | `PPM` | `*float64` | no | Clock correction in parts per million, must be greater than 0 |

`tune` keeps transmitting until it is stopped, so give `Exec` a timeout or call [`Stop`](running-modules.md#stopping-a-transmission). With `exitImmediate`, the process exits right away and the carrier stays on after `Exec` returns. Stopping that carrier is up to you.

`ppm` must be positive here, while `pifmrds` and `pift8` accept any value. A negative correction is refused.

## Example

```json
{
  "frequency": 434000000,
  "ppm": 2.5
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/tune" -f 434000000 -p 2.5
```
