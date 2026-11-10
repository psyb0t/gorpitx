# Errors

Prefix: ERR.

- **ERR1:** Validation errors wrap a `commonerrors` sentinel from `github.com/psyb0t/common-go/errors`: `ErrRequiredFieldNotSet` with the JSON field name as the message, `ErrInvalidValue` with what was wrong and the value received, `ErrFileNotFound` with the path.
- **ERR2:** Package sentinels live in `errors.go`. Existing exported sentinels are API and keep their names.
- **ERR3:** Return a validation error from `validate` without wrapping it a second time, as the modules do.
- **ERR4:** A timeout in `Exec` returns `commonerrors.ErrTimeout`, also when the stopped process reports `ErrTerminated` or `ErrKilled`. Keep that contract; callers test for it.
- **ERR5:** Log keys are lowercase: `"module"`, `"args"`, `"err"`.
