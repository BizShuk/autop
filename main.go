package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	autopcmd "github.com/bizshuk/autop/cmd"
	_ "github.com/bizshuk/gosdk/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := autopcmd.Execute(ctx); err != nil {
		slog.Error("autop command failed", "err", err)
		os.Exit(autopcmd.CommandExitCode(err))
	}
}
