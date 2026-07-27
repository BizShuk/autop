package autop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gosdkconfigcmd "github.com/bizshuk/gosdk/cmd/config"
	gosdkconfig "github.com/bizshuk/gosdk/config"
	"github.com/spf13/viper"
)

func TestResolveClientUsesDefault(t *testing.T) {
	settings := testSettings()

	clientID, client, err := resolveClient(settings, "")
	if err != nil {
		t.Fatalf("resolveClient() error = %v", err)
	}
	if clientID != "codex" {
		t.Fatalf("clientID = %q, want codex", clientID)
	}
	if client.Driver != "codex" {
		t.Fatalf("client.Driver = %q, want codex", client.Driver)
	}
}

func TestResolveClientRejectsDisabledClient(t *testing.T) {
	settings := testSettings()

	_, _, err := resolveClient(settings, "claudet")
	if err == nil {
		t.Fatal("resolveClient() error = nil, want disabled-client error")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("error = %q, want disabled", err)
	}
}

func TestValidateSettingsRejectsUnknownDriver(t *testing.T) {
	settings := testSettings()
	client := settings.Clients["codex"]
	client.Driver = "unknown"
	settings.Clients["codex"] = client

	err := validateSettings(settings)
	if err == nil {
		t.Fatal("validateSettings() error = nil, want unknown-driver error")
	}
	if !strings.Contains(err.Error(), "driver") {
		t.Fatalf("error = %q, want driver context", err)
	}
}

func TestRegisteredDefaultsProduceValidSettings(t *testing.T) {
	defaultViper := viper.New()
	registerDefaults(defaultViper)

	var settings Settings
	if err := defaultViper.Unmarshal(&settings); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if err := validateSettings(settings); err != nil {
		t.Fatalf("validateSettings() error = %v", err)
	}
	wantClients := []string{"agy", "claude", "claudem", "claudep", "claudet", "claudew", "codex"}
	gotClients := make([]string, 0, len(settings.Clients))
	for _, clientID := range wantClients {
		if _, ok := settings.Clients[clientID]; ok {
			gotClients = append(gotClients, clientID)
		}
	}
	if !reflect.DeepEqual(gotClients, wantClients) {
		t.Fatalf("configured clients = %#v, want %#v", gotClients, wantClients)
	}
	if !settings.Clients["claudet"].Disabled {
		t.Fatal("claudet must stay disabled until its settings and credential contract are configured")
	}
	wantAgyModels := []string{
		"gemini-3.6-flash-high",
		"gemini-3.6-flash-medium",
		"gemini-3.6-flash-low",
		"gemini-3.5-flash-high",
		"gemini-3.5-flash-medium",
		"gemini-3.5-flash-low",
		"gemini-3.1-pro-high",
		"gemini-3.1-pro-low",
		"claude-sonnet-4-6",
		"claude-opus-4-6-thinking",
		"gpt-oss-120b-medium",
	}
	if !reflect.DeepEqual(settings.Clients["agy"].Models, wantAgyModels) {
		t.Fatalf(
			"agy model choices = %#v, want %#v",
			settings.Clients["agy"].Models,
			wantAgyModels,
		)
	}
	if !contains(settings.Clients["codex"].Models, "gpt-5.5") {
		t.Fatalf(
			"codex model choices = %#v, want gpt-5.5",
			settings.Clients["codex"].Models,
		)
	}
	for clientID, client := range settings.Clients {
		if client.Disabled {
			continue
		}
		if len(client.Models) == 0 || !contains(client.Models, client.Model) {
			t.Errorf(
				"client %q model choices = %#v, want current model %q",
				clientID,
				client.Models,
				client.Model,
			)
		}
		if len(client.Efforts) == 0 || !contains(client.Efforts, client.Effort) {
			t.Errorf(
				"client %q effort choices = %#v, want current effort %q",
				clientID,
				client.Efforts,
				client.Effort,
			)
		}
	}

	profileChecks := map[string]struct {
		driver     string
		command    string
		model      string
		effort     string
		credential CredentialConfig
	}{
		"agy": {
			driver:     "agy",
			command:    "agy",
			model:      "gemini-3.6-flash-high",
			effort:     "high",
			credential: CredentialConfig{Mode: "oauth"},
		},
		"codex": {
			driver:     "codex",
			command:    "codex",
			model:      "gpt-5.6-sol",
			effort:     "xhigh",
			credential: CredentialConfig{Mode: "oauth"},
		},
		"claude": {
			driver:     "claude",
			command:    "claude",
			model:      "opus",
			effort:     "max",
			credential: CredentialConfig{Mode: "oauth"},
		},
		"claudem": {
			driver:  "claude",
			command: "claudem",
			model:   "MiniMax-M3",
			effort:  "xhigh",
			credential: CredentialConfig{
				Mode:      "env",
				SourceEnv: "MINIMAX_API_KEY",
				TargetEnv: "ANTHROPIC_AUTH_TOKEN",
			},
		},
		"claudew": {
			driver:  "claude",
			command: "claudew",
			model:   "minimax-m3",
			effort:  "xhigh",
			credential: CredentialConfig{
				Mode:      "env",
				SourceEnv: "TIKTOK_API_KEY",
				TargetEnv: "ANTHROPIC_AUTH_TOKEN",
			},
		},
		"claudep": {
			driver:  "claude",
			command: "claude",
			model:   "gemini-3.5-flash",
			effort:  "high",
			credential: CredentialConfig{
				Mode:      "env",
				SourceEnv: "AGENTSDK_PROXY_API_KEY",
				TargetEnv: "ANTHROPIC_AUTH_TOKEN",
			},
		},
	}
	for clientID, want := range profileChecks {
		client := settings.Clients[clientID]
		if client.Driver != want.driver ||
			client.Command != want.command ||
			client.Model != want.model ||
			client.Effort != want.effort ||
			client.Credential != want.credential {
			t.Errorf("client %q = %#v, want %#v", clientID, client, want)
		}
	}
}

