---
title: "Development"
description: "Build, lint and test gorpitx with its Make targets, the 90 percent coverage gate, the docs build, CI and releases, and how to add a module."
nav_order: 9
---

# Development

Working on gorpitx itself. You need Go 1.26 or newer, Make, and `shellcheck` for the lint target. golangci-lint comes from the module's own `tool` dependency in [`go.mod`](https://github.com/psyb0t/gorpitx/blob/main/go.mod), and every dependency is vendored in [`vendor/`](https://github.com/psyb0t/gorpitx/tree/main/vendor). No test needs a Pi, rpitx or root.

## Make targets

Run `make help` for the list. The [`Makefile`](https://github.com/psyb0t/gorpitx/blob/main/Makefile) has:

| Target | What it does |
| --- | --- |
| `make all` | Runs `dep`, `lint` and `test`, in that order |
| `make dep` | `go mod tidy` and `go mod vendor` |
| `make lint` | Fails if `go fix -diff ./...` has suggestions, runs `go tool golangci-lint run ./...` with the rules in [`.golangci.yml`](https://github.com/psyb0t/gorpitx/blob/main/.golangci.yml), then `shellcheck scripts/*.sh` |
| `make lint-fix` | Applies `go fix ./...`, runs golangci-lint with `--fix`, then shellcheck |
| `make test` | `go test -race ./...` |
| `make test-coverage` | Tests with a coverage profile; fails under 90% and writes the percentage to `coverage-percent.txt` |
| `make docs` | Builds this site into `.testing/docs-site` with GitHub's Pages build image (needs Docker) |

[`pre-commit.sh`](https://github.com/psyb0t/gorpitx/blob/main/pre-commit.sh) runs `make lint && make test-coverage`.

## Without a local Go toolchain

The same targets run in the official Go image. Install shellcheck inside it first:

```bash
docker run --rm \
  -v "$PWD:/src" \
  -w /src \
  golang:1.26 \
  bash -c 'apt-get update -qq \
    && apt-get install -y -qq shellcheck \
    && make lint test-coverage'
```

## Tests

Unit tests sit next to the file they test (`tune.go`, `tune_test.go`) and are table-driven with testify. Each module's tests feed it JSON and check either the exact argument slice or the error. Files a module must find on disk come from [`.fixtures/`](https://github.com/psyb0t/gorpitx/tree/main/.fixtures).

Execution tests build an `&RPITX{...}` directly instead of calling `GetInstance`, so they need no root. Production-path tests use `commander.NewMock()` and assert the exact `stdbuf -oL ...` command. [`gorpitx_integration_test.go`](https://github.com/psyb0t/gorpitx/blob/main/gorpitx_integration_test.go) uses a real `commander.New()` in dev mode, so real processes start, stream and stop, and nothing transmits.

Importing the package writes its scripts to `/tmp`, so the test process needs a writable `/tmp`.

## Adding a module

1. Create `<name>.go` with a `ModuleName<NAME>` constant set to the rpitx binary name, an argument struct with lowerCamelCase JSON tags and pointer types for optional fields, and `ParseArgs`, `validate` and `buildArgs` methods, following [`tune.go`](https://github.com/psyb0t/gorpitx/blob/main/tune.go).
2. Validate the frequency with `isValidFreqHz` and wrap `ErrFreqOutOfRange` as the other modules do.
3. Register it in the `modules` map in `newRPITX` in [`gorpitx.go`](https://github.com/psyb0t/gorpitx/blob/main/gorpitx.go).
4. Write `<name>_test.go` covering every validation branch and the argument slice. Keep coverage at 90% or above.
5. Add a page under `docs/`, a row on the [Modules](modules.md) page, and the module to the README list.

## Documentation

This site is the `docs/` directory, built by GitHub Pages from `main` with Jekyll and the just-the-docs theme. Every page needs front matter with a unique `title` and a `description` of 50 to 170 characters. Link between pages with the relative `.md` path; link to a file outside `docs/` with its full GitHub URL. Run `make docs` and open `.testing/docs-site` to check a change.

## CI and releases

[`.github/workflows/pipeline.yml`](https://github.com/psyb0t/gorpitx/blob/main/.github/workflows/pipeline.yml) calls the shared Go workflow from [psyb0t/reusable-github-workflows](https://github.com/psyb0t/reusable-github-workflows):

- Every branch push runs `make lint`, `make test-coverage` and govulncheck, using the committed `vendor/` directory.
- A push to `main` also refreshes the coverage, version, license and imported-by badges on the `badges` branch. A weekly scheduled run of the whole pipeline keeps the imported-by count current.
- A tag does not rerun the checks. It waits for the `main` run of the same commit to pass, creates the GitHub release, and refreshes the badges with that run's coverage.

Release notes for each version are in [`CHANGELOG.md`](https://github.com/psyb0t/gorpitx/blob/main/CHANGELOG.md).
