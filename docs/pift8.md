---
title: "pift8"
description: "The gorpitx pift8 module: transmit an FT8 message in a 15-second slot, with frequency offset, slot choice and repeat mode."
nav_order: 7
parent: "Modules"
---

# pift8: FT8

Sends an FT8 message, the weak-signal digital mode with 15-second transmit slots. The module name is `pift8`, the binary name, while the Go names are `FT8` and `ModuleNameFT8`. Code: [`ft8.go`](https://github.com/psyb0t/gorpitx/blob/main/ft8.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `frequency` | `Frequency` | `float64` | yes | Dial frequency in Hz, 5000 to 1500000000 |
| `message` | `Message` | `string` | yes | The FT8 message, not empty or only spaces |
| `ppm` | `PPM` | `*float64` | no | Clock correction in parts per million, any value; without it the program uses NTP |
| `offset` | `Offset` | `*float64` | no | Audio offset in Hz, 0 to 2500; the program defaults to 1240 |
| `slot` | `Slot` | `*int` | no | `0` first 15 s slot, `1` second slot, `2` every slot |
| `repeat` | `Repeat` | `*bool` | no | `true` adds `-r` and keeps sending every 15 s |

The tone goes out at `frequency + offset`. gorpitx does not check the message format; `pift8` encodes it, so a message FT8 cannot encode fails in the program, not in validation.

FT8 slots are tied to the clock, so the Pi's time has to be right. Keep NTP running.

## Example

```json
{
  "frequency": 14074000,
  "message": "CQ N0CALL FN31",
  "offset": 1500,
  "slot": 1,
  "repeat": true
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pift8" -f 14074000 -m "CQ N0CALL FN31" -o 1500 -s 1 -r
```

With `repeat`, the program never exits by itself. Give `Exec` a timeout or call [`Stop`](running-modules.md#stopping-a-transmission).

## Message shapes

A standard exchange goes like this, one message per slot:

| Step | Message |
| --- | --- |
| Call CQ | `CQ N0CALL FN31` |
| Answer with your grid | `N0CALL N1CALL EM69` |
| Signal report | `N1CALL N0CALL -15` |
| Report back, with R | `N0CALL N1CALL R-08` |
| Confirm | `N1CALL N0CALL RR73` |
| Sign off | `N0CALL N1CALL 73` |
