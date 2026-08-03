package autop

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	autopdriver "github.com/bizshuk/autop/cmd/driver"
	gosdkcmd "github.com/bizshuk/gosdk/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func rootCommandForTest(
	t *testing.T,
	settings Settings,
	dependencies commandDependencies,
) *cobra.Command {
	t.Helper()

	resetRootCommandForTest()
	configureCommandRuntime(settings, dependencies)
	t.Cleanup(func() {
		resetRootCommandForTest()
		configureCommandRuntime(defaultSettings(), defaultCommandDependencies())
	})
	return RootCmd
}

func resetRootCommandForTest() {
	clientFlag = ""
	templateFlag = ""
	bypassPermissionFlag = false
	modelFlag = ""
	effortFlag = ""
	dryRunFlag = false
	RootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
	})
	RootCmd.SetArgs(nil)
	RootCmd.SetIn(nil)
	RootCmd.SetOut(nil)
	RootCmd.SetErr(nil)
}

func TestCommandPackageExposesExecute(t *testing.T) {
	var execute func(context.Context) error = Execute
	if execute == nil {
		t.Fatal("Execute must be available to the root entrypoint")
	}
}

func TestRootCommandUsesDefaultClientAndTemplate(t *testing.T) {
	settings := testSettings()
	var captured autopdriver.Process
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			process autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			captured = process
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"-t", "system"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if captured.Name != "codex" {
		t.Fatalf("process.Name = %q, want default codex", captured.Name)
	}
	if captured.Stdin != "run $system-planner for current workspace" {
		t.Fatalf("process.Stdin = %q", captured.Stdin)
	}
}

func TestRootCommandRegistersGosdkConfigCommand(t *testing.T) {
	command := rootCommandForTest(t, testSettings(), defaultCommandDependencies())

	configCommand, _, err := command.Find([]string{"config"})
	if err != nil {
		t.Fatalf("Find(config) error = %v", err)
	}
	if configCommand != gosdkcmd.ConfigCmd {
		t.Fatalf("config command = %#v, want gosdk config command", configCommand)
	}
	if _, _, err := configCommand.Find([]string{"default"}); err != nil {
		t.Fatalf("config default subcommand is not registered: %v", err)
	}
}

func TestRootCommandDoesNotBypassWhenFlagIsOmitted(t *testing.T) {
	settings := testSettings()
	var captured autopdriver.Process
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			process autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			captured = process
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"-c", "codex", "-t", "system"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.Contains(strings.Join(captured.Args, " "), "--dangerously-bypass-approvals-and-sandbox") {
		t.Fatalf("process bypasses permissions without explicit flag: %#v", captured.Args)
	}
}

