# gorpitx

Instructions for coding agents (Codex, Claude Code and the like) working on this repository. People who want to use the library should read the [documentation site](https://psyb0t.github.io/gorpitx/) instead.

## Agent setup

`CLAUDE.md` is a committed symlink to `AGENTS.md`, and `.claude` is a committed symlink to `.agents`, so Claude Code reads the same instructions, rules and skills as Codex. Claude's Edit and Write tools refuse to write through a symlink, so edit `AGENTS.md` and `.agents/rules/*.md` directly.

## What it is

A Go library (`github.com/psyb0t/gorpitx`, one package at the repository root) that runs [rpitx](https://github.com/F5OEO/rpitx) transmitter programs on a Raspberry Pi. A caller passes a module name and a JSON argument blob to `RPITX.Exec`. The module validates the JSON, turns it into a command line (and sometimes stdin), and the library runs it as a child process, one at a time, with optional timeout, graceful stop and output streaming. There is no binary, no Docker image and no network listener.

## Architecture

- `gorpitx.go`: the `RPITX` controller. `GetInstance()` builds one instance through `sync.Once`. `newRPITX()` parses config, panics when not root outside dev mode, and registers the 12 modules in a `map[ModuleName]Module`. `Exec` takes an `atomic.Bool` (`ErrExecuting` when busy), calls `prepareCommand`, starts the process through `commander`, then waits, or calls `waitWithTimeout` when `timeout > 0`. `Stop`, `StreamOutputs` and `StreamOutputsAsync` act on the running process.
- `Module` interface: `ParseArgs(json.RawMessage) ([]string, io.Reader, error)`. Each module file (`pifmrds.go`, `tune.go`, `pichirp.go`, `morse.go`, `pocsag.go`, `spectrumpaint.go`, `ft8.go`, `pisstv.go`, `pirtty.go`, `fsk.go`, `audiosock_broadcast.go`, `sendiq.go`) defines its `ModuleName` constant, its argument struct, `ParseArgs`, `validate*` methods and `buildArgs`.
- Command shape in production: `stdbuf -oL <GORPITX_PATH>/<module name> <args>`. Script modules (`fsk`, `audiosock-broadcast`) run `stdbuf -oL /tmp/<script>.sh <args>` with `RPITX_PATH` set to `GORPITX_PATH`. In dev mode (`ENV=dev`, read by `goenv`) every module runs `sh -c` with an echo loop instead.
- `scripts.go`: embeds `scripts/fsk.sh`, `scripts/audiosock_broadcast.sh` and `scripts/modulation.sh` and writes them to `/tmp` from `init()`, exiting the process if that fails. `EnsureScriptExists` rewrites a missing script before a script module runs.
- `config.go`: `Config{Path}` from `GORPITX_PATH` through `gonfiguration`, default `$HOME/rpitx`. The default is not shell-expanded.
- `errors.go`: package sentinels. Validation failures wrap `commonerrors` sentinels from `github.com/psyb0t/common-go/errors`.
- `utils.go`: frequency conversion and the range check against `minFreqKHz` (5) and `maxFreqKHz` (1500000) from `gorpitx.go`.

Data flow: caller JSON -> `Module.ParseArgs` (unmarshal, validate, build args and stdin) -> `prepareCommand` (binary or script path, `stdbuf`) -> `commander.Start` -> `Wait` or timeout -> `cleanupExecution` kills whatever is left and clears the busy flag.

## Stack

- Go 1.26 (`go 1.26` in `go.mod`), module dependencies vendored in `vendor/`.
- `github.com/psyb0t/commander` (process start, stop with SIGTERM then SIGKILL to the process group, streaming, and a mock for tests), `ctxerrors`, `gonfiguration`, `goenv`, `common-go/errors`, `log/slog`.
- Tests: `testify`. Lint: golangci-lint v2 as a `tool` dependency in `go.mod` (`go tool golangci-lint`), config in `.golangci.yml`, plus `shellcheck` for `scripts/*.sh`.
- Runtime, outside this repository: rpitx binaries; `minimodem`, `sox` for `fsk`; `socat` and `csdr` for `audiosock-broadcast`.

## Layout

| Path | What lives there |
| --- | --- |
| `gorpitx.go` | `RPITX`, `Module`, `GetInstance`, `Exec`, `Stop`, streaming, frequency limits |
| `<module>.go` | One file per module: name constant, argument struct, validation, `buildArgs` |
| `scripts.go`, `scripts/` | Embedded shell scripts for `fsk` and `audiosock-broadcast`, and `modulation.sh` |
| `config.go`, `errors.go`, `utils.go` | Config, sentinels, frequency helpers |
| `*_test.go` | Unit tests beside each file; `gorpitx_integration_test.go` runs real processes in dev mode |
| `.fixtures/` | Audio, RGB and YUV files the tests point modules at |
| `vendor/` | Committed vendored dependencies, never edited by hand |
| `docs/` | The documentation site, published by GitHub Pages from `main` at https://psyb0t.github.io/gorpitx/ |
| `.github/workflows/pipeline.yml` | CI: lint, tests, govulncheck, badges, tag releases |
| `.agents/rules/` | Rules for every change |

## Commands

You need Go 1.26 and Make on the host. `make lint` also needs `shellcheck`. Without a local toolchain, run the same targets in `golang:1.26` with `shellcheck` installed (see `docs/development.md`).

```bash
make help            # every target
make dep             # go mod tidy and go mod vendor
make lint            # go fix -diff, golangci-lint, shellcheck scripts/*.sh
make lint-fix        # go fix and golangci-lint --fix
make test            # go test -race ./...
make test-coverage   # tests with the 90% coverage gate, writes coverage-percent.txt
make docs            # build the documentation site into .testing/docs-site (Docker)
```

`pre-commit.sh` runs `make lint && make test-coverage`.

## Project-specific landmines

- Importing the package writes three scripts to `/tmp` from `init()` and calls `os.Exit(1)` if it cannot. Tests and any program that imports the package do this. With `fs.protected_regular` on, scripts left by one user (for example a root run) make the write fail for another, and the test binary exits with status 1; delete `/tmp/fsk.sh`, `/tmp/audiosock_broadcast.sh` and `/tmp/modulation.sh`.
- `newRPITX` panics unless the effective UID is 0 or `ENV=dev`. Tests build `&RPITX{...}` directly with `commander.NewMock()` or `commander.New()` rather than calling `GetInstance()`.
- `pifmrds` takes `freq` in MHz with 0.1 MHz precision. Every other module takes Hz, and `sendiq` names the field `freq` while the rest use `frequency`.
- `ModuleName` values are rpitx binary names (`pift8`, not `ft8`), because the production path is `GORPITX_PATH` joined with the name. The constant for SSTV is `ModuleNamePISSSTV` (three S), keep it for compatibility.
- `sendiq` clamps `power` into 0 to 7 inside `validatePower` instead of rejecting it, mirroring the rpitx binary.
- The module structs in the `modules` map are reused across `Exec` calls and `ParseArgs` unmarshals into them without resetting, so a field left out of a call keeps the previous call's value. Documented in `docs/running-modules.md`; a fix must update that section.
- `getMockExecCmd` pastes the arguments unquoted into the `sh -c` script, so dev mode runs shell syntax found in text fields. Production passes an argument slice and is not affected.
- `POCSAGMessage.FunctionBits` is validated but `buildStdin` never writes it. Documented on `docs/pocsag.md`.

## Rules

Rules for every change: `.agents/rules/`. Read them before editing.
