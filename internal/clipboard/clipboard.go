package clipboard

import (
	"context"
	"fmt"
	"os/exec"
)

func Read(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "wl-paste", "-n").Output()
	if err != nil {
		return "", fmt.Errorf("wl-paste failed (clipboard empty or non-text?): %w", err)
	}
	return string(out), nil
}

func Write(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, "wl-copy", text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wl-copy failed: %w", err)
	}
	return nil
}
