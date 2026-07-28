package driver

import (
	"reflect"
	"strings"
	"testing"
)

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
	client := ClientConfig{
		Driver:      "claude",
		Command:     "claudem",
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
		"/tmp/minimax.json",
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
	if got.Name != "claudem" || !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("process = %#v, want name claudem args %#v", got, wantArgs)
	}
	if !reflect.DeepEqual(got.ExtraEnv, []string{"ANTHROPIC_AUTH_TOKEN=top-secret"}) {
		t.Fatalf("ExtraEnv = %#v", got.ExtraEnv)
	}
}

func TestBuildProcessDoesNotResolveCredentialEnvironment(t *testing.T) {
	client := ClientConfig{
		Driver:      "claude",
		Command:     "claudem",
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
	if got.Name != "claudem" {
		t.Fatalf("process.Name = %q, want claudem", got.Name)
	}
	if len(got.ExtraEnv) != 0 {
		t.Fatalf("process.ExtraEnv = %#v, want no resolved credentials", got.ExtraEnv)
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

func missingEnv(string) (string, bool) {
	return "", false
}
