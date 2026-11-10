---
title: "Troubleshooting"
description: "Fixes for real gorpitx failures: the root panic, exit status 127 from stdbuf, ErrExecuting, refused frequencies, silent scripts and stale arguments."
nav_order: 7
---

# Troubleshooting

Each entry starts with what you see, then the cause and the fix.

## GetInstance panics about root

```text
panic: PIrateRF must be run as root in production mode
```

The program is not running as root and `ENV` is not `dev`. The message names PIrateRF, a project built on gorpitx, but it comes from gorpitx's own check in `newRPITX`. Run the program with `sudo`, or set `ENV=dev` to develop without transmitting. See [Configuration](configuration.md#root).

## Exec fails with exit status 127

`Exec` returns an error like this (on one line, call sites shortened):

```text
failed to wait for process: (exit 127): stdbuf: failed to run command '$HOME/rpitx/tune': No such file or directory: failed [...]
```

`GORPITX_PATH` is not set, so the library used its literal default `$HOME/rpitx`, which nothing expands. The same status with a real path means the binary is not in that directory. Set `GORPITX_PATH` to the absolute path of the rpitx checkout, and pass it through `sudo` explicitly:

```bash
sudo GORPITX_PATH=/home/pi/rpitx ./your-program
```

## Exec returns ErrExecuting

```text
RPITX is busy executing another command
```

Another `Exec` is still running. Only one module runs at a time, and the second call does not wait. Programs like `tune`, `sendiq` with `loopMode`, `pift8` with `repeat`, and everything in dev mode never exit by themselves, so the first call holds the instance until its timeout, its context or [`Stop`](running-modules.md#stopping-a-transmission) ends it.

## A frequency is refused

`frequency out of RPiTX range` means the value is outside 5 kHz to 1500 MHz in that module's unit. The usual cause is units: `pifmrds` takes MHz (`107.9`), every other module takes Hz (`107900000`). A `pifmrds` frequency like `107.95` fails with `frequency precision too high`, because it takes one decimal place at most. See [Frequency](modules.md#frequency).

## Stop returns an error after stopping the process

```text
failed to stop process: terminated
```

That is a successful stop. The process exited on SIGTERM, and `Stop` reports it as an error wrapping `commonerrors.ErrTerminated`. Check with `errors.Is(err, commonerrors.ErrTerminated)` and carry on.

## A field from the previous call is still applied

You sent `ppm` once, left it out of the next call, and the command line still has `-p`. Each module keeps the values from the previous call, because the same struct is reused and not reset. Send every field you care about on each call, or send `null` to clear an optional one. See [Arguments persist between calls](running-modules.md#arguments-persist-between-calls).

## audiosock-broadcast exits at once

The output shows:

```text
ERROR: Unix socket /tmp/audio.sock does not exist
```

The path is missing or is not a socket. A FIFO made with `mkfifo` fails this check. The script connects to the socket as a client, so start a listener that writes audio first, for example `socat -u - UNIX-LISTEN:/tmp/audio.sock` fed by ffmpeg. See [audiosock-broadcast](audiosock-broadcast.md#the-socket-and-the-audio-format).

`ERROR: sendiq not found at <path>` means `GORPITX_PATH` does not point at the rpitx checkout.

## fsk fails before transmitting

`Failed to encode input to FSK audio` means `minimodem` is missing or failed. `Failed to convert and transmit FSK data` means `sox` is missing or `$RPITX_PATH/sendiq` could not run. Install the tools from [Getting started](getting-started.md#4-install-rpitx-on-the-pi) and check `GORPITX_PATH`.

## The program exits before main

```text
ERROR failed to write FSK script path=/tmp/fsk.sh err="open /tmp/fsk.sh: permission denied"
```

The package `init` could not write its scripts to `/tmp` and called `os.Exit(1)`. The usual cause is that the scripts were written by a different user. Run the program or its tests once as your normal user and `/tmp/fsk.sh` belongs to you; run it as root afterwards and the write fails, even for root, on systems with `fs.protected_regular` enabled (the default on current Debian and Raspberry Pi OS). The same happens the other way round after a root run.

Delete the three files and run again as the user you want:

```bash
sudo rm -f /tmp/fsk.sh /tmp/audiosock_broadcast.sh /tmp/modulation.sh
```

## Stream channels stop receiving lines

A stream channel that blocks for more than 100 ms is dropped and is never closed, so a `range` over it hangs. Use buffered channels and read them in their own goroutine. See [Streaming output](running-modules.md#streaming-output).
