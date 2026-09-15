package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/acnosov/cliptrans/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := cmd.ExecuteContext(ctx)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cliptrans:", err)
		os.Exit(1)
	}
}
