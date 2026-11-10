# gorpitx

[![CI](https://github.com/psyb0t/gorpitx/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gorpitx/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gorpitx/badges/coverage.svg)](https://github.com/psyb0t/gorpitx/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gorpitx/badges/version.svg)](https://github.com/psyb0t/gorpitx/releases)
[![license](https://raw.githubusercontent.com/psyb0t/gorpitx/badges/license.svg)](https://github.com/psyb0t/gorpitx/blob/main/LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/gorpitx/badges/importers.svg)](https://github.com/psyb0t/gorpitx/blob/badges/importers.md)

**Go wrapper that runs rpitx modules without the hassle.**

Tired of wrestling with raw C binaries that print a usage line and die when you get one argument wrong? gorpitx wraps [rpitx](https://github.com/F5OEO/rpitx) so you hand it a module name and a JSON blob. It checks every field, builds the command line, and runs the transmitter as a child process with a timeout, a graceful stop and live output. One transmission at a time, because the Pi has one transmitter.

## Table of Contents

- [What it does](#what-it-does)
- [Quick start](#quick-start)
- [Requirements](#requirements)
- [Documentation](#documentation)
- [Development](#development)
- [Legal notice](#legal-notice)
- [Sponsor](#sponsor)
- [License](#license)

## What it does

Runs 12 rpitx programs through one API, with dev mode (a harmless shell loop instead of the transmitter, for tests and laptops) and production mode (actual RF, as root).

| Module | What it sends | Frequency unit |
| --- | --- | --- |
| [`pifmrds`](https://psyb0t.github.io/gorpitx/pifmrds/) | FM broadcast with RDS station name and radiotext | MHz |
| [`tune`](https://psyb0t.github.io/gorpitx/tune/) | A plain carrier | Hz |
| [`pichirp`](https://psyb0t.github.io/gorpitx/pichirp/) | A carrier sweep | Hz |
| [`morse`](https://psyb0t.github.io/gorpitx/morse/) | Morse code | Hz |
| [`pocsag`](https://psyb0t.github.io/gorpitx/pocsag/) | Pager messages | Hz |
| [`spectrumpaint`](https://psyb0t.github.io/gorpitx/spectrumpaint/) | A picture on a waterfall display | Hz |
| [`pift8`](https://psyb0t.github.io/gorpitx/pift8/) | FT8 | Hz |
| [`pisstv`](https://psyb0t.github.io/gorpitx/pisstv/) | SSTV pictures, Martin 1 | Hz |
| [`pirtty`](https://psyb0t.github.io/gorpitx/pirtty/) | RTTY | Hz |
| [`fsk`](https://psyb0t.github.io/gorpitx/fsk/) | FSK text through minimodem and sox | Hz |
| [`audiosock-broadcast`](https://psyb0t.github.io/gorpitx/audiosock-broadcast/) | Audio from a unix socket as FM, AM, DSB, USB, LSB or raw | Hz |
| [`sendiq`](https://psyb0t.github.io/gorpitx/sendiq/) | Raw I/Q samples, with run-time control over shared memory | Hz |

## Quick start

```bash
go get github.com/psyb0t/gorpitx
```

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/psyb0t/gorpitx"
)

func main() {
	rpitx := gorpitx.GetInstance()

	args, err := json.Marshal(gorpitx.MORSE{
		Frequency: 14070000, // Hz
		Rate:      20,       // dits per minute
		Message:   "CQ CQ DE N0CALL",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = rpitx.Exec(
		context.Background(),
		gorpitx.ModuleNameMORSE,
		args,
		30*time.Second,
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

Try it anywhere with `ENV=dev go run .`: the module still validates its arguments, then a shell loop prints the command it would have run until the timeout hits and `Exec` returns `commonerrors.ErrTimeout`. On the Pi, run it as root with `GORPITX_PATH` pointing at your rpitx checkout.

## Requirements

- A Raspberry Pi that rpitx supports, with rpitx installed.
- Root in production mode. `GetInstance` panics otherwise.
- `GORPITX_PATH` set to the directory with the rpitx binaries, as an absolute path. The default, `$HOME/rpitx`, is used literally and does not work.
- `minimodem` and `sox` for `fsk`, `socat` and `csdr` for `audiosock-broadcast`.

The [getting started guide](https://psyb0t.github.io/gorpitx/getting-started/) walks through all of it.

## Documentation

The full documentation lives at **https://psyb0t.github.io/gorpitx/**.

| Section | What's in it |
| --- | --- |
| [Getting started](https://psyb0t.github.io/gorpitx/getting-started/) | rpitx install, extra tools, dev mode, first real transmission |
| [Configuration](https://psyb0t.github.io/gorpitx/configuration/) | `GORPITX_PATH`, `ENV`, the root check, files written to `/tmp` |
| [Running modules](https://psyb0t.github.io/gorpitx/running-modules/) | `Exec`, timeouts, `Stop`, output streaming, concurrency |
| [Modules](https://psyb0t.github.io/gorpitx/modules/) | Every field, rule and example for all 12 modules |
| [Errors](https://psyb0t.github.io/gorpitx/errors/) | The sentinels to match with `errors.Is` |
| [Troubleshooting](https://psyb0t.github.io/gorpitx/troubleshooting/) | Real failures with the exact message and the fix |
| [Architecture](https://psyb0t.github.io/gorpitx/architecture/) | How a call becomes a process |
| [Development](https://psyb0t.github.io/gorpitx/development/) | Make targets, tests, CI, adding a module |

## Development

```bash
make lint            # go fix -diff, golangci-lint, shellcheck
make test-coverage   # tests with the 90% coverage gate
make docs            # build the docs site into .testing/docs-site
```

No test needs a Pi or root. See [Development](https://psyb0t.github.io/gorpitx/development/) for the rest, and [CHANGELOG.md](https://github.com/psyb0t/gorpitx/blob/main/CHANGELOG.md) for what changed in each release.

## Legal notice

**RF transmission is regulated.** Get the license before you transmit. This software is for licensed amateur radio operators, low-power experiments in bands you may use, and education. An unfiltered Pi also radiates harmonics far from the frequency you picked, so put a filter on it.

Absolutely NOT for broadcasting without authorization. The fines are severe.

## Sponsor

<a href="https://www.buymeacoffee.com/psyb0t"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me a Coffee" height="60" width="217"></a>

## License

MIT License, see [LICENSE](https://github.com/psyb0t/gorpitx/blob/main/LICENSE). Use responsibly.
