# cliptrans — Specification

Source of truth for what the app does. Documentation roles (agent
instructions vs this specification vs user docs): see `AGENTS.md`.

## 1. Goal

Go CLI, no GUI. Invoked via a hotkey, it translates the current clipboard text
through an LLM API and writes the translation back to the clipboard.

Core loop:

```
hotkey → cliptrans translate → wl-paste -n →
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

## 4. CLI

- Hotkey invokes the explicit subcommand: `cliptrans translate`.
- Bare `cliptrans` with no args must NOT translate (prints help).
- Persistent flags (all commands): `--config` (config file path,
  default `~/.config/cliptrans/config.yaml`), `--verbose`.
- Subcommands:
  - `translate` — the hotkey path. Flags: `--to` (override target),
    `--model` (override config model), `--dry-run` (prints resolved
    target, effective model, and char count; no API call, no clipboard
    write).
  - `version` — prints version.

## 5. Clipboard access

- Only via `wl-paste` / `wl-copy` subprocesses (`os/exec`). No Go clipboard
  libraries (most assume X11), no `xclip`/`xsel`.
- Read: `wl-paste -n` (text, no trailing-newline munging).
- Write: `wl-copy <text>` (forks to background by default; that is normal).
- Regular (Ctrl+C) selection only — no `-p`/`--primary` unless requested.

## 6. Translation

- Target: a single target language, any value the model understands
  (`"en"`, `"ru"`, `"uk"`, `"de"`, ...). Precedence: `--to` flag, then
  `to` config value, then the built-in default `"en"`. The source
  language is never identified or sent anywhere: the model translates
  whatever text it receives into the target language.
  `cliptrans translate` with no flags translates into the configured
  default (e.g. `to: "en"`); `cliptrans translate --to uk` translates
  the same clipboard (e.g. Russian text) into Ukrainian.
- Provider: generic OpenAI-compatible HTTP endpoint
  (`base_url` + `model` + API key). Covers OpenAI, OpenRouter, Ollama, vLLM
  with one code path.
- Sampling: temperature 0 (deterministic translator).
- System prompt: one universal prompt, `system_prompt`, stored in the
  config file so it is fully visible and editable. `{{target}}` in the
  prompt is replaced with the resolved target language label before
  sending: a known code expands to `Name (code)` (e.g. `uk` →
  `Ukrainian (uk)`), an unknown value is substituted verbatim.
  The clipboard text is sent as the user message wrapped in
  `<text>...</text>`, which is the wrapper the prompt refers to.
  The built-in default asks for a faithful translation without adding
  anything of its own. Configs predating this field keep working:
  the built-in default fills in for a missing value.

## 7. Config

- Format: YAML. Location: standard user config dir (`os.UserConfigDir()`),
  i.e. `~/.config/cliptrans/config.yaml`.
- Fields: `base_url`, `model`, `temperature` (default 0),
  `max_chars` (default 8000), `to` (default `"en"`),
  `notify_on_success` (default true), `notify_on_error` (default true),
  `system_prompt` (default built-in, see §6).
  Optional `api_key` as fallback.
- Secret primary: `CLIPTRANS_API_KEY` environment variable.
- Bootstrap: no config at the default path and no `--config` override →
  `translate` auto-creates a commented default config there and logs its
  path (edit it to taste). An explicit `--config <path>` never
  auto-creates: a missing file there is a hard error.

## 8. Input limits

- Hard cap `max_chars` (default 8000, tunable in config).
  Zero or negative values are rejected as invalid config.
- Over-cap input: refuse with a desktop notification (see §9), leave the
  clipboard untouched. No silent truncation, no chunk-and-join.

## 9. Failure handling

- Fail fast: exactly one API attempt, no retries, 30s request timeout.
- On ANY failure (empty/non-text clipboard, over-cap, missing config,
  API/network error): leave the clipboard untouched.
- Notify: best-effort `notify-send`; if `notify-send` is absent,
  log to stderr. `notify-send` is not a hard dependency.
  - Failures notify only when `notify_on_error` is true (default true).
  - Successes notify only when `notify_on_success` is true
    (default true). The success notification body carries a preview
    (first 500 chars) of the translation; the full translation is
    always logged to stderr.
  - Logging to stderr happens in both cases regardless of the toggles.
- Exit non-zero on failure (exact exit-code scheme TBD at implementation).

## 10. Logging

- Destination: stderr only. `--verbose` raises the level.
- Hotkey runs have no terminal: user-visible failures go through the
  notification from §9, never through logs alone.

## 11. Distribution

- Downloadable binaries from the repo.
- Installable via `go install`. If the repo goes open-source, `mise` and
  similar tools become options.
