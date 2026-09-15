package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/acnosov/cliptrans/internal/config"
)

type translatePlan struct {
	cfg    *config.Config
	clip   string
	target string
	model  string
	chars  int
}

func buildPlan(
	ctx context.Context,
	opts translateOpts,
	cfg *config.Config,
) (*translatePlan, string, error) {
	clip, chars, err := readClip(ctx)
	if err != nil {
		return nil, "cliptrans: empty clipboard", err
	}

	target := cfg.To
	if strings.TrimSpace(opts.to) != "" {
		target, err = config.NormalizeTarget(opts.to)
		if err != nil {
			return nil, "cliptrans: bad target", fmt.Errorf("resolve target: %w", err)
		}
	}
	plan := &translatePlan{
		clip:   clip,
		chars:  chars,
		target: target,
		cfg:    cfg,
		model:  effectiveModel(cfg, opts.model),
	}

	if err := checkLimit(plan.chars, plan.cfg.MaxChars); err != nil {
		return nil, "cliptrans: clipboard over limit", err
	}
	return plan, "", nil
}

func effectiveModel(cfg *config.Config, override string) string {
	if override != "" {
		return override
	}
	return cfg.Model
}

func requireAPIKey(cfg *config.Config) error {
	if cfg.APIKey() != "" {
		return nil
	}
	return fmt.Errorf("set %s (or `api_key` in config)", config.KeyEnvName)
}
