package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/spf13/cobra"
)

// Version is overridden at build time via:
// go build -ldflags "-X github.com/acnosov/cliptrans/internal/cmd.Version=x.y.z".
var Version = "dev"

func NewRoot(ctx context.Context) *cobra.Command {
	// Flag state is kept in per-tree locals (not package globals) so concurrent
	// NewRoot trees — e.g. parallel tests — never share mutable state.
	var verbose bool
	var configFile string
	root := &cobra.Command{
		Use:   "cliptrans",
		Short: "Translate the Wayland clipboard via an LLM API",
		Long: `cliptrans translates the current clipboard text through an OpenAI-compatible
LLM API and writes the translation back to the clipboard.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}
			slog.SetDefault(slog.New(tint.NewTextHandler(os.Stderr, &tint.Options{Level: level})))
		},
	}

	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose logging to stderr")
	root.PersistentFlags().
		StringVar(&configFile, "config", "", "config file path (default: ~/.config/cliptrans/config.yaml)")
	root.AddCommand(newTranslateCmd(ctx, &configFile))
	root.AddCommand(newVersionCmd())

	return root
}

func ExecuteContext(ctx context.Context) error {
	if err := NewRoot(ctx).ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}
