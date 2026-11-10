---
title: "Home"
description: "gorpitx is a Go library that runs rpitx transmitter programs on a Raspberry Pi from validated JSON arguments, one transmission at a time."
nav_order: 1
permalink: /
seo:
  type: WebSite
---

# gorpitx

gorpitx runs [rpitx](https://github.com/F5OEO/rpitx) from Go. You hand it a module name and a JSON blob, it checks every field, builds the command line, and runs the rpitx program as a child process with a timeout, a graceful stop and live output. Each rpitx binary has its own argument style, most of them print a usage line and exit when you get one wrong, and all of them leave process handling to you. This library does that part.

It wraps 12 rpitx programs: FM with RDS, a plain carrier, a chirp sweep, Morse, POCSAG pagers, spectrum painting, FT8, SSTV, RTTY, FSK text, audio from a unix socket, and raw I/Q samples. Only one runs at a time, on purpose: the Pi has one transmitter.

## Install

```bash
go get github.com/psyb0t/gorpitx
```

The module declares `go 1.26` in its [`go.mod`](https://github.com/psyb0t/gorpitx/blob/main/go.mod). Transmitting needs a Raspberry Pi with rpitx installed and root. Development and tests do not: set `ENV=dev` and every module runs a harmless shell loop instead of a transmitter.

## Quick look

```go
rpitx := gorpitx.GetInstance()

args, _ := json.Marshal(gorpitx.TUNE{
    Frequency: 434000000, // Hz
})

err := rpitx.Exec(
    context.Background(),
    gorpitx.ModuleNameTUNE,
    args,
    10*time.Second,
)
```

That runs `stdbuf -oL $GORPITX_PATH/tune -f 434000000` for ten seconds, then stops it and returns `commonerrors.ErrTimeout`. The [getting started](getting-started.md) page has the full program.

## Where to go next

| You want to | Read |
| --- | --- |
| Install rpitx and run a first transmission | [Getting started](getting-started.md) |
| Set the rpitx path, dev mode and root | [Configuration](configuration.md) |
| Run, time out, stop and stream a module | [Running modules](running-modules.md) |
| Look up every field of a module | [Modules](modules.md) |
| Tell errors apart with `errors.Is` | [Errors](errors.md) |
| Fix a panic, a refused frequency or a silent script | [Troubleshooting](troubleshooting.md) |
| See how a call becomes a process | [Architecture](architecture.md) |
| Build, lint and test the library | [Development](development.md) |

RF transmission is regulated. Most of these modules put a real signal on the air, and an unfiltered Pi spits harmonics well outside the frequency you asked for. Transmit only where your license allows it.

The Go package docs are on [pkg.go.dev](https://pkg.go.dev/github.com/psyb0t/gorpitx), and the source is on [GitHub](https://github.com/psyb0t/gorpitx).
