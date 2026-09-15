# AGENTS.md

Go CLI that translates the Wayland clipboard via LLM API. No GUI.
Behavior source of truth: `SPEC.md` — read it before changing behavior.

## Documentation — one fact, one home

- `AGENTS.md` (this file): instructions for the agent — how to work,
  how to behave, what to run. Audience: agents.
- `SPEC.md`: specification of what the program does and how it behaves.
  Audience: agents implementing changes; source of truth for behavior.
- `README.md` (planned): information for users — how to install, run,
  and use the app. Not for agents, not a specification.
- No duplicated information across these files. Each fact lives in
  exactly one file; the others cross-reference it (`see §X of SPEC.md`).
- After changing behavior, update `SPEC.md` (and `README.md` once it
  exists, if users are affected) before commit/PR. After changing
  workflow, update this file. Code and docs must not drift.

## Commands

- `just check` — mandatory verification before commit/PR (runs `tidy-diff` + `oxfmt` + `lint` + `test-coverage` + `actionlint` + `goreleaser-check`; see `justfile`). `govulncheck` stays out of `check` (slow) and runs in prek hooks instead. `fmt` stays out of `check` (mutates files, races parallel readers) and runs separately.
- Narrower commands (e.g. `go test -run TestFoo ./internal/...`) only for granular diagnosis after `just check` fails.
- `go run . translate --dry-run` — safe manual check (no API call, no clipboard write)
- `go run .` — full manual run (needs Wayland session + config, see SPEC.md); never in automated checks.
- `go build -o cliptrans .` — binary the hotkey invokes
- Releases via GoReleaser; don't hand-roll install scripts.

## Clipboard — the one non-obvious thing

Contract and environment: §§3, 5 of SPEC.md. The pitfall beyond that:
`wl-paste` exits non-zero on empty/non-text clipboard — handle explicitly.

## Testing — no real side effects

- Automated tests (`go test`, hence `just check`) must NEVER touch the real clipboard (`wl-paste`/`wl-copy`) or real notifications (`notify-send`).
- Treat clipboard/notify like the external translate API: no real calls, use mocks/fakes/interfaces or local test doubles instead (the `translate` package pattern — `httptest` server instead of the real API — is the model).
- Real Wayland clipboard round-trip is manual-only (from a Wayland terminal session), never in automated tests.
- Integration tests hitting the real API live in `//go:build integration` files, excluded from plain `go test` and run via `go test -tags=integration ./...` (see `.env`: `CLIPTRANS_BASE_URL`, `CLIPTRANS_API_KEY`, optional `CLIPTRANS_MODEL`).

## Skills

Load via skill tool before work of that type:

- Always: `karpathy-guidelines`
- Go code: `golang-pro`, `golang-error-handling`, `golang-safety`, `golang-code-style`, `golang-naming`
- CLI: `golang-cli`, `golang-spf13-cobra`
- Tests: `golang-testing`, `golang-stretchr-testify`
- Debugging: `golang-troubleshooting`, `golang-gopls`
- Maintenance: `golang-lint`, `golang-refactoring`, `golang-modernize`, `golang-dependency-management`, `golang-security`
- Review before PR/push: `open-code-review`

## Conventions

- Cobra (`github.com/spf13/cobra`) for all commands/flags; `tint` (`github.com/lmittmann/tint`) as the `log/slog` handler. No other CLI/log libs.
- Failure handling is deliberate — see §9 of SPEC.md and follow it exactly.
- Never add `//nolint` (or any lint-suppression) comments to code. If a lint finding cannot be fixed cleanly, or suppressing it would hide a real problem, stop and ask the user what to do instead of adding a suppression.
- No comments in Go code by default. Write self-explanatory code and add a comment only in rare cases where a non-obvious/non-standard approach is used and the reader needs to know why it is done that way. Never narrate what the code does.
