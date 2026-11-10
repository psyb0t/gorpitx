---
title: "Architecture"
description: "How a gorpitx Exec call becomes an rpitx process: the singleton, the Module interface, command building, scripts under /tmp and process control."
nav_order: 8
---

# Architecture

gorpitx is one Go package at the root of the repository, with no binary and no network listener. This page follows one `Exec` call from JSON to a running rpitx program, then lists where each piece lives.

## The path of one call

```text
caller JSON
  -> Exec: atomic busy flag, or ErrExecuting
  -> Module.ParseArgs: unmarshal, validate, build args (and stdin)
  -> prepareCommand: binary path, script path or dev mode loop, wrapped in stdbuf -oL
  -> commander.Start: child process in its own process group
  -> Wait, or waitWithTimeout: SIGTERM, 3 s, SIGKILL, ErrTimeout
  -> cleanupExecution: kill what is left, clear the busy flag
```

## Pieces

**The instance.** `RPITX` in [`gorpitx.go`](https://github.com/psyb0t/gorpitx/blob/main/gorpitx.go) holds the config, a `commander.Commander`, the module map, an `atomic.Bool` busy flag, and the current process behind an `RWMutex`. `GetInstance` builds it once with `sync.Once`. There is one transmitter on a Pi, so there is one instance.

**Modules.** Every module satisfies one interface:

```go
type Module interface {
    ParseArgs(json.RawMessage) ([]string, io.Reader, error)
}
```

`ParseArgs` unmarshals the JSON into the module's struct, validates each field, and returns the program's arguments plus an optional stdin reader. `pocsag` returns its `address:message` lines as stdin, `fsk` returns the text or the opened file plus a newline, and the others return `nil`. The 12 modules are registered in a map in `newRPITX`, keyed by the rpitx binary name. Because the map holds one struct per module and `ParseArgs` unmarshals into it, values from a previous call survive into the next one; see [Running modules](running-modules.md#arguments-persist-between-calls).

**Command building.** `prepareCommand` picks one of three shapes:

| Mode | Command |
| --- | --- |
| Production, binary module | `stdbuf -oL <GORPITX_PATH>/<name> <args>` |
| Production, script module | `stdbuf -oL /tmp/<script>.sh <args>`, with `RPITX_PATH=<GORPITX_PATH>` in the environment |
| Dev mode (`ENV=dev`) | `sh -c 'while true; do echo "mocking execution of <name> <args>..."; sleep 1; done'` |

`stdbuf -oL` makes the program flush stdout per line, so streamed output arrives as it happens. In production, arguments are passed as an argument list, never through a shell, so a message with spaces or quotes reaches the program intact, and the script modules quote every expansion. Dev mode is the exception: it pastes the arguments into the `sh -c` script text, so a message containing `$(...)` or a double quote runs as shell code. Feed dev mode only input you trust.

**Embedded scripts.** [`scripts.go`](https://github.com/psyb0t/gorpitx/blob/main/scripts.go) embeds the three files in [`scripts/`](https://github.com/psyb0t/gorpitx/tree/main/scripts) with `go:embed`, writes them to fixed paths under `/tmp` from the package `init`, and exits the program if it cannot. `EnsureScriptExists` writes a script again before a script module runs if it has gone missing. `audiosock_broadcast.sh` pipes `socat` into `modulation.sh`, which picks a csdr chain per modulation, and on into `sendiq`. `fsk.sh` runs `minimodem`, then `sox`, then `sendiq`.

**Process control.** [commander](https://github.com/psyb0t/commander) starts each process in its own process group. `Stop` and the timeout path send SIGTERM to the whole group, wait (3 seconds by default), then send SIGKILL, so the children of a script die with it. Cancelling the `Exec` context kills the process through `exec.CommandContext`. Output goes to any attached channels, which are closed when the process exits.

**Config and errors.** [`config.go`](https://github.com/psyb0t/gorpitx/blob/main/config.go) reads `GORPITX_PATH` with gonfiguration. [goenv](https://github.com/psyb0t/goenv) reads `ENV`. [`errors.go`](https://github.com/psyb0t/gorpitx/blob/main/errors.go) holds the package sentinels, and validation wraps the `common-go` sentinels with [ctxerrors](https://github.com/psyb0t/ctxerrors). [`utils.go`](https://github.com/psyb0t/gorpitx/blob/main/utils.go) has the frequency conversions and the range check.

## Dependencies

| Module | Used for |
| --- | --- |
| [`github.com/psyb0t/commander`](https://github.com/psyb0t/commander) | Starting, stopping and streaming processes, and a mock for tests |
| [`github.com/psyb0t/ctxerrors`](https://github.com/psyb0t/ctxerrors) | Error wrapping with call sites |
| [`github.com/psyb0t/gonfiguration`](https://github.com/psyb0t/gonfiguration) | Reading `GORPITX_PATH` |
| [`github.com/psyb0t/goenv`](https://github.com/psyb0t/goenv) | Dev or production mode from `ENV` |
| [`github.com/psyb0t/common-go`](https://github.com/psyb0t/common-go) | Shared sentinel errors |
| `log/slog` | Debug and warning logs |
