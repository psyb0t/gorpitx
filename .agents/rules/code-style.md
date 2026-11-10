# Code style

Prefix: STY. `.golangci.yml` enforces most of this; `make lint` must pass with zero issues.

- **STY1:** Return early. Prefer `continue` in loops and `switch` over `if`/`else` chains.
- **STY2:** Lines stay within 80 columns (`lll` with tab width 1). Split long calls one argument per line, as the existing code does.
- **STY3:** Name numbers. A limit or default is a constant next to the code that uses it (`psMaxLength`, `defaultFSKBaudRate`, `ft8OffsetMax`), not a literal in the check.
- **STY4:** Shared state goes through `atomic.Bool` (`isExecuting`) and `processMu`. Read `r.process` only under the lock, as `Stop` and `StreamOutputs` do.
- **STY5:** A `//nolint` names the linter and gives the reason after `//`.
- **STY6:** Imports outside the standard library must pass the `depguard` allow list in `.golangci.yml`. Add a dependency only when the existing ones (`commander`, `ctxerrors`, `gonfiguration`, `goenv`, `common-go`) do not cover it, then run `make dep` and commit `vendor/`.
