package autop

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	autopdriver "github.com/bizshuk/autop/cmd/driver"
	gosdkconfig "github.com/bizshuk/gosdk/config"
)

// updateCommandMatrix regenerates the checked-in command matrix instead of
// comparing against it: go test ./cmd/... -run CommandMatrix -update-matrix
var updateCommandMatrix = flag.Bool(
	"update-matrix",
	false,
	"rewrite docs/command-matrix.md from the current driver mapping",
)

const (
	matrixDocPath  = "../docs/command-matrix.md"
	matrixWorkDir  = "/workspace"
	matrixPrompt   = "summarize current repository"
	matrixExtra    = "focus on dependency boundaries"
	matrixStdinTxt = "review the staged diff"
)

type matrixCase struct {
	title string
	// args are appended after "-c <client>"; {{model}} and {{effort}} are
	// replaced with the client's second model and lowest effort choice.
	args  []string
	stdin string
}

var matrixCases = []matrixCase{
	{title: "預設 model 與 effort", args: []string{"--", matrixPrompt}},
	{
		title: "--bypass-permission=true",
		args:  []string{"--bypass-permission=true", "--", matrixPrompt},
	},
	{
		title: "--model 覆寫",
		args:  []string{"--model", "{{model}}", "--", matrixPrompt},
	},
	{
		title: "--effort 覆寫",
		args:  []string{"--effort", "{{effort}}", "--", matrixPrompt},
	},
	{title: "-t system (template 無 prompt)", args: []string{"-t", "system"}},
	{
		title: "-t system + prompt",
		args:  []string{"-t", "system", "--", matrixExtra},
	},
	{
		title: "全 flag 組合",
		args: []string{
			"-t", "auto-evolving",
			"--bypass-permission=true",
			"--model", "{{model}}",
			"--effort", "{{effort}}",
			"--", matrixExtra,
		},
	},
	{title: "prompt 由 stdin 傳入", stdin: matrixStdinTxt},
}

// TestCommandMatrix renders every client x flag combination through the real
// root command and asserts the executed command line matches the checked-in
// docs/command-matrix.md.
func TestCommandMatrix(t *testing.T) {
	settings := defaultSettings()
	requireSettingsFiles(t, settings)
	generated := renderCommandMatrix(t, settings)

	if *updateCommandMatrix {
		if err := os.MkdirAll(filepath.Dir(matrixDocPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(matrixDocPath, []byte(generated), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", matrixDocPath)
		return
	}

	want, err := os.ReadFile(matrixDocPath)
	if err != nil {
		t.Fatalf("read command matrix: %v (run with -update-matrix)", err)
	}
	if string(want) != generated {
		t.Fatalf(
			"%s is stale; rerun with -update-matrix\n--- generated ---\n%s",
			matrixDocPath,
			generated,
		)
	}
}

// requireSettingsFiles skips the matrix on a machine that does not carry the
// referenced provider settings files, because the run path preflights them.
func requireSettingsFiles(t *testing.T, settings Settings) {
	t.Helper()

	for _, clientID := range enabledClientIDs(settings) {
		path := settings.Clients[clientID].Settings
		if strings.TrimSpace(path) == "" {
			continue
		}
		if _, err := os.Stat(gosdkconfig.ExpandHome(path)); err != nil {
			t.Skipf("client %q settings file is unavailable here: %v", clientID, err)
		}
	}
}

func renderCommandMatrix(t *testing.T, settings Settings) string {
	t.Helper()

	var doc strings.Builder
	doc.WriteString("# Autop command matrix\n\n")
	doc.WriteString(
		"由 `go test ./cmd/... -run CommandMatrix -update-matrix` 產生，" +
			"請勿手動編輯。\n\n" +
			"每組顯示 `autop` 原始命令與 driver mapping 後實際執行的 command line。" +
			"Credential 以 `TARGET=\"$SOURCE\"` 形式呈現，不含 secret 值。" +
			"Working directory 固定為 `" + matrixWorkDir + "`。\n",
	)

	for _, clientID := range enabledClientIDs(settings) {
		client := settings.Clients[clientID]
		fmt.Fprintf(
			&doc,
			"\n## %s\n\ndriver `%s` / command `%s` / credential `%s`\n",
			clientID,
			client.Driver,
			client.Command,
			credentialSummary(client.Credential),
		)
		for _, testCase := range matrixCases {
			autopArgs, executed := runMatrixCase(t, settings, clientID, client, testCase)
			fmt.Fprintf(
				&doc,
				"\n### %s\n\n```sh\n%s\n```\n\n```sh\n%s\n```\n",
				testCase.title,
				formatCommand(autopdriver.Process{Name: "autop", Args: autopArgs}),
				executed,
			)
		}
	}
	return doc.String()
}

func runMatrixCase(
	t *testing.T,
	settings Settings,
	clientID string,
	client ClientConfig,
	testCase matrixCase,
) (autopArgs []string, executed string) {
	t.Helper()

	autopArgs = append([]string{"-c", clientID}, expandMatrixArgs(testCase.args, client)...)

	var captured autopdriver.Process
	dependencies := commandDependencies{
		getwd:     func() (string, error) { return matrixWorkDir, nil },
		lookupEnv: func(string) (string, bool) { return "redacted", true },
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

	resetRootCommandForTest()
	configureCommandRuntime(settings, dependencies)
	RootCmd.SetIn(strings.NewReader(testCase.stdin))
	RootCmd.SetOut(io.Discard)
	RootCmd.SetErr(io.Discard)
	RootCmd.SetArgs(autopArgs)
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("client %q case %q: %v", clientID, testCase.title, err)
	}
	resetRootCommandForTest()
	configureCommandRuntime(defaultSettings(), defaultCommandDependencies())

	return autopArgs, collapseHomePath(t, formatCommand(captured))
}

// collapseHomePath keeps the checked-in matrix machine independent: settings
// paths are expanded to an absolute home directory at runtime.
func collapseHomePath(t *testing.T, command string) string {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Fatalf("resolve home directory: %v", err)
	}
	return strings.ReplaceAll(command, home+"/", "~/")
}

func expandMatrixArgs(args []string, client ClientConfig) []string {
	models := clientModelChoices(client)
	efforts := clientEffortChoices(client)
	replacements := strings.NewReplacer(
		"{{model}}", models[min(1, len(models)-1)],
		"{{effort}}", efforts[len(efforts)-1],
	)
	expanded := make([]string, 0, len(args))
	for _, arg := range args {
		expanded = append(expanded, replacements.Replace(arg))
	}
	return expanded
}

func credentialSummary(credential CredentialConfig) string {
	if credential.Mode != "env" {
		return credential.Mode
	}
	return credential.SourceEnv + " → " + credential.TargetEnv
}
