# Embedded scripts

Prefix: SCR. `scripts/*.sh` are embedded with `go:embed` in `scripts.go` and run as root on the Pi.

- **SCR1:** A script module receives its arguments positionally from `buildArgs` and the rpitx directory in the `RPITX_PATH` environment variable set in `startProcess`. Change the argument order in the script and in `buildArgs` together, and update the usage comment at the top of the script.
- **SCR2:** Scripts are written to fixed paths under `/tmp` (`fskScriptPath`, `audioSockBroadcastPath`, `modulationPath`) with mode `0700`. A new script gets a path constant, an embed variable, a write in `writeAllScripts`, and a case in `ModuleNameToScriptName` and `getScriptContent`.
- **SCR3:** Script arguments, such as the socket path, come straight from caller JSON.
- **SCR4:** A script's external tools (`minimodem`, `sox`, `socat`, `csdr`) are listed in `docs/getting-started.md` and on the module's docs page.
