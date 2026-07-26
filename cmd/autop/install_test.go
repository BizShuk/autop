package autop

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestInstallEcosystemCreatesFileFromClientAndTemplate(t *testing.T) {
	workDir := t.TempDir()
	path := filepath.Join(workDir, "ecosystem.config.js")

	if err := installEcosystem(path, "codex", "system"); err != nil {
		t.Fatalf("installEcosystem() error = %v", err)
	}
	got := readTestFile(t, path)
	for _, want := range []string{
		`name: "AutoP codex system"`,
		`script: "autop"`,
		`args: ["-c", "codex", "-t", "system"]`,
		"cwd: " + strconv.Quote(workDir),
		`// autop:begin autop-codex-system`,
		`// autop:end autop-codex-system`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("ecosystem config missing %q:\n%s", want, got)
		}
	}
}

func TestInstallEcosystemWritesPromptAsPositionalArgument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")

	if err := installEcosystemTask(path, ecosystemTask{
		ClientID:              "codex",
		TemplateID:            "system",
		Prompt:                "review current workspace",
		BypassPermission:      true,
		Model:                 "gpt-5.5",
		Effort:                "high",
		IncludeRuntimeOptions: true,
	}); err != nil {
		t.Fatalf("installEcosystemTask() error = %v", err)
	}

	got := readTestFile(t, path)
	want := `args: ["-c", "codex", "-t", "system", "--bypass-permission=true", "--model", "gpt-5.5", "--effort", "high", "--", "review current workspace"]`
	if !strings.Contains(got, want) {
		t.Fatalf("ecosystem config missing prompt argument %q:\n%s", want, got)
	}
}

func TestInstallEcosystemRequiresPromptForTemplateFreeWizardTask(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")
	err := installEcosystemTask(path, ecosystemTask{
		ClientID:              "agy",
		BypassPermission:      true,
		Model:                 "gemini-3.6-flash-high",
		Effort:                "high",
		IncludeRuntimeOptions: true,
	})
	if err == nil || !strings.Contains(err.Error(), "prompt is required") {
		t.Fatalf("installEcosystemTask() error = %v, want required-prompt error", err)
	}
}

func TestInstallEcosystemPreservesExistingApps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")
	existing := `module.exports = {
    apps: [
        {
            name: "Existing",
            script: "existing"
        }
    ]
};
`
	if err := os.WriteFile(path, []byte(existing), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := installEcosystem(path, "agy", ""); err != nil {
		t.Fatalf("installEcosystem() error = %v", err)
	}
	got := readTestFile(t, path)
	if !strings.Contains(got, `name: "Existing"`) {
		t.Fatalf("existing app was lost:\n%s", got)
	}
	if !strings.Contains(got, `args: ["-c", "agy"]`) {
		t.Fatalf("autop args missing or include unexpected template:\n%s", got)
	}
	if !strings.Contains(got, `name: "AutoP agy"`) {
		t.Fatalf("autop display name has wrong casing or fields:\n%s", got)
	}
	if strings.Contains(got, `"-t"`) {
		t.Fatalf("ecosystem config contains -t without template:\n%s", got)
	}
}

func TestInstallEcosystemIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")

	if err := installEcosystem(path, "codex", "system"); err != nil {
		t.Fatal(err)
	}
	if err := installEcosystem(path, "codex", "system"); err != nil {
		t.Fatal(err)
	}
	got := readTestFile(t, path)
	if count := strings.Count(got, "// autop:begin autop-codex-system"); count != 1 {
		t.Fatalf("managed task count = %d, want 1:\n%s", count, got)
	}
}

func TestInstallEcosystemSupportsMultipleManagedTasks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")

	if err := installEcosystem(path, "codex", "system"); err != nil {
		t.Fatal(err)
	}
	if err := installEcosystem(path, "agy", ""); err != nil {
		t.Fatal(err)
	}
	got := readTestFile(t, path)
	if strings.Contains(got, "// autop:end autop-codex-system,") {
		t.Fatalf("separator was appended inside a line comment:\n%s", got)
	}
	if !strings.Contains(got, "},\n        // autop:end autop-codex-system") {
		t.Fatalf("first managed app has no JavaScript separator:\n%s", got)
	}
	if strings.Count(got, "// autop:begin ") != 2 {
		t.Fatalf("managed task count is not 2:\n%s", got)
	}
}

func TestInstallEcosystemDoesNotOverwriteInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecosystem.config.js")
	existing := "module.exports = { invalid: true };\n"
	if err := os.WriteFile(path, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	err := installEcosystem(path, "codex", "system")
	if err == nil {
		t.Fatal("installEcosystem() error = nil, want apps-array error")
	}
	if got := readTestFile(t, path); got != existing {
		t.Fatalf("invalid ecosystem config was overwritten:\n%s", got)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
