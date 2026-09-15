# cliptrans

[![Go Version](https://img.shields.io/github/go-mod/go-version/acnosov/cliptrans)](https://go.dev/) [![License](https://img.shields.io/github/license/acnosov/cliptrans)](./LICENSE) [![Build](https://img.shields.io/github/actions/workflow/status/acnosov/cliptrans/validate.yml?branch=main)](https://github.com/acnosov/cliptrans/actions)

Hotkey-driven Wayland clipboard translator: reads clipboard text, translates it via an OpenAI-compatible LLM API, writes the translation back.

```sh
# Copy text anywhere, press the hotkey, then paste:
cliptrans translate
# Ctrl+V → translated text
```

Behavior specification for agents and contributors: [`SPEC.md`](./SPEC.md).

## Requirements

- Linux + Wayland session (`WAYLAND_DISPLAY`, `XDG_RUNTIME_DIR` set)
- `wl-clipboard` (`wl-paste`, `wl-copy`) installed
- `notify-send` optional — desktop notifications; falls back to stderr when absent
- An OpenAI-compatible API endpoint + key (OpenAI, OpenRouter, Ollama, vLLM)

Normative environment contract: see §3 and §5 of `SPEC.md`.

## Getting Started

### Install

Download a pre-built binary from [GitHub Releases](https://github.com/acnosov/cliptrans/releases/latest) (linux amd64/arm64 tarballs), or build from source:

```sh
go install github.com/acnosov/cliptrans@latest
# or from a clone:
go build -o cliptrans .
```

### First run

```sh
export CLIPTRANS_API_KEY="sk-..."
cliptrans translate --dry-run
```

The first run auto-creates `~/.config/cliptrans/config.yaml` with commented defaults — edit `base_url`, `model`, and `to` to taste. The secret lives in `CLIPTRANS_API_KEY`; `api_key` in the file is only a fallback. Config precedence and bootstrap rules: see §7 of `SPEC.md`.

Verify without side effects (no API call, no clipboard write):

```sh
cliptrans translate --dry-run
# target: en
# model: gpt-4o-mini
# chars: 42
```

### Hotkey

Bind a custom shortcut to `cliptrans translate`. Typical flow: select + Ctrl+C, press hotkey, wait for the "translated" notification, Ctrl+V. On KDE this preserves history via Klipper, so overwriting the clipboard is safe.

## Usage

```sh
cliptrans translate [--to uk] [--model gpt-4o-mini] [--dry-run]
cliptrans version
cliptrans --help
```

| Command / flag           | What it does                                                    |
| ------------------------ | --------------------------------------------------------------- |
| `translate`              | Translate current clipboard text, write result back             |
| `translate --to <lang>`  | Per-run target override (e.g. `--to uk`)                        |
| `translate --model <id>` | Per-run model override                                          |
| `translate --dry-run`    | Print resolved target, model, char count; no API call, no write |
| `version`                | Print version                                                   |
| `--config <path>`        | Config file path (default `~/.config/cliptrans/config.yaml`)    |
| `--verbose`              | Debug logging to stderr                                         |

Target resolution (`--to` → config `to` → built-in default) and translation details: see §6 of `SPEC.md`. Input cap and failure behavior (clipboard left untouched, notification + non-zero exit): see §§8–9 of `SPEC.md`.

## Configuration

```yaml
base_url: https://api.openai.com/v1
model: gpt-4o-mini
to: "en"
# Secret via CLIPTRANS_API_KEY env var; api_key here is only a fallback.
```

Full field list, defaults, and the editable `system_prompt` (`{{target}}` placeholder, `<text>…</text>` wrapper): see §§6–7 of `SPEC.md`. The auto-created file documents every field inline — it is the reference you edit.

## Contributing

```sh
just check   # mandatory before commit/PR: tidy-diff + oxfmt + lint + test-coverage + actionlint + goreleaser-check
```

Workflow and test constraints (never touch the real clipboard or `notify-send` in automated tests): see `AGENTS.md`.

## License

MIT — see [LICENSE](./LICENSE).
