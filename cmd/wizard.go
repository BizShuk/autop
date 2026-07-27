package autop

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const noTemplateChoice = "(none)"

// WizardCmd interactively configures an autop PM2 task.
var WizardCmd = &cobra.Command{
	Use:     "wizard",
	Aliases: []string{"w"},
	Short:   "Interactively configure an autop PM2 task",
	Args:    cobra.NoArgs,
	RunE:    runWizardCommand,
}

func runWizardCommand(command *cobra.Command, _ []string) error {
	workDir, err := activeDependencies.getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}
	reader := bufio.NewReader(command.InOrStdin())
	writer := command.OutOrStdout()

	clientID, err := promptChoice(
		reader,
		writer,
		"Choose CLI",
		enabledClientIDs(activeSettings),
		activeSettings.DefaultClient,
	)
	if err != nil {
		return err
	}
	_, client, err := resolveClient(activeSettings, clientID)
	if err != nil {
		return err
	}

	templateID, err := promptChoice(
		reader,
		writer,
		"Choose template",
		templateChoices(activeSettings),
		noTemplateChoice,
	)
	if err != nil {
		return err
	}
	if templateID == noTemplateChoice {
		templateID = ""
	}

	bypassDefault := "no"
	if client.AutoApprove {
		bypassDefault = "yes"
	}
	bypassChoice, err := promptChoice(
		reader,
		writer,
		"Bypass permission",
		[]string{"yes", "no"},
		bypassDefault,
	)
	if err != nil {
		return err
	}
	bypassPermission := bypassChoice == "yes"

	model, err := promptChoice(
		reader,
		writer,
		"Choose model",
		clientModelChoices(client),
		client.Model,
	)
	if err != nil {
		return err
	}
	effort, err := promptChoice(
		reader,
		writer,
		"Choose effort",
		clientEffortChoices(client),
		client.Effort,
	)
	if err != nil {
		return err
	}
	if _, err := applyClientOverrides(
		clientID,
		client,
		model,
		effort,
		&bypassPermission,
	); err != nil {
		return err
	}

	prompt, err := promptTaskInput(reader, writer, templateID)
	if err != nil {
		return err
	}

	cronSchedule, err := promptCronSchedule(reader, writer)
	if err != nil {
		return err
	}

	path, workspaceDir, err := resolveWizardTarget(workDir)
	if err != nil {
		return err
	}
	task := ecosystemTask{
		ClientID:              clientID,
		TemplateID:            templateID,
		Prompt:                prompt,
		WorkspaceDir:          workspaceDir,
		BypassPermission:      bypassPermission,
		Model:                 model,
		Effort:                effort,
		Cron:                  cronSchedule,
		IncludeRuntimeOptions: true,
	}
	if err := installEcosystemTask(path, task); err != nil {
		return err
	}
	activeDependencies.logger.Info(
		"configured autop PM2 task",
		"path",
		path,
		"workspace",
		workspaceDir,
		"client",
		clientID,
		"template",
		templateID,
		"bypass_permission",
		bypassPermission,
		"model",
		model,
		"effort",
		effort,
		"cron",
		cronSchedule,
	)
	return nil
}

