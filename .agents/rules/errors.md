# Errors

Prefix: ERR.

- **ERR1:** Validation errors wrap a `commonerrors` sentinel from `github.com/psyb0t/common-go/errors`: `ErrRequiredFieldNotSet` with the JSON field name as the message, `ErrInvalidValue` with what was wrong and the value received, `ErrFileNotFound` with the path.
- **ERR2:** Package sentinels live in `errors.go`. Add one only when a caller needs to tell the case apart with `errors.Is`; otherwise wrap a `commonerrors` sentinel. Existing exported sentinels are API and keep their names.
- **ERR3:** Wrap every returned error with `ctxerrors.Wrap` or `ctxerrors.Wrapf` and a short phrase. Return a validation error from `validate` unwrapped a second time, as the modules do.
- **ERR4:** A timeout in `Exec` returns `commonerrors.ErrTimeout`, also when the stopped process reports `ErrTerminated` or `ErrKilled`. Keep that contract; callers test for it.
- **ERR5:** Log with `log/slog` and lowercase keys (`"module"`, `"args"`, `"err"`). Do not log with `fmt` or `log`.
