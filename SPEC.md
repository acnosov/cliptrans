# cliptrans — Specification

Source of truth for what the app does. `AGENTS.md` holds only agent workflow instructions.

## 1. Goal

Go CLI, no GUI. Invoked via a hotkey, it translates the current clipboard text
through an LLM API and writes the translation back to the clipboard.

Core loop:

```
hotkey → cliptrans translate → wl-paste -n → detect direction →
HTTP translate → wl-copy → user pastes manually (Ctrl+V)
```

## 2. Non-goals (deliberate)

- No auto-paste into the focused window (no `wtype`/`ydotool`/fake keys).
- No daemon / clipboard-watch mode; single-shot invocations only.
- No GUI/TUI.
- No multi-provider driver abstraction; one OpenAI-compatible HTTP client.
- No retries on failed translates (fail fast, see §9).

## 3. Runtime environment

- Linux Wayland only. Developed/tested on KDE with the stock Klipper
  clipboard manager.
- Overwriting the clipboard is acceptable — Klipper keeps history.
- Must run inside the user's Wayland session (`WAYLAND_DISPLAY`,
  `XDG_RUNTIME_DIR` set). Breaking over plain SSH is expected, not a bug.

## 4. CLI (Cobra, `github.com/spf13/cobra`)

- Hotkey invokes the explicit subcommand: `cliptrans translate`.
- Bare `cliptrans` with no args must NOT translate (prints help).
- Subcommands:
  - `translate` — the hotkey path. Flags: `--to` (override direction),
    `--model` (override config model), `--verbose`, `--dry-run`.
  - `config init` — writes a commented default YAML config.
  - `config path` — prints the config file path.
  - `version` — prints version.
- All commands/flags go through Cobra.

## 5. Clipboard access

- Only via `wl-paste` / `wl-copy` subprocesses (`os/exec`). No Go clipboard
  libraries (most assume X11), no `xclip`/`xsel`.
- Read: `wl-paste -n` (text, no trailing-newline munging).
- Write: `wl-copy <text>` (forks to background by default; that is normal).
- Regular (Ctrl+C) selection only — no `-p`/`--primary` unless requested.

## 6. Translation

- Direction: auto-detect RU↔EN by script heuristic — any Cyrillic letter
  present → translate to English, otherwise to Russian. `--to` overrides.
- Provider: generic OpenAI-compatible HTTP endpoint
  (`base_url` + `model` + API key). Covers OpenAI, OpenRouter, Ollama, vLLM
  with one code path.
- Sampling: temperature 0 (deterministic translator).
- System prompt (draft): translator that outputs only the translated text —
  no quotes, no commentary, no explanations. Multiline structure preserved.

## 7. Config

- Format: YAML. Location: standard user config dir (`os.UserConfigDir()`),
  i.e. `~/.config/cliptrans/config.yaml`.
- Fields: `base_url`, `model`, `temperature` (default 0),
  `max_chars` (default 8000). Optional `api_key` as fallback.
- Secret primary: `CLIPTRANS_API_KEY` environment variable.
- Bootstrap: no config on first run → `translate` fails with an error that
  prints the config path and points at `config init`. Never silently
  auto-create defaults.

## 8. Input limits

- Hard cap `max_chars` (default 8000, tunable in config).
- Over-cap input: refuse with a desktop notification (see §9), leave the
  clipboard untouched. No silent truncation, no chunk-and-join.

## 9. Failure handling

- Fail fast: exactly one API attempt, no retries, 30s request timeout.
- On ANY failure (empty/non-text clipboard, over-cap, missing config,
  API/network error): leave the clipboard untouched.
- Notify: best-effort `notify-send` summary; if `notify-send` is absent,
  log to stderr. `notify-send` is not a hard dependency.
- Exit non-zero on failure (exact exit-code scheme TBD at implementation).

## 10. Logging

- `tint` (`github.com/lmittmann/tint`) as the `log/slog` handler. No other
  loggers. Destination: stderr only. `--verbose` raises the level.
- Hotkey runs have no terminal: user-visible failures go through the
  notification from §9, never through logs alone.

## 11. Verification

- `translate --dry-run`: prints detected direction, effective model, and
  char count; calls no API, touches no clipboard.
- Clipboard round-trip (`wl-paste`/`wl-copy`) must be tested from a real
  Wayland terminal session.

## 12. Distribution

- Releases via GoReleaser: downloadable binaries from the repo.
- Installable via `go install`. If the repo goes open-source, `mise` and
  similar tools become options. No hand-rolled install scripts.
