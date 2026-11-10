---
title: "Modules"
description: "The 12 rpitx programs gorpitx wraps, their module names, Go types and frequency units, and the validation rules every module shares."
nav_order: 5
has_children: true
---

# Modules

A module is one rpitx program behind the `Module` interface. Its name is the rpitx binary name, its Go struct defines the JSON fields, and its `ParseArgs` method validates them and builds the command line. Pick a module from the table, read its page for the fields, and pass the name to [`Exec`](running-modules.md#exec).

| Module name | Constant | Go type | What it sends | Program run |
| --- | --- | --- | --- | --- |
| [`pifmrds`](pifmrds.md) | `ModuleNamePIFMRDS` | `PIFMRDS` | FM broadcast with RDS | `pifmrds` |
| [`tune`](tune.md) | `ModuleNameTUNE` | `TUNE` | A plain carrier | `tune` |
| [`pichirp`](pichirp.md) | `ModuleNamePICHIRP` | `PICHIRP` | A carrier sweep | `pichirp` |
| [`morse`](morse.md) | `ModuleNameMORSE` | `MORSE` | Morse code | `morse` |
| [`pocsag`](pocsag.md) | `ModuleNamePOCSAG` | `POCSAG` | POCSAG pager messages | `pocsag` |
| [`spectrumpaint`](spectrumpaint.md) | `ModuleNameSPECTRUMPAINT` | `SPECTRUMPAINT` | A picture on a waterfall display | `spectrumpaint` |
| [`pift8`](pift8.md) | `ModuleNameFT8` | `FT8` | FT8 | `pift8` |
| [`pisstv`](pisstv.md) | `ModuleNamePISSSTV` | `PISSTV` | SSTV pictures, Martin 1 | `pisstv` |
| [`pirtty`](pirtty.md) | `ModuleNamePIRTTY` | `PIRTTY` | RTTY | `pirtty` |
| [`fsk`](fsk.md) | `ModuleNameFSK` | `FSK` | FSK text via minimodem | `/tmp/fsk.sh` |
| [`audiosock-broadcast`](audiosock-broadcast.md) | `ModuleNameAudioSockBroadcast` | `AudioSockBroadcast` | Audio from a unix socket, AM, DSB, SSB, FM or raw | `/tmp/audiosock_broadcast.sh` |
| [`sendiq`](sendiq.md) | `ModuleNameSENDIQ` | `SENDIQ` | Raw I/Q samples | `sendiq` |

The SSTV constant is spelled `ModuleNamePISSSTV`, with three S. It is exported under that name, so it stays.

## Rules every module shares

### Frequency

Every module checks its carrier frequency against one range, 5 kHz to 1500 MHz. (The rpitx usage texts say 50 kHz; gorpitx allows down to 5 kHz.) The limits are `minFreqKHz` and `maxFreqKHz` in [`gorpitx.go`](https://github.com/psyb0t/gorpitx/blob/main/gorpitx.go).

| Module | Field | Unit | Accepted |
| --- | --- | --- | --- |
| `pifmrds` | `freq` | MHz | 0.1 to 1500, in steps of 0.1 |
| `sendiq` | `freq` | Hz | 5000 to 1500000000 |
| all others | `frequency` | Hz | 5000 to 1500000000 |

Hz values are passed to the program rounded to a whole number. A missing or zero frequency fails with `frequency must be positive` (`commonerrors.ErrInvalidValue`), except in `pifmrds`, where it fails as a missing required field. A value outside the range fails with `ErrFreqOutOfRange`.

The range check is the hardware limit, not a license. Most of that range is not yours to transmit on.

### Required and optional fields

- Required fields are plain values. Optional fields are pointers in the Go structs (`*int`, `*float64`, `*bool`, `*string`) with `omitempty`, so leaving one out differs from sending zero.
- An optional flag goes on the command line only when you set it. Where the program takes positional arguments, gorpitx fills in a default instead; each page says which.
- Text fields that must not be empty also reject strings of only spaces.
- File fields are checked with `os.Stat` before anything starts, and a missing file fails with `commonerrors.ErrFileNotFound`.
- Validation stops at the first bad field. The [Errors](errors.md) page lists what comes back.

Remember that values you leave out keep what the previous call set. See [Arguments persist between calls](running-modules.md#arguments-persist-between-calls).

## rpitx programs not wrapped

rpitx ships more programs than these 12. The ones that would fit the same pattern, `freedv` (FreeDV digital voice, `freedv vco.rf frequency [samplerate]`) and `piopera` (OPERA, `piopera CALLSIGN mode frequency`), have no module yet. Run them yourself, or add a module as described in [Development](development.md#adding-a-module).
