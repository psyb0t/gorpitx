---
title: "Running modules"
description: "How Exec runs one gorpitx module at a time: arguments, timeouts, context cancellation, Stop, output streaming, and what each call returns."
nav_order: 4
---

# Running modules

Everything goes through one `*gorpitx.RPITX` value. You get it from `GetInstance`, pass a module name and JSON arguments to `Exec`, and optionally stream output or stop the process from another goroutine. The signatures are in [`gorpitx.go`](https://github.com/psyb0t/gorpitx/blob/main/gorpitx.go).

```go
func GetInstance() *RPITX

func (r *RPITX) Exec(ctx context.Context, name ModuleName, args []byte, timeout time.Duration) error
func (r *RPITX) Stop(ctx context.Context) error
func (r *RPITX) StreamOutputs(stdout, stderr chan<- string)
func (r *RPITX) StreamOutputsAsync(stdout, stderr chan<- string)
func (r *RPITX) GetSupportedModules() []ModuleName
func (r *RPITX) IsSupportedModule(name ModuleName) bool
```

## Getting the instance

`GetInstance` builds the instance on its first call and returns the same pointer forever after. The first call reads [`GORPITX_PATH`](configuration.md#gorpitx_path) and, outside [dev mode](configuration.md#env-and-dev-mode), panics if the program is not running as root. It also panics if the config cannot be parsed.

`GetSupportedModules` returns the 12 registered names in no fixed order (it walks a map). `IsSupportedModule` checks one name. Both take the rpitx binary name, so `IsSupportedModule("pift8")` is true and `IsSupportedModule("ft8")` is false. Use the `ModuleName*` constants rather than strings.

## Exec

```go
args, err := json.Marshal(gorpitx.PICHIRP{
    Frequency: 434000000, // Hz
    Bandwidth: 100000,    // Hz
    Time:      5,         // seconds
})
if err != nil {
    return err
}

err = rpitx.Exec(ctx, gorpitx.ModuleNamePICHIRP, args, 0)
```

`args` is JSON. Marshal the module's struct as above, or send JSON you received from somewhere else; the field names are on each [module page](modules.md). Unknown fields are ignored.

`Exec` blocks, and it goes through these steps in order:

1. If another `Exec` is running, it returns `ErrExecuting` at once. Only one module runs at a time.
2. It looks up the module (`ErrUnknownModule` if the name is not registered) and calls its `ParseArgs`, which unmarshals and validates the JSON. A validation error comes back wrapped with `failed to parse args`, and nothing starts.
3. It starts the process: `stdbuf -oL <GORPITX_PATH>/<name> <args>`, a script under `/tmp` for `fsk` and `audiosock-broadcast`, or the shell loop in dev mode. Modules that feed the program on stdin (`pocsag`, `fsk`) pass that data here.
4. It waits for the process to exit, then kills anything still running in it and clears the busy flag.

The `timeout` argument decides how the wait ends:

| `timeout` | What happens |
| --- | --- |
| `0` | `Exec` waits until the process exits on its own, or until `ctx` is cancelled. |
| `> 0` | When the timeout passes, `Exec` sends SIGTERM to the process group, gives it 3 seconds, sends SIGKILL if it is still alive, and returns `commonerrors.ErrTimeout`. |

If the program exits by itself before the timeout, `Exec` returns `nil` for status 0. A non-zero status comes back wrapped with `failed to wait for process`, as `commonerrors.ErrFailed` with the exit code and the program's stderr in the message.

Cancelling `ctx` kills the process. `Exec` then returns an error that wraps `commonerrors.ErrKilled`.

Some programs never exit by themselves: `tune` holds a carrier, `audiosock-broadcast` streams until the socket closes, `sendiq` with `loopMode` repeats the file, and every module in dev mode loops. Give those a timeout, a cancellable context, or call `Stop`.

### Arguments persist between calls

Each module is one struct inside the instance, and `ParseArgs` unmarshals into it without resetting it first. A field you set in one call and leave out of the next keeps its old value:

| Call | JSON sent to `tune` | Command line |
| --- | --- | --- |
| First | `{"frequency": 434000000, "ppm": 2.5}` | `tune -f 434000000 -p 2.5` |
| Second | `{"frequency": 144000000}` | `tune -f 144000000 -p 2.5` |

Until that changes, send every field you care about on each call. To clear an optional field, send it as `null`, or as `""` or `false` for the plain string and bool fields such as `pi` or `loopMode`.

## Stopping a transmission

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

err := rpitx.Stop(ctx)
```

`Stop` sends SIGTERM to the running process group and waits for it to exit. When `ctx` has no deadline it waits 3 seconds. If the process is still alive after that, it gets SIGKILL. Call it from another goroutine while `Exec` is blocked.

What it returns:

| Situation | `Stop` returns |
| --- | --- |
| Nothing is running | `ErrNotExecuting` |
| The process exited on SIGTERM | an error wrapping `commonerrors.ErrTerminated` |
| The process needed SIGKILL | an error wrapping `commonerrors.ErrKilled` |
| The process exited cleanly | `nil` |

A successful graceful stop is reported as an error wrapping `ErrTerminated`, so check for it with `errors.Is` instead of treating every non-nil result as a failure. The blocked `Exec` returns an error wrapping the same sentinel.

## Streaming output

Each line the process writes to stdout or stderr can go to a channel. There are two ways to attach one.

`StreamOutputsAsync` can be called before `Exec`. It starts a goroutine that waits until a process is running and then attaches the channels:

```go
stdout := make(chan string, 100)
stderr := make(chan string, 100)

rpitx.StreamOutputsAsync(stdout, stderr)

go func() {
    for line := range stdout {
        fmt.Println("stdout:", line)
    }
}()

err := rpitx.Exec(ctx, gorpitx.ModuleNameMORSE, args, 30*time.Second)
```

`StreamOutputs` attaches the channels right now, and only works while a process is running. Called while idle, it logs a warning and does nothing, so you have to call it from another goroutine after `Exec` has started:

```go
go func() {
    err := rpitx.Exec(ctx, gorpitx.ModuleNameMORSE, args, 30*time.Second)
    if err != nil {
        log.Println(err)
    }
}()

time.Sleep(100 * time.Millisecond)
rpitx.StreamOutputs(stdout, stderr)
```

The sleep is a guess, which is why `StreamOutputsAsync` is the better choice.

Rules that hold for both:

- The channels are closed when the process exits, so a channel pair serves one `Exec`. Make new channels for the next one.
- Use buffered channels and read them promptly. If a send blocks for more than 100 ms, that channel is dropped from the stream and never closed, and a `range` over it hangs.
- Lines written before the channels are attached are not replayed.
- `StreamOutputsAsync` polls every 10 ms until a module starts. If no `Exec` follows, its goroutine keeps polling for the life of the program.
- The production command runs under `stdbuf -oL`, so the program's stdout arrives line by line instead of in 4 KB blocks.

## Concurrency

The instance is safe to share between goroutines. `Exec` uses an atomic flag, so two concurrent calls never both start a process: the second gets `ErrExecuting`. `Stop` and the stream methods read the current process under a lock. There is no queue; if you want transmissions to wait their turn, serialize the calls yourself.
