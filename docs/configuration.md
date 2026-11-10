---
title: "Configuration"
description: "The two environment variables gorpitx reads, GORPITX_PATH and ENV, the root check, and the files the package writes under /tmp."
nav_order: 3
---

# Configuration

gorpitx reads two environment variables. There is no config file and no options struct. `GORPITX_PATH` is read once, when [`GetInstance`](running-modules.md#getting-the-instance) first runs, so changing it later does nothing until the program restarts. `ENV` is read by that first call for the root check, and again by every `Exec` to decide between the real program and the dev mode loop.

| Variable | Default | Read by | Meaning |
| --- | --- | --- | --- |
| `GORPITX_PATH` | `$HOME/rpitx` (literal, see below) | [`config.go`](https://github.com/psyb0t/gorpitx/blob/main/config.go) | Directory that holds the rpitx binaries |
| `ENV` | unset, which means production | [goenv](https://github.com/psyb0t/goenv) | `dev` turns on dev mode; any other value is production |

## GORPITX_PATH

gorpitx runs `stdbuf -oL <GORPITX_PATH>/<module name> <args>`. The module name is the rpitx binary name (`pifmrds`, `tune`, `pift8` and so on, listed on the [Modules](modules.md) page). The two script modules, `fsk` and `audiosock-broadcast`, get the same directory in the `RPITX_PATH` environment variable and call `$RPITX_PATH/sendiq`.

Set it to an absolute path:

```bash
export GORPITX_PATH=/home/pi/rpitx
```

The default value is the string `$HOME/rpitx`. Neither the config parser nor the process launcher expands `$HOME`, so with the default the library tries to run a file called `$HOME/rpitx/tune`, literally, and `stdbuf` fails. Treat `GORPITX_PATH` as required.

## ENV and dev mode

`ENV=dev` changes two things:

- `GetInstance` skips the root check.
- `Exec` still validates the arguments and builds the real command line, then runs `sh -c` with a loop that prints `mocking execution of <module> <args>...` once a second instead of the rpitx program. The loop runs until the timeout, a cancelled context or `Stop` ends it.

Nothing transmits in dev mode, and no rpitx binary or script tool has to exist. Tests and development machines use it.

The loop is built by pasting the arguments into the shell script text, without quoting. A text field containing `$(...)`, backticks or a double quote is run by `sh` in dev mode. Do not pass untrusted input to a dev mode instance.

## Root

Outside dev mode, `GetInstance` panics unless the effective user ID is 0:

```text
PIrateRF must be run as root in production mode
```

rpitx drives the Pi's clock hardware through `/dev/mem` and needs root anyway. Run the whole program as root, for example with `sudo`, and pass `GORPITX_PATH` explicitly because `sudo` drops most of the environment:

```bash
sudo GORPITX_PATH=/home/pi/rpitx ./your-program
```

## Files written to /tmp

Importing the package runs an `init` function that writes three shell scripts, embedded from [`scripts/`](https://github.com/psyb0t/gorpitx/tree/main/scripts), and makes them executable (mode `0700`):

| Path | Used by |
| --- | --- |
| `/tmp/fsk.sh` | [`fsk`](fsk.md) |
| `/tmp/audiosock_broadcast.sh` | [`audiosock-broadcast`](audiosock-broadcast.md) |
| `/tmp/modulation.sh` | `audiosock_broadcast.sh` |

This happens in every program that imports gorpitx, in dev mode too, before `main` runs. If a write fails, the package logs the error and calls `os.Exit(1)`. Files left in `/tmp` by a different user cause exactly that; see [Troubleshooting](troubleshooting.md#the-program-exits-before-main). Before a script module runs, `Exec` writes the script again if it has gone missing.

`audiosock_broadcast.sh` also appends to `/tmp/audiosock_broadcast.log`, and `fsk.sh` keeps a temporary WAV file at `/tmp/fsk_<pid>.wav` while it runs.

The paths are fixed and the program runs as root, so another local user who can write to `/tmp` before your program starts can plant a file or symlink there. Run gorpitx on a Pi where you trust every local account.

## Tools the script modules call

| Module | External tools |
| --- | --- |
| [`fsk`](fsk.md) | `minimodem`, `sox`, `$RPITX_PATH/sendiq` |
| [`audiosock-broadcast`](audiosock-broadcast.md) | `socat`, `csdr`, `$RPITX_PATH/sendiq` |

Every other module calls only its own rpitx binary and `stdbuf`, which comes with coreutils.