func TestRootCommandEnablesBypassWhenFlagIsExplicit(t *testing.T) {
	settings := testSettings()
	var captured autopdriver.Process
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			process autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			captured = process
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{
		"-c",
		"codex",
		"-t",
		"system",
		"--bypass-permission=true",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(strings.Join(captured.Args, " "), "--dangerously-bypass-approvals-and-sandbox") {
		t.Fatalf("process does not bypass permissions with explicit flag: %#v", captured.Args)
	}
}

func TestWizardCommandWritesSelectedFlags(t *testing.T) {
	workDir := t.TempDir()
	settings := testSettings()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	var wizardOutput bytes.Buffer
	command.SetIn(strings.NewReader("agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\n\n"))
	command.SetOut(&wizardOutput)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := readTestFile(t, filepath.Join(workDir, "ecosystem.config.js"))
	if !strings.Contains(
		got,
		`args: ["-c", "agy", "-t", "system", "--bypass-permission=false", `+
			`"--model", "gemini-3.6-flash-high", "--effort", "high", "--", `+
			`"'review current workspace'"]`,
	) {
		t.Fatalf("ecosystem config has wrong args:\n%s", got)
	}
	if strings.Contains(got, "cron:") {
		t.Fatalf("ecosystem config contains cron when wizard selected no:\n%s", got)
	}
	if !strings.Contains(got, "optional: true") {
		t.Fatalf("ecosystem config does not mark the task optional:\n%s", got)
	}
	lastPromptIndex := -1
	for _, label := range []string{
		"Choose CLI:",
		"Choose template:",
		"Bypass permission:",
		"Choose model:",
		"Choose effort:",
		"Prompt (optional for templates, required without one): ",
		"Cron schedule [N=none, r=random 02:00-08:00, or 5-field cron]:",
	} {
		index := strings.Index(wizardOutput.String(), label)
		if index <= lastPromptIndex {
			t.Fatalf("wizard prompt %q is missing or out of order:\n%s", label, wizardOutput.String())
		}
		lastPromptIndex = index
	}
}

func TestWizardCommandPrintsCommandAndEcosystemSummary(t *testing.T) {
	workDir := t.TempDir()
	settings := testSettings()
	codex := settings.Clients["codex"]
	codex.Models = []string{"gpt-5.6-sol", "gpt-5.6-luna"}
	codex.Efforts = []string{"xhigh", "medium"}
	settings.Clients["codex"] = codex
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	var wizardOutput bytes.Buffer
	command.SetIn(strings.NewReader(
		"codex\n(none)\nyes\ngpt-5.6-luna\nmedium\n$find-activity find 5 event s in EU\n\n",
	))
	command.SetOut(&wizardOutput)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	configPath := filepath.Join(workDir, "ecosystem.config.js")
	configContent := readTestFile(t, configPath)
	wantOriginal := "\x1b[36mOriginal command (autop):\x1b[0m\n" +
		"  autop -c codex --bypass-permission=true --model gpt-5.6-luna" +
		" --effort medium -- '$find-activity find 5 event s in EU'\n"
	wantExecute := "\x1b[32mExecute command (codex):\x1b[0m\n" +
		"  printf '%s' '$find-activity find 5 event s in EU'" +
		" | codex exec --dangerously-bypass-approvals-and-sandbox" +
		" --model gpt-5.6-luna -c 'model_reasoning_effort=\"medium\"'" +
		" -C " + workDir + " -\n"
	wantPath := "\x1b[35mecosystem.config.js path:\x1b[0m\n  " + configPath + "\n"
	wantConfiguration := "\x1b[35mecosystem.config.js configuration:\x1b[0m\n" +
		configContent
	output := wizardOutput.String()
	for _, want := range []string{wantOriginal, wantExecute, wantPath, wantConfiguration} {
		if !strings.Contains(output, want) {
			t.Fatalf("wizard output is missing summary %q:\n%s", want, output)
		}
	}
}

func TestWizardCommandWritesEcosystemAtWorkspaceRoot(t *testing.T) {
	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, testSettings(), dependencies)
	command.SetIn(strings.NewReader("agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\n\n"))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	configPath := filepath.Join(workDir, "ecosystem.config.js")
	got := readTestFile(t, configPath)
	if !strings.Contains(got, "cwd: "+strconv.Quote(workDir)) {
		t.Fatalf("ecosystem config has wrong workspace cwd:\n%s", got)
	}
	if !strings.Contains(got, `name: "Autop agy `+filepath.Base(workDir)+`"`) {
		t.Fatalf("ecosystem config has wrong project task name:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(workDir, "cmd", "ecosystem.config.js")); !os.IsNotExist(err) {
		t.Fatalf("wizard unexpectedly wrote cmd/ecosystem.config.js: %v", err)
	}
}

func TestWizardCommandWritesOptionalCronSchedule(t *testing.T) {
	workDir := t.TempDir()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, testSettings(), dependencies)
	command.SetIn(strings.NewReader(
		"agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\n0 3 * * *\n",
	))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := readTestFile(t, filepath.Join(workDir, "ecosystem.config.js"))
	if !strings.Contains(got, `cron: "0 3 * * *"`) {
		t.Fatalf("ecosystem config is missing selected cron schedule:\n%s", got)
	}
}

func TestWizardCommandWritesRandomDailyCronWithinWindow(t *testing.T) {
	workDir := t.TempDir()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, testSettings(), dependencies)
	command.SetIn(strings.NewReader("agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\nr\n"))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := readTestFile(t, filepath.Join(workDir, "ecosystem.config.js"))
	marker := `cron: "`
	start := strings.Index(got, marker)
	if start < 0 {
		t.Fatalf("ecosystem config is missing random cron schedule:\n%s", got)
	}
	start += len(marker)
	end := strings.Index(got[start:], `"`)
	if end < 0 {
		t.Fatalf("ecosystem config has invalid cron string:\n%s", got)
	}
	fields := strings.Fields(got[start : start+end])
	if len(fields) != 5 {
		t.Fatalf("random cron fields = %#v, want 5 fields", fields)
	}
	minute, err := strconv.Atoi(fields[0])
	if err != nil {
		t.Fatalf("random cron minute = %q: %v", fields[0], err)
	}
	hour, err := strconv.Atoi(fields[1])
	if err != nil {
		t.Fatalf("random cron hour = %q: %v", fields[1], err)
	}
	if hour < 2 || hour > 8 || minute < 0 || minute > 59 || (hour == 8 && minute != 0) {
		t.Fatalf("random cron time = %02d:%02d, want 02:00 through 08:00", hour, minute)
	}
	if strings.Join(fields[2:], " ") != "* * *" {
		t.Fatalf("random cron recurrence = %q, want daily", strings.Join(fields[2:], " "))
	}
}

