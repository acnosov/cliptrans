package notify

import (
	"context"
	"log/slog"
	"os/exec"
	"time"
)

const notifyTimeout = 5 * time.Second

func Success(ctx context.Context, summary, body string) {
	send(ctx, "--urgency=normal", summary, body)
	slog.InfoContext(ctx, summary, slog.String("detail", body))
}

func Failure(ctx context.Context, summary, body string) {
	send(ctx, "--urgency=critical", summary, body)
	slog.ErrorContext(ctx, summary, slog.String("detail", body))
}

func send(ctx context.Context, urgency, summary, body string) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return
	}
	args := []string{"--app-name=cliptrans", urgency, summary}
	if body != "" {
		args = append(args, body)
	}
	ctx, cancel := context.WithTimeout(ctx, notifyTimeout)
	defer cancel()
	if err := exec.CommandContext(ctx, "notify-send", args...).Run(); err != nil {
		slog.ErrorContext(ctx, "notify-send failed", slog.Any("error", err))
	}
}