func resolveWizardTarget(workDir string) (configPath string, workspaceDir string, err error) {
	workspaceDir, err = filepath.Abs(workDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve wizard workspace %q: %w", workDir, err)
	}

	for current := workspaceDir; ; current = filepath.Dir(current) {
		packageDir := filepath.Join(current, "cmd")
		info, statErr := os.Stat(packageDir)
		if statErr == nil {
			if info.IsDir() {
				return filepath.Join(packageDir, "ecosystem.config.js"), current, nil
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return "", "", fmt.Errorf("inspect autop package directory %q: %w", packageDir, statErr)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}

	return filepath.Join(workspaceDir, "ecosystem.config.js"), workspaceDir, nil
}

func promptCronSchedule(
	reader *bufio.Reader,
	writer io.Writer,
) (string, error) {
	const prompt = "Cron schedule [N=none, r=random 02:00-08:00, or 5-field cron]: "
	for {
		if _, err := fmt.Fprint(writer, prompt); err != nil {
			return "", fmt.Errorf("write cron schedule prompt: %w", err)
		}
		input, readErr := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		switch strings.ToLower(input) {
		case "", "n":
			return "", nil
		case "r":
			return randomDailyCronSchedule(), nil
		}
		if isFiveFieldCronSchedule(input) {
			return input, nil
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return "", fmt.Errorf("cron schedule must be a valid 5-field cron")
			}
			return "", fmt.Errorf("read cron schedule: %w", readErr)
		}
		if _, err := fmt.Fprintln(
			writer,
			"Cron schedule must be a valid 5-field cron. Try again.",
		); err != nil {
			return "", fmt.Errorf("write cron schedule validation error: %w", err)
		}
	}
}

func isFiveFieldCronSchedule(input string) bool {
	fields := strings.Fields(input)
	if len(fields) != 5 {
		return false
	}
	for _, field := range fields {
		if strings.Trim(field, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789*/,-") != "" {
			return false
		}
	}
	return true
}

func promptTaskInput(
	reader *bufio.Reader,
	writer io.Writer,
	templateID string,
) (string, error) {
	const promptLabel = "Prompt (optional for templates, required without one): "
	if _, err := fmt.Fprint(writer, promptLabel); err != nil {
		return "", fmt.Errorf("write prompt input prompt: %w", err)
	}
	input, readErr := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" && templateID == "" {
		return "", fmt.Errorf("prompt is required when no template is selected")
	}
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", fmt.Errorf("read prompt: %w", readErr)
	}
	return input, nil
}

func randomDailyCronSchedule() string {
	const (
		firstMinute = 2 * 60
		lastMinute  = 8 * 60
	)
	minuteOfDay := firstMinute + rand.IntN(lastMinute-firstMinute+1)
	return fmt.Sprintf("%d %d * * *", minuteOfDay%60, minuteOfDay/60)
}

func promptChoice(
	reader *bufio.Reader,
	writer io.Writer,
	label string,
	options []string,
	defaultValue string,
) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("%s has no available options", label)
	}
	for {
		if _, err := fmt.Fprintf(writer, "%s:\n", label); err != nil {
			return "", fmt.Errorf("write wizard prompt: %w", err)
		}
		for index, option := range options {
			if _, err := fmt.Fprintf(writer, "  %d) %s\n", index+1, option); err != nil {
				return "", fmt.Errorf("write wizard option: %w", err)
			}
		}
		if _, err := fmt.Fprintf(writer, "Select [%s]: ", defaultValue); err != nil {
			return "", fmt.Errorf("write wizard selection prompt: %w", err)
		}

		input, readErr := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}
		if index, err := strconv.Atoi(input); err == nil {
			if index >= 1 && index <= len(options) {
				return options[index-1], nil
			}
		}
		if contains(options, input) {
			return input, nil
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return "", fmt.Errorf("invalid %s selection %q", label, input)
			}
			return "", fmt.Errorf("read %s selection: %w", label, readErr)
		}
		if _, err := fmt.Fprintf(writer, "Invalid selection %q. Try again.\n", input); err != nil {
			return "", fmt.Errorf("write wizard validation error: %w", err)
		}
	}
}

func enabledClientIDs(settings Settings) []string {
	clientIDs := make([]string, 0, len(settings.Clients))
	for clientID, client := range settings.Clients {
		if !client.Disabled {
			clientIDs = append(clientIDs, clientID)
		}
	}
	sort.Strings(clientIDs)
	return clientIDs
}

func templateChoices(settings Settings) []string {
	templates := make([]string, 0, len(settings.Templates)+1)
	templates = append(templates, noTemplateChoice)
	templateIDs := make([]string, 0, len(settings.Templates))
	for templateID := range settings.Templates {
		templateIDs = append(templateIDs, templateID)
	}
	sort.Strings(templateIDs)
	return append(templates, templateIDs...)
}
