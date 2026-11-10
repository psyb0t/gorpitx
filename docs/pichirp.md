---
title: "pichirp"
description: "The gorpitx pichirp module: sweep a carrier across a bandwidth around a center frequency over a set number of seconds."
nav_order: 3
parent: "Modules"
---

# pichirp: a carrier sweep

Sweeps a carrier across a band around a center frequency, over a given time. Code: [`pichirp.go`](https://github.com/psyb0t/gorpitx/blob/main/pichirp.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Center frequency in Hz, 5000 to 1500000000 |
| `bandwidth` | `Bandwidth` | `float64` | yes | Sweep width in Hz, greater than 0 |
| `time` | `Time` | `float64` | yes | Sweep duration in seconds, greater than 0 |

## Example

```json
{
  "frequency": 434000000,
  "bandwidth": 100000,
  "time": 5
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pichirp" 434000000 100000 5
```

The arguments are positional: frequency, bandwidth, time. Frequency and bandwidth are rounded to whole Hz; time keeps its decimals.
