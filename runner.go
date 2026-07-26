package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	autopdriver "github.com/bizshuk/cc-plugin/cmd/autop/driver"
)

func runProcess(
	ctx context.Context,
	process autopdriver.Process,
	stdout io.Writer,
	stderr io.Writer,
	logger *slog.Logger,
) error {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("executing command", "command", formatCommand(process))

	command := exec.CommandContext(ctx, process.Name, process.Args...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.Env = append(os.Environ(), process.ExtraEnv...)
	if process.Stdin != "" {
		command.Stdin = strings.NewReader(process.Stdin)
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("run %s: %w", process.Name, err)
	}
	return nil
}

func formatCommand(process autopdriver.Process) string {
	parts := make([]string, 0, len(process.Args)+1)
	parts = append(parts, quoteCommandPart(process.Name))
	for _, arg := range process.Args {
		parts = append(parts, quoteCommandPart(arg))
	}
	return strings.Join(parts, " ")
}

func quoteCommandPart(value string) string {
	if value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			strings.ContainsRune("_@%+=:,./-", r))
	}) == -1 {
		return value
	}
	return strconv.Quote(value)
}
