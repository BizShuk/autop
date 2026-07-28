package autop

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	autopdriver "github.com/bizshuk/autop/cmd/driver"
)

func TestRunProcessLogsCommandBeforeExecution(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "fake-client")
	script := "#!/bin/sh\nprintf 'ran:%s' \"$1\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	var stdout bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	process := autopdriver.Process{
		Name: scriptPath,
		Args: []string{"hello world"},
	}

	if err := runProcess(context.Background(), process, &stdout, &logs, logger); err != nil {
		t.Fatalf("runProcess() error = %v", err)
	}
	if stdout.String() != "ran:hello world" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	logOutput := logs.String()
	if !strings.Contains(logOutput, "executing command") {
		t.Fatalf("log = %q, want executing command", logOutput)
	}
	if !strings.Contains(logOutput, `hello world`) {
		t.Fatalf("log = %q, want command argument", logOutput)
	}
}

func TestFormatCommandDoesNotIncludeEnvironment(t *testing.T) {
	process := autopdriver.Process{
		Name:     "claude",
		Args:     []string{"--model", "opus"},
		ExtraEnv: []string{"ANTHROPIC_AUTH_TOKEN=top-secret"},
	}

	got := formatCommand(process)
	if strings.Contains(got, "top-secret") || strings.Contains(got, "ANTHROPIC_AUTH_TOKEN") {
		t.Fatalf("formatCommand() leaked environment: %q", got)
	}
	if got != `claude --model opus` {
		t.Fatalf("formatCommand() = %q", got)
	}
}

func TestFormatCommandIncludesStdinPipeline(t *testing.T) {
	process := autopdriver.Process{
		Name:  "codex",
		Args:  []string{"exec", "-c", `model_reasoning_effort="medium"`, "-"},
		Stdin: "$find-activity find 5 event s in EU",
	}

	got := formatCommand(process)
	want := `printf '%s' '$find-activity find 5 event s in EU' | codex exec -c 'model_reasoning_effort="medium"' -`
	if got != want {
		t.Fatalf("formatCommand() = %q, want %q", got, want)
	}
}

func TestFormatCommandEscapesSingleQuotes(t *testing.T) {
	process := autopdriver.Process{
		Name:  "codex",
		Args:  []string{"exec", "-"},
		Stdin: "don't expand $skill",
	}

	got := formatCommand(process)
	want := `printf '%s' 'don'"'"'t expand $skill' | codex exec -`
	if got != want {
		t.Fatalf("formatCommand() = %q, want %q", got, want)
	}
}

func TestRunProcessPreservesChildExitCode(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "failing-client")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 7\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := runProcess(
		context.Background(),
		autopdriver.Process{Name: scriptPath},
		io.Discard,
		io.Discard,
		logger,
	)
	if err == nil {
		t.Fatal("runProcess() error = nil, want child failure")
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("error = %v, want exec.ExitError", err)
	}
	if got := CommandExitCode(err); got != 7 {
		t.Fatalf("CommandExitCode() = %d, want 7", got)
	}
}