func TestClientEffortChoicesSortsFromMaximumToLow(t *testing.T) {
	tests := []struct {
		name   string
		client ClientConfig
		want   []string
	}{
		{
			name: "codex",
			client: ClientConfig{
				Driver:  "codex",
				Effort:  "xhigh",
				Efforts: []string{"low", "medium", "high", "xhigh", "max"},
			},
			want: []string{"max", "xhigh", "high", "medium", "low"},
		},
		{
			name: "agy supported subset",
			client: ClientConfig{
				Driver:  "agy",
				Effort:  "high",
				Efforts: []string{"low", "medium", "high"},
			},
			want: []string{"high", "medium", "low"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := clientEffortChoices(test.client)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("clientEffortChoices() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestExampleSettingsIsValid(t *testing.T) {
	content, err := os.ReadFile("settings.example.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var settings Settings
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if err := validateSettings(settings); err != nil {
		t.Fatalf("validateSettings() error = %v", err)
	}
}

func TestExampleSettingsIsRegisteredAsSDKDefault(t *testing.T) {
	if !contains(gosdkconfigcmd.RegisteredDefaults(), "settings.json") {
		t.Fatalf("registered defaults = %#v, want settings.json", gosdkconfigcmd.RegisteredDefaults())
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	previousAppName := gosdkconfig.GetAppName()
	gosdkconfig.SetAppName(appName)
	t.Cleanup(func() { gosdkconfig.SetAppName(previousAppName) })

	report, err := gosdkconfigcmd.Default(
		"settings.json",
		gosdkconfigcmd.DefaultModeSkip,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if !report.Registered || !report.Written {
		t.Fatalf("Default() report = %#v, want registered and written", report)
	}

	got, err := os.ReadFile(filepath.Join(home, ".config", appName, "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want, err := os.ReadFile("settings.example.json")
	if err != nil {
		t.Fatalf("ReadFile(settings.example.json) error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("seeded settings differ from settings.example.json")
	}
}

func TestClientLeafOverrideKeepsOtherDefaults(t *testing.T) {
	defaultViper := viper.New()
	registerDefaults(defaultViper)
	if err := defaultViper.MergeConfigMap(map[string]any{
		"default_client": "agy",
		"clients": map[string]any{
			"codex": map[string]any{
				"model":  "gpt-custom",
				"effort": "high",
			},
		},
	}); err != nil {
		t.Fatalf("MergeConfigMap() error = %v", err)
	}

	var settings Settings
	if err := defaultViper.Unmarshal(&settings); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if settings.DefaultClient != "agy" {
		t.Fatalf("DefaultClient = %q, want agy", settings.DefaultClient)
	}
	codex := settings.Clients["codex"]
	if codex.Command != "codex" || codex.Driver != "codex" || !codex.AutoApprove {
		t.Fatalf("codex defaults were lost: %#v", codex)
	}
	if codex.Model != "gpt-custom" || codex.Effort != "high" {
		t.Fatalf("codex overrides were not applied: %#v", codex)
	}
}

func testSettings() Settings {
	return Settings{
		DefaultClient: "codex",
		Clients: map[string]ClientConfig{
			"agy": {
				Driver:          "agy",
				Command:         "agy",
				AutoApprove:     true,
				Model:           defaultAgyModels[0],
				Models:          append([]string(nil), defaultAgyModels...),
				Effort:          "high",
				PromptTransport: "argument",
				Credential:      CredentialConfig{Mode: "oauth"},
			},
			"codex": {
				Driver:          "codex",
				Command:         "codex",
				AutoApprove:     true,
				Model:           "gpt-5.6-sol",
				Effort:          "xhigh",
				PromptTransport: "stdin",
				Credential:      CredentialConfig{Mode: "oauth"},
			},
			"claudem": {
				Driver:          "claude",
				Command:         "claudem",
				AutoApprove:     true,
				Model:           "MiniMax-M3",
				Effort:          "xhigh",
				Settings:        "/tmp/minimax.json",
				PromptTransport: "argument",
				Credential: CredentialConfig{
					Mode:      "env",
					SourceEnv: "MINIMAX_API_KEY",
					TargetEnv: "ANTHROPIC_AUTH_TOKEN",
				},
			},
			"claudet": {
				Disabled:        true,
				Driver:          "claude",
				Command:         "claude",
				PromptTransport: "argument",
			},
		},
		Templates: map[string]PromptTemplate{
			"system": {
				Content: "run {{if eq .Driver \"codex\"}}$system-planner{{else}}/system-planner{{end}} for current workspace{{if .Prompt}}\n\n{{.Prompt}}{{end}}",
			},
		},
	}
}

func missingEnv(string) (string, bool) {
	return "", false
}
