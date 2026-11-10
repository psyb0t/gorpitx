# Modules

Prefix: MOD. A module is one rpitx program wrapped behind the `Module` interface in `gorpitx.go`.

- **MOD1:** A module lives in its own file named after it (`tune.go`) with its tests beside it (`tune_test.go`). The file declares a `ModuleName<NAME>` constant whose value is the rpitx binary name, because `prepareCommand` joins `GORPITX_PATH` with that value. Script modules return their path from `ModuleNameToScriptName` instead.
- **MOD2:** `ParseArgs` does three things in order: `json.Unmarshal` into the receiver (wrapped with `ctxerrors.Wrap(err, "failed to unmarshal args")`), `validate()`, then `buildArgs()` and, when the binary reads stdin, a reader (`POCSAG.buildStdin`, `FSK.prepareStdin`). It never starts a process.
- **MOD3:** `validate()` calls one `validate<Field>` method per field, in declaration order, and returns the first error. Required fields come first, then ranges and formats, then file existence with `os.Stat`.
- **MOD4:** Frequencies are Hz and checked with `isValidFreqHz`, wrapping `ErrFreqOutOfRange` with the `"(%d kHz to %.0f MHz), got: %f Hz"` message the other modules use. `pifmrds` is the one MHz module, and it also checks `hasValidFreqPrecision`. Do not add a second unit.
- **MOD5:** Optional fields are pointers with `omitempty` (`*int`, `*float64`, `*bool`, `*string`), so "not set" differs from zero. `buildArgs` adds a flag only when the pointer is non-nil, or passes a documented default when the binary takes positional arguments (`PIRTTY`, `FSK`, `AudioSockBroadcast`).
- **MOD6:** JSON field names are lowerCamelCase and are part of the public API. Do not rename one; add a field instead.
- **MOD7:** A new module is registered in the `modules` map in `newRPITX()`, gets a docs page under `docs/` and a row in `docs/modules.md`, and appears in the README module list.
