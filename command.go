package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	autopdriver "github.com/bizshuk/autop/driver"
	gosdkcmd "github.com/bizshuk/gosdk/cmd"
	"github.com/spf13/cobra"
)

type commandDependencies struct {
	getwd     func() (string, error)
	lookupEnv autopdriver.LookupEnv
	run       func(context.Context, autopdriver.Process, io.Writer, io.Writer, *slog.Logger) error
	logger    *slog.Logger
}

var (
	activeSettings     = defaultSettings()
	activeDependencies = defaultCommandDependencies()

	clientFlag           string
	templateFlag         string
	bypassPermissionFlag bool
	modelFlag            string
	effortFlag           string
)

func defaultCommandDependencies() commandDependencies {
	return commandDependencies{
		getwd:     os.Getwd,
		lookupEnv: os.LookupEnv,
		run:       runProcess,
		logger:    slog.Default(),
	}
}

func init() {
	flags := RootCmd.Flags()
	flags.StringVarP(&clientFlag, "client", "c", "", "configured LLM CLI client")
	flags.StringVarP(
		&templateFlag,
		"template",
		"t",
		"",
		"configured prompt template",
	)
	flags.BoolVar(
		&bypassPermissionFlag,
		"bypass-permission",
		false,
		"bypass the selected CLI permission checks",
	)
	flags.StringVar(&modelFlag, "model", "", "override the configured client model")
	flags.StringVar(&effortFlag, "effort", "", "override the configured client effort")

	RootCmd.AddCommand(WizardCmd, gosdkcmd.ConfigCmd)
}

func configureCommandRuntime(settings Settings, dependencies commandDependencies) {
	activeSettings = settings
	activeDependencies = dependencies
}

func runRootCommand(command *cobra.Command, args []string) error {
	workDir, err := activeDependencies.getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}
	clientID, client, err := resolveClient(activeSettings, clientFlag)
	if err != nil {
		return err
	}
	// Permission bypass is an explicit command-level opt-in. The profile's
	// AutoApprove value is used by the wizard as its default choice, but
	// must not implicitly enable dangerous provider flags for direct runs.
	bypassPermission := &bypassPermissionFlag
	client, err = applyClientOverrides(
		clientID,
		client,
		modelFlag,
		effortFlag,
		bypassPermission,
	)
	if err != nil {
		return err
	}
	rawPrompt, err := readPrompt(command.InOrStdin(), args)
	if err != nil {
		return err
	}
	prompt, err := renderPrompt(
		activeSettings,
		templateFlag,
		rawPrompt,
		workDir,
		clientID,
	)
	if err != nil {
		return err
	}
	process, err := autopdriver.Prepare(
		clientID,
		client,
		prompt,
		workDir,
		activeDependencies.lookupEnv,
	)
	if err != nil {
		return err
	}
	return activeDependencies.run(
		command.Context(),
		process,
		command.OutOrStdout(),
		command.ErrOrStderr(),
		activeDependencies.logger,
	)
}

func readPrompt(stdin io.Reader, args []string) (string, error) {
	positionalPrompt := strings.Join(args, " ")
	if !readerHasPipedInput(stdin) {
		return positionalPrompt, nil
	}
	content, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("read prompt from stdin: %w", err)
	}
	stdinPrompt := string(content)
	if positionalPrompt != "" && strings.TrimSpace(stdinPrompt) != "" {
		return "", fmt.Errorf("prompt must come from positional arguments or stdin, not both")
	}
	if positionalPrompt != "" {
		return positionalPrompt, nil
	}
	return stdinPrompt, nil
}

func readerHasPipedInput(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return true
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}
