# Testing

Prefix: TEST. How to run the suites is in `AGENTS.md` and `docs/development.md`.

- **TEST1:** Tests sit beside the code in `package gorpitx`. Each module test covers valid input, every validation branch, and the exact argument slice `buildArgs` returns.
- **TEST2:** Execution tests construct `&RPITX{...}` directly. Production-path tests use `commander.NewMock()` with `Expect("stdbuf", "-oL", ...)` and `t.Setenv(goenv.EnvVarName, goenv.Prod)`. Tests that need a real process use `commander.New()` with `t.Setenv(goenv.EnvVarName, goenv.Dev)`, so the module runs the `sh -c` echo loop and nothing transmits.
- **TEST3:** Files a module must find on disk come from `.fixtures/`. Do not point tests at paths outside the repository, except the `/tmp` script paths `scripts.go` owns.
- **TEST4:** The coverage floor is 90% (`MIN_TEST_COVERAGE` in the `Makefile`).
- **TEST5:** No test may run an rpitx binary or need root. CI runners have neither.
