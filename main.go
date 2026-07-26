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
	"github.com/spf13/cobra"
)

// RootCmd is the top-level autop command.
var RootCmd = &cobra.Command{
	Use:           "autop [prompt]",
	Short:         "Run a configured local LLM CLI through one facade",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.ArbitraryArgs,
	RunE:          runRootCommand,
}

func main() {
	settings, err := loadSettings()
	if err != nil {
		slog.Error("load autop settings", "err", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	configureCommandRuntime(settings, defaultCommandDependencies())
	if err := RootCmd.ExecuteContext(ctx); err != nil {
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
