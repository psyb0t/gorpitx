# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.5 — 2026-07-27

- Added a GitHub Actions CI status badge to the README.

## v0.1.4 — 2026-07-27

Fix badges CI job — add needs dependency so the coverage badge waits for the coverage artifact.

- Added `needs: [call-go-workflow]` to the `badges` job in `.github/workflows/pipeline.yml`. Previously `badges` had no dependency on `call-go-workflow`, so it could start before the `coverage-percent.txt` artifact produced by `make test-coverage` existed, breaking the coverage badge.

## v0.1.3 — 2026-07-27

Modernize to Go 1.26 and switch to structured logging.

- Go 1.26; dropped the `modernize` dev tool from `go.mod` (its checks are now covered by `go fix`).
- `make lint` now runs `go fix -diff ./...` first (fails with a pointer to `make lint-fix` if it finds anything) before `golangci-lint`.
- Added a coverage badge (`make test-coverage` now writes `coverage-percent.txt`; wired into the `badges` CI job and README).
- Replaced `github.com/sirupsen/logrus` with the standard library `log/slog` for structured logging throughout.
- Error wrapping already used `github.com/psyb0t/ctxerrors` consistently; no changes needed there.

## v0.1.2 — 2026-07-27

Add README status badges.

- Added self-hosted version and license badges (rendered as SVGs on the `badges` branch by the `create-badges` CI job, no third-party render service). Wired a badges job into pipeline.yml.
