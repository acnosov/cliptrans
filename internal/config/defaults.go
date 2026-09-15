package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultPrompt = `Your role is an expert professional translator. Your goal is to produce a faithful translation without adding, omitting, or altering meaning.

Task: Translate the following text into {{target}}.

Definitions:
- Source text: the content inside <text>…</text>, in whatever language it is written.

Constraints:
- Fidelity first: Preserve all facts, nuances, and implications. Do not summarize, explain, or paraphrase creatively.
- Tone & register: Preserve the original tone, formality, and voice.
- Formatting: Preserve paragraph breaks, list structure, punctuation, and emphasis.
- Non‑translatables: Leave code, math, URLs, hashtags, usernames, and inline tags unchanged.
- Entities, numbers, and units: Keep names, numbers, and units as written (no conversions; do not localize digits).
- Safety/meta: Do not add warnings, disclaimers, or commentary.

Forbidden:
- Do not add titles, prefaces, translator notes, bracketed explanations, or safety disclaimers.
- Do not convert units or rewrite numerals.

Output (strict): Return only the translation text, preserving paragraphing. No title/preface/explanations.`

func DefaultYAML() string {
	return `# cliptrans config — ~/.config/cliptrans/config.yaml
# Secret lives in the CLIPTRANS_API_KEY env var; api_key below is only a fallback.

# OpenAI-compatible endpoint prefix, e.g.:
#   OpenAI:     https://api.openai.com/v1
#   OpenRouter: https://openrouter.ai/api/v1
#   Ollama:     http://localhost:11434/v1
base_url: https://api.openai.com/v1

# Default model; override per-run with --model.
model: gpt-4o-mini

# Sampling temperature. 0 = deterministic translator.
temperature: 0

# Hard cap on input chars; over-cap input is refused, never truncated.
max_chars: 8000

# Optional fallback if CLIPTRANS_API_KEY is unset.
# api_key: sk-...

# Translation target: any language the model understands, e.g. en, ru, uk.
# The --to flag overrides this value per run.
to: "en"

# Desktop notifications via notify-send. Every run still logs to stderr;
# these toggles only control the popup.
notify_on_success: true
notify_on_error: true

# System prompt. {{target}} is replaced with the resolved target language.
# The clipboard text is sent as the user message wrapped in
# <text>...</text>, which is the wrapper the prompt below refers to.
# Edit freely — this file is the full source of truth for what is sent.
system_prompt: |
` + indentBlock(DefaultPrompt)
}

func indentBlock(s string) string {
	var out strings.Builder
	for line := range strings.SplitSeq(strings.TrimRight(s, "\n"), "\n") {
		if line != "" {
			out.WriteString("  ")
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(dir, "cliptrans", "config.yaml"), nil
}

func EnsureDefault() (string, bool, error) {
	path, err := DefaultPath()
	if err != nil {
		return "", false, err
	}
	if err = os.MkdirAll(filepath.Dir(path), configDirPerm); err != nil {
		return path, false, fmt.Errorf("create config dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, configFilePerm)
	if err != nil {
		if os.IsExist(err) {
			return path, false, nil
		}
		return path, false, fmt.Errorf("write config %s: %w", path, err)
	}
	if _, err := f.WriteString(DefaultYAML()); err != nil {
		return path, false, fmt.Errorf("write config %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return path, false, fmt.Errorf("write config %s: %w", path, err)
	}
	return path, true, nil
}
