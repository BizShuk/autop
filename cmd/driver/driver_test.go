package driver

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeSettingsFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "minimax.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPrepareAgyProcess(t *testing.T) {
	client := ClientConfig{
		Driver:      "agy",
		Command:     "agy",
		AutoApprove: true,
		Model:       "gemini-3.6-flash-high",
		Effort:      "high",
		Credential:  CredentialConfig{Mode: "oauth"},
	}

	got, err := Prepare("agy", client, "do $(touch /tmp/nope)", "/workspace", missingEnv)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	wantArgs := []string{
		"--dangerously-skip-permissions",
		"--model=gemini-3.6-flash-high",
		"--effort=high",
		"--add-dir",
		"/workspace",
		"-p",
		"do $(touch /tmp/nope)",
	}
	if got.Name != "agy" || !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("process = %#v, want name agy args %#v", got, wantArgs)
	}
	if got.Stdin != "" {
		t.Fatalf("Stdin = %q, want empty", got.Stdin)
	}
}

func TestPrepareClaudeProfileProcess(t *testing.T) {
	settingsPath := writeSettingsFile(t)
	client := ClientConfig{
		Driver:      "claude",
		Command:     "claude",
		AutoApprove: true,
		Model:       "MiniMax-M3",
		Effort:      "xhigh",
		Settings:    settingsPath,
		Credential: CredentialConfig{
			Mode:      "env",
			SourceEnv: "MINIMAX_API_KEY",
			TargetEnv: "ANTHROPIC_AUTH_TOKEN",
		},
	}
	getenv := func(key string) (string, bool) {
		if key == "MINIMAX_API_KEY" {
			return "top-secret", true
		}
		return "", false
	}

	got, err := Prepare("claudem", client, "review repo", "/workspace", getenv)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	wantArgs := []string{
		"--dangerously-skip-permissions",
		"--settings",
		settingsPath,
		"--model",
		"MiniMax-M3",
		"--effort",
		"xhigh",
		"--add-dir",
		"/workspace",
		"--output-format",
		"text",
		"-p",
		"review repo",
	}
	if got.Name != "claude" || !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("process = %#v, want name claude args %#v", got, wantArgs)
	}
	if !reflect.DeepEqual(got.ExtraEnv, []string{"ANTHROPIC_AUTH_TOKEN=top-secret"}) {
		t.Fatalf("ExtraEnv = %#v", got.ExtraEnv)
	}
	wantPreview := []string{`ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY"`}
	if !reflect.DeepEqual(got.EnvPreview, wantPreview) {
		t.Fatalf("EnvPreview = %#v, want %#v", got.EnvPreview, wantPreview)
	}
}

func TestBuildProcessOmitsEnvPreviewForOAuthCredential(t *testing.T) {
	client := ClientConfig{
		Driver:     "codex",
		Command:    "codex",
		Model:      "gpt-5.6-sol",
		Effort:     "xhigh",
		Credential: CredentialConfig{Mode: "oauth"},
	}

	got, err := BuildProcess("codex", client, "review repo", "/workspace")
	if err != nil {
		t.Fatalf("BuildProcess() error = %v", err)
	}
	if len(got.EnvPreview) != 0 {
		t.Fatalf("process.EnvPreview = %#v, want none for oauth", got.EnvPreview)
	}
}

func TestBuildProcessDoesNotResolveCredentialEnvironment(t *testing.T) {
	client := ClientConfig{
		Driver:      "claude",
		Command:     "claude",
		AutoApprove: true,
		Model:       "MiniMax-M3",
		Effort:      "xhigh",
		Settings:    "/tmp/minimax.json",
		Credential: CredentialConfig{
			Mode:      "env",
			SourceEnv: "MINIMAX_API_KEY",
			TargetEnv: "ANTHROPIC_AUTH_TOKEN",
		},
	}

	got, err := BuildProcess("claudem", client, "review repo", "/workspace")
	if err != nil {
		t.Fatalf("BuildProcess() error = %v", err)
	}
	if got.Name != "claude" {
		t.Fatalf("process.Name = %q, want claude", got.Name)
	}
	if len(got.ExtraEnv) != 0 {
		t.Fatalf("process.ExtraEnv = %#v, want no resolved credentials", got.ExtraEnv)
	}
	wantPreview := []string{`ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY"`}
	if !reflect.DeepEqual(got.EnvPreview, wantPreview) {
		t.Fatalf("process.EnvPreview = %#v, want %#v", got.EnvPreview, wantPreview)
	}
}

