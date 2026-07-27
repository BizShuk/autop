package autop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderPromptWithoutTemplateReturnsRawPrompt(t *testing.T) {
	got, err := renderPrompt(testSettings(), "", "keep $HOME and 'quotes'", "/workspace", "codex")
	if err != nil {
		t.Fatalf("renderPrompt() error = %v", err)
	}
	if got != "keep $HOME and 'quotes'" {
		t.Fatalf("renderPrompt() = %q", got)
	}
}

func TestRenderPromptSupportsTemplateOnlyTask(t *testing.T) {
	got, err := renderPrompt(testSettings(), "system", "", "/workspace", "codex")
	if err != nil {
		t.Fatalf("renderPrompt() error = %v", err)
	}
	if got != "run $system-planner for current workspace" {
		t.Fatalf("renderPrompt() = %q", got)
	}
}

func TestDefaultTemplatesUseDriverSkillPrefixAndNames(t *testing.T) {
	settings := defaultSettings()
	if _, ok := settings.Templates["auto-evolving"]; !ok {
		t.Fatal("default templates do not include auto-evolving")
	}
	if _, ok := settings.Templates["auto-evolve"]; ok {
		t.Fatal("default templates still include the old auto-evolve ID")
	}

	codexSystem, err := renderPrompt(settings, "system", "", "/workspace", "codex")
	if err != nil {
		t.Fatalf("renderPrompt(codex system) error = %v", err)
	}
	if codexSystem != "run $system-planner for current workspace" {
		t.Fatalf("codex system prompt = %q", codexSystem)
	}

	agySystem, err := renderPrompt(settings, "system", "", "/workspace", "agy")
	if err != nil {
		t.Fatalf("renderPrompt(agy system) error = %v", err)
	}
	if agySystem != "run /system-planner for current workspace" {
		t.Fatalf("agy system prompt = %q", agySystem)
	}

	codexEvolving, err := renderPrompt(settings, "auto-evolving", "", "/workspace", "codex")
	if err != nil {
		t.Fatalf("renderPrompt(codex auto-evolving) error = %v", err)
	}
	if codexEvolving != "run $auto-evolving for current workspace" {
		t.Fatalf("codex auto-evolving prompt = %q", codexEvolving)
	}
}

func TestRenderPromptLoadsTemplateFile(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "prompt.tmpl")
	if err := os.WriteFile(templatePath, []byte("client={{.Client}}\n{{.Prompt}}"), 0o600); err != nil {
		t.Fatal(err)
	}
	settings := testSettings()
	settings.Templates["file"] = PromptTemplate{File: templatePath}

	got, err := renderPrompt(settings, "file", "hello", "/workspace", "agy")
	if err != nil {
		t.Fatalf("renderPrompt() error = %v", err)
	}
	if got != "client=agy\nhello" {
		t.Fatalf("renderPrompt() = %q", got)
	}
}

func TestRenderPromptRejectsUnknownTemplate(t *testing.T) {
	_, err := renderPrompt(testSettings(), "missing", "hello", "/workspace", "codex")
	if err == nil {
		t.Fatal("renderPrompt() error = nil, want unknown-template error")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("error = %q, want template ID", err)
	}
}

func TestRenderPromptRejectsEmptyPromptWithoutTemplate(t *testing.T) {
	_, err := renderPrompt(testSettings(), "", " \n", "/workspace", "codex")
	if err == nil {
		t.Fatal("renderPrompt() error = nil, want empty-prompt error")
	}
}
