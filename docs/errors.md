---
title: "Errors"
description: "Every error gorpitx returns, the sentinels to match with errors.Is, and what the error text looks like with its call-site suffix."
nav_order: 6
---

# Errors

gorpitx returns wrapped sentinel errors. Match them with `errors.Is`, never by comparing strings. The package's own sentinels are in [`errors.go`](https://github.com/psyb0t/gorpitx/blob/main/errors.go); the rest come from `github.com/psyb0t/common-go/errors`, usually imported as `commonerrors`.

## Execution errors

| Sentinel | Text | Returned when |
| --- | --- | --- |
| `gorpitx.ErrExecuting` | `RPITX is busy executing another command` | `Exec` is called while another module runs |
| `gorpitx.ErrUnknownModule` | `unknown module` | The module name is not registered, wrapped with the name |
| `gorpitx.ErrNotExecuting` | `RPITX is not executing a command` | `Stop` is called while nothing runs |
| `commonerrors.ErrTimeout` | `timeout` | The `Exec` timeout passed and the process was stopped |
| `commonerrors.ErrTerminated` | `terminated` | `Stop` ended the process with SIGTERM; both `Stop` and the blocked `Exec` return it wrapped |
| `commonerrors.ErrKilled` | `killed` | The process needed SIGKILL, or `ctx` was cancelled during `Exec` |
| `commonerrors.ErrFailed` | `failed` | The program exited by itself with a non-zero status |

A process that exits by itself with a non-zero status comes back from `Exec` wrapped with `failed to wait for process`, as `commonerrors.ErrFailed` (`failed`) with the exit code and the process's stderr in the message: `(exit 127): <stderr text>`. A process that cannot start at all, for example because `stdbuf` is missing, comes back wrapped with `failed to start process`. A missing rpitx binary shows up differently: `stdbuf` starts, cannot run it, and exits with status 127 (see [Troubleshooting](troubleshooting.md#exec-fails-with-exit-status-127)).

## Validation errors

Validation runs inside `Exec` before anything starts, and the error comes back wrapped with `failed to parse args`.

| Sentinel | Text | Returned when |
| --- | --- | --- |
| `commonerrors.ErrRequiredFieldNotSet` | `required field is not set` | A required field is missing or blank, wrapped with the JSON field name, such as `message` or `messages` |
| `commonerrors.ErrInvalidValue` | `invalid value` | A value is out of range or not in the allowed set, wrapped with the rule and the value you sent |
| `commonerrors.ErrFileNotFound` | `file not found` | A file field points at a path that does not exist |
| `gorpitx.ErrFreqOutOfRange` | `frequency out of RPiTX range` | The frequency is outside 5 kHz to 1500 MHz |
| `gorpitx.ErrFreqPrecision` | `frequency precision too high` | A `pifmrds` frequency has more than one decimal place |
| `gorpitx.ErrPIInvalidHex` | `PI code must be valid hex` | A `pifmrds` PI code has 4 characters that are not hex |
| `gorpitx.ErrPSTooLong` | `PS text must be 8 characters or less` | A `pifmrds` station name is too long |

A missing frequency is reported as `ErrInvalidValue` (`frequency must be positive, got: 0.000000`) in every module except `pifmrds`, where it is `ErrRequiredFieldNotSet`. JSON that does not fit the struct, such as a string where a number belongs, comes back wrapped with `failed to unmarshal args`.

## Reading the text

Errors are wrapped with [ctxerrors](https://github.com/psyb0t/ctxerrors), so each wrap adds the file, line and function where it happened. A refused `tune` frequency reads like this, on one line:

```text
failed to parse args: (5 kHz to 1500 MHz), got: 4000.000000 Hz: frequency out of RPiTX range [/path/to/gorpitx/tune.go:90 in github.com/psyb0t/gorpitx.(*TUNE).validateFreq] [/path/to/gorpitx/gorpitx.go:167 in github.com/psyb0t/gorpitx.(*RPITX).prepareCommand]
```

That suffix is for your logs. Do not show the raw text to end users, and do not parse it.

## Matching in code

```go
err := rpitx.Exec(ctx, gorpitx.ModuleNameTUNE, args, 10*time.Second)

switch {
case err == nil:
    // finished on its own
case errors.Is(err, commonerrors.ErrTimeout):
    // ran for the full 10 seconds, then stopped
case errors.Is(err, gorpitx.ErrExecuting):
    // something else is transmitting, try later
case errors.Is(err, commonerrors.ErrInvalidValue),
    errors.Is(err, commonerrors.ErrRequiredFieldNotSet),
    errors.Is(err, commonerrors.ErrFileNotFound),
    errors.Is(err, gorpitx.ErrFreqOutOfRange):
    // bad arguments, nothing was transmitted
default:
    // the process failed
}
```
