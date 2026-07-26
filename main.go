package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	_ "github.com/bizshuk/gosdk/log"
)

func main() {
	settings, err := loadSettings()
	if err != nil {
		slog.Error("load autop settings", "err", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command := newRootCommand(settings, defaultCommandDependencies())
	if err := command.ExecuteContext(ctx); err != nil {
		slog.Error("autop command failed", "err", err)
		os.Exit(commandExitCode(err))
	}
}

func commandExitCode(err error) int {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return 2
}
