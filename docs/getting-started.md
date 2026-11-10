---
title: "Getting started"
description: "Install rpitx and the extra tools on a Raspberry Pi, add gorpitx to a Go program, try it in dev mode, then send a first real transmission as root."
nav_order: 2
---

# Getting started

From nothing to a Morse transmission. The first run happens in dev mode on any Linux machine, so you can see the whole flow before a Pi is involved.

## 1. Add the library

```bash
go get github.com/psyb0t/gorpitx
```

You need Go 1.26 or newer.

## 2. Write the program

This program sends a Morse message, prints whatever the process writes, and gives up after three seconds.

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/gorpitx"
)

func main() {
	rpitx := gorpitx.GetInstance()

	args, err := json.Marshal(gorpitx.MORSE{
		Frequency: 14070000, // Hz
		Rate:      20,       // dits per minute
		Message:   "CQ CQ DE N0CALL",
	})
	if err != nil {
		log.Fatal(err)
	}

	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	rpitx.StreamOutputsAsync(stdout, stderr)

	go func() {
		for line := range stdout {
			fmt.Println("stdout:", line)
		}
	}()

	go func() {
		for line := range stderr {
			fmt.Println("stderr:", line)
		}
	}()

	err = rpitx.Exec(
		context.Background(),
		gorpitx.ModuleNameMORSE,
		args,
		3*time.Second,
	)

	switch {
	case err == nil:
		fmt.Println("transmission finished")
	case errors.Is(err, commonerrors.ErrTimeout):
		fmt.Println("stopped after the timeout")
	default:
		log.Fatal(err)
	}
}
```

`commonerrors` is `github.com/psyb0t/common-go/errors`, which gorpitx already depends on.

## 3. Run it in dev mode

```bash
ENV=dev go run .
```

With `ENV=dev`, gorpitx skips the root check and replaces the rpitx program with a shell loop that prints one line a second, so the output looks like this:

```text
stdout: mocking execution of morse 14070000 20 CQ CQ DE N0CALL...
stdout: mocking execution of morse 14070000 20 CQ CQ DE N0CALL...
stdout: mocking execution of morse 14070000 20 CQ CQ DE N0CALL...
2026/01/01 12:00:03 WARN failed to gracefully stop process after timeout err="failed to stop process: terminated [...]"
stopped after the timeout
```

The `WARN` line comes from the library's timeout path. It logs the `terminated` result of a successful graceful stop as a failure, so you can ignore it.

The loop never ends by itself, so in dev mode every `Exec` ends through its timeout, a cancelled context or [`Stop`](running-modules.md#stopping-a-transmission). The arguments in the line are the exact ones the real program would get. Argument validation runs in dev mode too, so a bad frequency fails here the same way it fails on the Pi.

## 4. Install rpitx on the Pi

Use a Raspberry Pi model that [rpitx](https://github.com/F5OEO/rpitx) supports, with Raspberry Pi OS.

```bash
sudo apt update
git clone https://github.com/F5OEO/rpitx.git
cd rpitx
./install.sh
```

Then install the tools the script-based modules call. Skip the ones you do not use.

| Module | Needs |
| --- | --- |
| [`fsk`](fsk.md) | `minimodem`, `sox` |
| [`audiosock-broadcast`](audiosock-broadcast.md) | `socat`, and `csdr` on the `PATH` (rpitx's installer builds `csdr`) |

```bash
sudo apt install minimodem sox socat
```

## 5. Point gorpitx at rpitx

gorpitx runs `<GORPITX_PATH>/<module name>`. Set `GORPITX_PATH` to the directory that holds the rpitx binaries, as an absolute path:

```bash
export GORPITX_PATH=/home/pi/rpitx
```

Do not rely on the default. It is the literal string `$HOME/rpitx`, and nothing expands `$HOME` in it. See [Configuration](configuration.md#gorpitx_path).

## 6. Transmit for real

Build the program and run it as root without `ENV`. Pass `GORPITX_PATH` through, because `sudo` drops most of the environment.

```bash
go build -o morse .
sudo GORPITX_PATH=/home/pi/rpitx ./morse
```

Root is required: `GetInstance` panics with `PIrateRF must be run as root in production mode` otherwise. The message names another project, but the check is this one. See [Troubleshooting](troubleshooting.md#getinstance-panics-about-root).

Before you key up anything, check that the frequency is one you may transmit on, and put a filter between the Pi's GPIO pin and any antenna.

## Next

- [Running modules](running-modules.md) covers timeouts, stopping, streaming and the one-at-a-time rule.
- [Modules](modules.md) lists every field for all 12 modules.
