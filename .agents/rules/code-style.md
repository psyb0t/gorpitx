# Code style

Prefix: STY. `.golangci.yml` enforces most of this through `make lint`.

- **STY2:** Lines stay within 80 columns (`lll` with tab width 1).
- **STY4:** Shared state goes through `atomic.Bool` (`isExecuting`) and `processMu`. Read `r.process` only under the lock, as `Stop` and `StreamOutputs` do.
- **STY6:** This repository keeps `depguard`: imports outside the standard library must pass its allow list in `.golangci.yml`. The existing dependencies are `commander`, `ctxerrors`, `gonfiguration`, `goenv` and `common-go`. After adding one, run `make dep` and commit `vendor/`.