func TestWizardCommandRequiresValidCustomCronFormat(t *testing.T) {
	workDir := t.TempDir()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, testSettings(), dependencies)
	var wizardOutput bytes.Buffer
	command.SetIn(strings.NewReader(
		"agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\n0 3 * *\n0 4 * * *\n",
	))
	command.SetOut(&wizardOutput)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(wizardOutput.String(), "must be a valid 5-field cron") {
		t.Fatalf("wizard did not reject incomplete cron:\n%s", wizardOutput.String())
	}
	got := readTestFile(t, filepath.Join(workDir, "ecosystem.config.js"))
	if !strings.Contains(got, `cron: "0 4 * * *"`) {
		t.Fatalf("ecosystem config is missing corrected cron schedule:\n%s", got)
	}
}

func TestWizardCommandRejectsCronWithInvalidCharacters(t *testing.T) {
	workDir := t.TempDir()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return workDir, nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, testSettings(), dependencies)
	var wizardOutput bytes.Buffer
	command.SetIn(strings.NewReader(
		"agy\nsystem\nno\ngemini-3.6-flash-high\nhigh\nreview current workspace\n0 3 * * $(date)\n*/15 2-8 * * MON-FRI\n",
	))
	command.SetOut(&wizardOutput)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"wizard"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(wizardOutput.String(), "must be a valid 5-field cron") {
		t.Fatalf("wizard did not reject unsafe cron characters:\n%s", wizardOutput.String())
	}
	got := readTestFile(t, filepath.Join(workDir, "ecosystem.config.js"))
	if !strings.Contains(got, `cron: "*/15 2-8 * * MON-FRI"`) {
		t.Fatalf("ecosystem config is missing corrected cron schedule:\n%s", got)
	}
	if strings.Contains(got, "$(date)") {
		t.Fatalf("ecosystem config contains rejected cron input:\n%s", got)
	}
}

func TestWizardCommandExposesOnlyWAlias(t *testing.T) {
	if len(WizardCmd.Aliases) != 1 || WizardCmd.Aliases[0] != "w" {
		t.Fatalf("wizard aliases = %#v, want []string{\"w\"}", WizardCmd.Aliases)
	}
}

func TestRootCommandRejectsPromptFromArgumentsAndStdin(t *testing.T) {
	settings := testSettings()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run:       runProcess,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	command.SetIn(bytes.NewBufferString("stdin prompt"))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"argument prompt"})

	err := command.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want conflicting-prompt error")
	}
	if !strings.Contains(err.Error(), "not both") {
		t.Fatalf("error = %q, want input conflict", err)
	}
}

func TestRootCommandAppliesWizardRuntimeFlags(t *testing.T) {
	settings := testSettings()
	var captured autopdriver.Process
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			process autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			captured = process
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{
		"-c",
		"agy",
		"--bypass-permission=false",
		"--model",
		"gemini-3.6-flash-low",
		"--effort",
		"low",
		"--",
		"review workspace",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := strings.Join(captured.Args, " ")
	if strings.Contains(got, "--dangerously-skip-permissions") {
		t.Fatalf("process bypasses permissions when wizard selected no: %s", got)
	}
	if !strings.Contains(got, "--model=gemini-3.6-flash-low") ||
		!strings.Contains(got, "--effort=low") {
		t.Fatalf("process did not apply model and effort overrides: %s", got)
	}
	if len(captured.Args) == 0 || captured.Args[len(captured.Args)-1] != "review workspace" {
		t.Fatalf("process did not preserve terminated prompt: %#v", captured.Args)
	}
}

func TestRootCommandDryRunPrintsCommandWithoutExecuting(t *testing.T) {
	settings := testSettings()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			_ autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			t.Fatal("dry run must not execute the child process")
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	var out bytes.Buffer
	command.SetIn(strings.NewReader(""))
	command.SetOut(&out)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--dry-run", "-c", "codex", "--", "review workspace"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := strings.TrimSpace(out.String())
	if !strings.HasPrefix(got, "printf '%s' 'review workspace' | codex exec") {
		t.Fatalf("dry run output = %q, want shell-safe codex command line", got)
	}
}

func TestRootCommandDryRunSkipsPreflight(t *testing.T) {
	settings := testSettings()
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return "/workspace", nil },
		lookupEnv: missingEnv,
		run: func(
			_ context.Context,
			_ autopdriver.Process,
			_ io.Writer,
			_ io.Writer,
			_ *slog.Logger,
		) error {
			t.Fatal("dry run must not execute the child process")
			return nil
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	command := rootCommandForTest(t, settings, dependencies)
	var out bytes.Buffer
	command.SetIn(strings.NewReader(""))
	command.SetOut(&out)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--dry-run", "-c", "claudem", "--", "review workspace"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want dry run to skip credential preflight", err)
	}
	got := strings.TrimSpace(out.String())
	if !strings.HasPrefix(got, `ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude`) {
		t.Fatalf("dry run output = %q, want credential preview prefix", got)
	}
	if strings.Contains(got, "MINIMAX_API_KEY=") &&
		!strings.Contains(got, `"$MINIMAX_API_KEY"`) {
		t.Fatalf("dry run output leaked a resolved credential: %q", got)
	}
}
