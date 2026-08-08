# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.7 — 2026-08-08

Repository infrastructure only. No library code changed.

- Added the imported-by badge: a count of the public packages importing this
  module, linking to `importers.md` on the `badges` branch — the importing
  repositories, grouped, package counts descending, and flagged when the owner
  differs from this repo's.
- It measures **blast radius, not adoption**. The number tells you how much
  breaks when an exported name moves; the external mark tells you whether any of
  that is someone else's problem, which is what decides how strictly the module
  has to be versioned.
- Refreshed weekly rather than daily, because pkg.go.dev's crawl lags
  publication by days and each run drags the full test suite along (the badges
  job needs the coverage artifact). The whole pipeline runs rather than a
  badges-only job: the badge publisher republishes only what a run produced, so
  a badge-only refresh would delete the coverage, version and license badges.
- The cron slot is derived from a hash of the repository name rather than
  chosen — GitHub cron has no randomness, and its scheduler sheds queued runs
  hardest at the round times a human would pick.

## v0.1.6 — 2026-08-01

Repository infrastructure only, no library change.

- Every push now mirrors the repo to GitLab and Codeberg, so the source stays fetchable if GitHub is unavailable. Gitee is wired but left off — it binds repo creation to a mobile number and silently creates the repo private without one.
- Pushes to the default branch and every tag are archived to the Wayback Machine and Software Heritage, through the authenticated Save Page Now API, with README outlinks captured too. Feature-branch pushes are skipped because the archive is rate-limited.
- Issues filed on the Codeberg and GitLab mirrors are pulled back into GitHub every six hours, so a bug reported on a mirror reaches the same tracker. Scheduled runs jitter to avoid stampeding the mirrors; a manual run does not.
- Pull requests are auto-closed and locked with a pointer to the issue tracker (this landed earlier but was never written down).
- Note on tags: `v0.1.2` through `v0.1.5` were released as commits and changelog entries but never actually tagged, so `v0.1.6` is the first tag since `v0.1.1` and carries all of that work — the Go 1.26 bump, the `logrus` → `log/slog` switch and the badges — along with the CI changes above.

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
