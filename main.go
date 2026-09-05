// Hello-world milestone: verifies clipboard round-trip over Wayland.
// Reads text via `wl-paste -n`, prints it, writes a marker back via `wl-copy`.
// Run with something in the clipboard: `go run .`
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cliptrans:", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Read current clipboard (fails if empty or non-text).
	raw, err := exec.Command("wl-paste", "-n").Output()
	if err != nil {
		return fmt.Errorf("wl-paste failed (clipboard empty or non-text?): %w", err)
	}
	fmt.Printf("read %d bytes from clipboard:\n%s\n", len(raw), raw)

	// 2. Write hello-world marker back. wl-copy forks to background by default.
	msg := "cliptrans hello world"
	if err := exec.Command("wl-copy", msg).Run(); err != nil {
		return fmt.Errorf("wl-copy failed: %w", err)
	}
	fmt.Printf("wrote %q back to clipboard (check Klipper history, original is preserved there)\n", msg)
	return nil
}
