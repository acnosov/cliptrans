# AGENTS.md

Go CLI that translates the Wayland clipboard via LLM API. No GUI.
Behavior source of truth: `SPEC.md` — read it before changing behavior.

## Commands

- `go run . translate --dry-run` — safe check (no API call, no clipboard write)
- `go run .` — full run (needs Wayland session + config, see SPEC.md)
- `go build -o cliptrans .` — binary the hotkey invokes
- `go vet ./...` — verify before commit
- Releases via GoReleaser; don't hand-roll install scripts.

## Clipboard — the one non-obvious thing

- Only via `wl-paste -n` / `wl-copy` subprocesses (`os/exec`). No Go clipboard libs (most assume X11), no `xclip`/`xsel`, no `-p`/`--primary`.
- `wl-paste` fails on empty/non-text clipboard — handle explicitly.
- `wl-copy` forking to background is normal.
- Must run in the user's Wayland session (`WAYLAND_DISPLAY`, `XDG_RUNTIME_DIR`). Broken over plain SSH is expected, not a bug.

## Conventions

- Cobra (`github.com/spf13/cobra`) for all commands/flags; `tint` (`github.com/lmittmann/tint`) as the `log/slog` handler. No other CLI/log libs.
- Config: YAML in `os.UserConfigDir()` (`~/.config/cliptrans/`); secret via `CLIPTRANS_API_KEY`. Never silently auto-create it.
- Fail fast, no retries — deliberate (§9 of SPEC.md). Never clobber the clipboard on failure.
- Errors to stderr, non-zero exit (hotkey shows nothing otherwise); user-visible failures also via best-effort `notify-send`.
- After changing behavior, update `SPEC.md` (and this file if workflow changed) before commit/PR. Code and docs must not drift.
