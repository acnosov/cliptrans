package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/acnosov/cliptrans/internal/clipboard"
	"github.com/acnosov/cliptrans/internal/config"
	"github.com/acnosov/cliptrans/internal/notify"
	"github.com/acnosov/cliptrans/internal/translate"

	"github.com/spf13/cobra"
)

const successPreviewLimit = 500

type translateOpts struct {
	to     string
	model  string
	dryRun bool
}

func newTranslateCmd(ctx context.Context, configFile *string) *cobra.Command {
	var opts translateOpts
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate clipboard text and write it back",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTranslate(ctx, cmd, opts, *configFile)
		},
	}
	cmd.Flags().
		StringVar(&opts.to, "to", "", `translation target, e.g. "en", "ru", "uk" (default: from config)`)
	cmd.Flags().StringVar(&opts.model, "model", "", "model override (default: from config)")
	cmd.Flags().
		BoolVar(&opts.dryRun, "dry-run", false, "print target, model and char count; no API call, no clipboard write")
	return cmd
}

func loadEffectiveConfig(configFile string) (*config.Config, error) {
	if configFile != "" {
		cfg, err := config.Load(configFile)
		if err != nil {
			return nil, fmt.Errorf("load config %s: %w", configFile, err)
		}
		return cfg, nil
	}
	path, created, err := config.EnsureDefault()
	if err != nil {
		return nil, fmt.Errorf("ensure default config: %w", err)
	}
	if created {
		slog.Info("created default config", slog.String("path", path))
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	return cfg, nil
}

func runTranslate(ctx context.Context, cmd *cobra.Command, opts translateOpts, configFile string) error {
	cfg, err := loadEffectiveConfig(configFile)
	if err != nil {
		notify.Failure(ctx, "cliptrans: no config", err.Error())
		return err
	}

	plan, summary, err := buildPlan(ctx, opts, cfg)
	if err != nil {
		return reportFailure(ctx, cfg, summary, err)
	}

	if opts.dryRun {
		printDryRun(cmd, plan.target, plan.model, plan.chars)
		return nil
	}

	if keyErr := requireAPIKey(plan.cfg); keyErr != nil {
		return reportFailure(ctx, cfg, "cliptrans: missing API key", keyErr)
	}

	out, err := finishTranslate(ctx, plan.cfg, plan.model, plan.clip, plan.target, plan.chars)
	if err != nil {
		return reportFailure(ctx, cfg, "cliptrans: translate failed", err)
	}
	if plan.cfg.NotifyOnSuccessEnabled() {
		notify.Success(ctx, "cliptrans: translated to "+plan.target, preview(out))
	}
	return nil
}

func reportFailure(ctx context.Context, cfg *config.Config, summary string, err error) error {
	detail := detailOf(err, summary)
	if cfg.NotifyOnErrorEnabled() {
		notify.Failure(ctx, summary, detail)
	} else {
		slog.ErrorContext(ctx, summary, slog.String("detail", detail))
	}
	if err != nil {
		return err
	}
	return errors.New(summary)
}

func finishTranslate(
	ctx context.Context,
	cfg *config.Config,
	model, clip, target string,
	chars int,
) (string, error) {
	slog.DebugContext(ctx, "translating", slog.String("target", target), slog.String("model", model))
	client := &translate.Client{
		BaseURL:      cfg.BaseURL,
		Model:        model,
		Temperature:  cfg.Temperature,
		APIKey:       cfg.APIKey(),
		SystemPrompt: promptForTarget(cfg, target),
	}
	out, err := client.Translate(ctx, clip)
	if err != nil {
		return "", fmt.Errorf("translate via API: %w", err)
	}
	if err := clipboard.Write(ctx, out); err != nil {
		return "", fmt.Errorf("write clipboard: %w", err)
	}
	slog.InfoContext(ctx, "translated",
		slog.String("target", target),
		slog.Int("chars", chars),
		slog.String("translation", out))
	return out, nil
}

func preview(s string) string {
	runes := []rune(s)
	if len(runes) <= successPreviewLimit {
		return s
	}
	return string(runes[:successPreviewLimit]) + "…"
}

func detailOf(err error, summary string) string {
	if err != nil {
		return err.Error()
	}
	return summary
}

func readClip(ctx context.Context) (string, int, error) {
	clip, err := clipboard.Read(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("read clipboard: %w", err)
	}
	chars := len([]rune(clip))
	if chars == 0 {
		return "", 0, errors.New("clipboard is empty")
	}
	return clip, chars, nil
}

func checkLimit(chars, limit int) error {
	if chars > limit {
		return fmt.Errorf(
			"clipboard has %d chars, limit is %d — clipboard left untouched",
			chars,
			limit,
		)
	}
	return nil
}

func printDryRun(cmd *cobra.Command, target, model string, chars int) {
	fmt.Fprintf(cmd.OutOrStdout(), "target: %s\nmodel: %s\nchars: %d\n",
		target, model, chars)
}

func promptForTarget(cfg *config.Config, target string) string {
	return cfg.RenderPrompt(target)
}