func TestPrepareCodexProcessUsesStdin(t *testing.T) {
	client := ClientConfig{
		Driver:      "codex",
		Command:     "codex",
		AutoApprove: true,
		Model:       "gpt-5.6-sol",
		Effort:      "xhigh",
		Credential:  CredentialConfig{Mode: "oauth"},
	}

	got, err := Prepare("codex", client, "implement feature", "/workspace", missingEnv)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	wantArgs := []string{
		"exec",
		"--dangerously-bypass-approvals-and-sandbox",
		"--model",
		"gpt-5.6-sol",
		"-c",
		`model_reasoning_effort="xhigh"`,
		"-C",
		"/workspace",
		"-",
	}
	if got.Name != "codex" || !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("process = %#v, want name codex args %#v", got, wantArgs)
	}
	if got.Stdin != "implement feature" {
		t.Fatalf("Stdin = %q", got.Stdin)
	}
}

func TestPrepareCodexProcessConvertsSlashSkillPrefix(t *testing.T) {
	client := ClientConfig{
		Driver:     "codex",
		Command:    "codex",
		Model:      "gpt-5.6-sol",
		Effort:     "xhigh",
		Credential: CredentialConfig{Mode: "oauth"},
	}

	got, err := Prepare("codex", client, "/xxxxxx review workspace", "/workspace", missingEnv)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if got.Stdin != "$xxxxxx review workspace" {
		t.Fatalf("Stdin = %q, want slash skill prefix converted for Codex", got.Stdin)
	}
}

func TestPrepareGrokProcess(t *testing.T) {
	client := ClientConfig{
		Driver:      "grok",
		Command:     "grok",
		AutoApprove: true,
		Model:       "grok-4.5",
		Effort:      "high",
		Credential:  CredentialConfig{Mode: "oauth"},
	}

	got, err := Prepare("grok", client, "review workspace", "/workspace", missingEnv)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	wantArgs := []string{
		"--always-approve",
		"--permission-mode",
		"auto",
		"--model",
		"grok-4.5",
		"--reasoning-effort",
		"high",
		"--cwd",
		"/workspace",
		"--output-format",
		"plain",
		"--single",
		"review workspace",
	}
	if got.Name != "grok" || !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("process = %#v, want name grok args %#v", got, wantArgs)
	}
	if got.Stdin != "" {
		t.Fatalf("Stdin = %q, want empty", got.Stdin)
	}
}

func TestPrepareRejectsMissingCredentialWithoutLeakingSecret(t *testing.T) {
	client := ClientConfig{
		Driver:  "claude",
		Command: "claudem",
		Credential: CredentialConfig{
			Mode:      "env",
			SourceEnv: "MINIMAX_API_KEY",
			TargetEnv: "ANTHROPIC_AUTH_TOKEN",
		},
	}

	_, err := Prepare("claudem", client, "review", "/workspace", missingEnv)
	if err == nil {
		t.Fatal("Prepare() error = nil, want missing credential error")
	}
	if !strings.Contains(err.Error(), "MINIMAX_API_KEY") {
		t.Fatalf("error = %q, want source environment variable", err)
	}
	if strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("error leaked credential: %q", err)
	}
}

func TestPrepareRejectsMissingSettingsFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "absent.json")
	client := ClientConfig{
		Driver:     "claude",
		Command:    "claude",
		Model:      "opus",
		Effort:     "max",
		Settings:   missingPath,
		Credential: CredentialConfig{Mode: "oauth"},
	}

	_, err := Prepare("claude", client, "review", "/workspace", missingEnv)
	if err == nil {
		t.Fatal("Prepare() error = nil, want missing settings file error")
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Fatalf("error = %q, want settings path", err)
	}
}

func TestPrepareRejectsSettingsDirectory(t *testing.T) {
	client := ClientConfig{
		Driver:     "claude",
		Command:    "claude",
		Model:      "opus",
		Effort:     "max",
		Settings:   t.TempDir(),
		Credential: CredentialConfig{Mode: "oauth"},
	}

	_, err := Prepare("claude", client, "review", "/workspace", missingEnv)
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("Prepare() error = %v, want directory rejection", err)
	}
}

func TestPrepareSkipsSettingsCheckWhenNotRequired(t *testing.T) {
	client := ClientConfig{
		Driver:     "codex",
		Command:    "codex",
		Model:      "gpt-5.6-sol",
		Effort:     "xhigh",
		Credential: CredentialConfig{Mode: "oauth"},
	}

	if _, err := Prepare("codex", client, "review", "/workspace", missingEnv); err != nil {
		t.Fatalf("Prepare() error = %v, want no settings check", err)
	}
}

func TestBuildProcessSkipsSettingsFileCheck(t *testing.T) {
	client := ClientConfig{
		Driver:     "claude",
		Command:    "claude",
		Model:      "opus",
		Effort:     "max",
		Settings:   filepath.Join(t.TempDir(), "absent.json"),
		Credential: CredentialConfig{Mode: "oauth"},
	}

	if _, err := BuildProcess("claude", client, "review", "/workspace"); err != nil {
		t.Fatalf("BuildProcess() error = %v, want no preflight", err)
	}
}

func missingEnv(string) (string, bool) {
	return "", false
}
