package main

import (
	"fmt"
	"regexp"
	"strings"

	autopdriver "github.com/bizshuk/autop/driver"
	gosdkconfig "github.com/bizshuk/gosdk/config"
	"github.com/spf13/viper"
)

const appName = "autop"

var configIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var defaultAgyModels = []string{
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

// Settings is the complete runtime configuration for autop.
type Settings struct {
	DefaultClient string                    `json:"default_client" mapstructure:"default_client"`
	Clients       map[string]ClientConfig   `json:"clients" mapstructure:"clients"`
	Templates     map[string]PromptTemplate `json:"templates" mapstructure:"templates"`
}

type (
	ClientConfig     = autopdriver.ClientConfig
	CredentialConfig = autopdriver.CredentialConfig
)

// PromptTemplate is an inline or file-backed Go text template.
type PromptTemplate struct {
	Content string `json:"content,omitempty" mapstructure:"content"`
	File    string `json:"file,omitempty" mapstructure:"file"`
}

func loadSettings() (Settings, error) {
	registerDefaults(viper.GetViper())
	gosdkconfig.Default(gosdkconfig.WithAppName(appName))

	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return Settings{}, fmt.Errorf("decode autop settings: %w", err)
	}
	if err := validateSettings(settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func registerDefaults(v *viper.Viper) {
	defaults := defaultSettings()
	v.SetDefault("default_client", defaults.DefaultClient)
	for id, client := range defaults.Clients {
		prefix := "clients." + id + "."
		v.SetDefault(prefix+"disabled", client.Disabled)
		v.SetDefault(prefix+"driver", client.Driver)
		v.SetDefault(prefix+"command", client.Command)
		v.SetDefault(prefix+"auto_approve", client.AutoApprove)
		v.SetDefault(prefix+"model", client.Model)
		v.SetDefault(prefix+"models", client.Models)
		v.SetDefault(prefix+"effort", client.Effort)
		v.SetDefault(prefix+"efforts", client.Efforts)
		v.SetDefault(prefix+"settings", client.Settings)
		v.SetDefault(prefix+"prompt_transport", client.PromptTransport)
		v.SetDefault(prefix+"credential.mode", client.Credential.Mode)
		v.SetDefault(prefix+"credential.source_env", client.Credential.SourceEnv)
		v.SetDefault(prefix+"credential.target_env", client.Credential.TargetEnv)
	}
	for id, promptTemplate := range defaults.Templates {
		prefix := "templates." + id + "."
		v.SetDefault(prefix+"content", promptTemplate.Content)
		v.SetDefault(prefix+"file", promptTemplate.File)
	}
}

func defaultSettings() Settings {
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
				Efforts:         []string{"high", "medium", "low"},
				PromptTransport: "argument",
				Credential:      CredentialConfig{Mode: "oauth"},
			},
			"codex": {
				Driver:          "codex",
				Command:         "codex",
				AutoApprove:     true,
				Model:           "gpt-5.6-sol",
				Models:          []string{"gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.5", "gpt-5.3-codex-spark"},
				Effort:          "xhigh",
				Efforts:         []string{"max", "xhigh", "high", "medium", "low"},
				PromptTransport: "stdin",
				Credential:      CredentialConfig{Mode: "oauth"},
			},
			"claude": {
				Driver:          "claude",
				Command:         "claude",
				AutoApprove:     true,
				Model:           "opus",
				Models:          []string{"opus", "sonnet", "haiku"},
				Effort:          "max",
				Efforts:         []string{"max", "xhigh", "high", "medium", "low"},
				Settings:        "~/projects/cc-plugin/config/settings.json",
				PromptTransport: "argument",
				Credential:      CredentialConfig{Mode: "oauth"},
			},
			"claudem": {
				Driver:          "claude",
				Command:         "claude",
				AutoApprove:     true,
				Model:           "MiniMax-M3",
				Models:          []string{"MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.5"},
				Effort:          "xhigh",
				Efforts:         []string{"max", "xhigh", "high", "medium", "low"},
				Settings:        "~/projects/cc-plugin/config/minimax.json",
				PromptTransport: "argument",
				Credential: CredentialConfig{
					Mode:      "env",
					SourceEnv: "MINIMAX_API_KEY",
					TargetEnv: "ANTHROPIC_AUTH_TOKEN",
				},
			},
			"claudew": {
				Driver:          "claude",
				Command:         "claude",
				AutoApprove:     true,
				Model:           "minimax-m3",
				Models:          []string{"minimax-m3", "glm-5.2", "minimax-m2.7", "gpt-5.3-codex"},
				Effort:          "xhigh",
				Efforts:         []string{"max", "xhigh", "high", "medium", "low"},
				Settings:        "~/projects/cc-plugin/config/llmbox.json",
				PromptTransport: "argument",
				Credential: CredentialConfig{
					Mode:      "env",
					SourceEnv: "TIKTOK_API_KEY",
					TargetEnv: "ANTHROPIC_AUTH_TOKEN",
				},
			},
			"claudep": {
				Driver:      "claude",
				Command:     "claude",
				AutoApprove: true,
				Model:       "gemini-3.5-flash",
				Models: []string{
					"gemini-3.5-flash",
					"gpt-5.6-sol",
					"Minimax-M3",
					"claude-haiku-4-5",
					"grok-4.5-latest",
				},
				Effort:          "high",
				Efforts:         []string{"max", "xhigh", "high", "medium", "low"},
				Settings:        "~/projects/cc-plugin/config/proxy.json",
				PromptTransport: "argument",
				Credential: CredentialConfig{
					Mode:      "env",
					SourceEnv: "AGENTSDK_PROXY_API_KEY",
					TargetEnv: "ANTHROPIC_AUTH_TOKEN",
				},
			},
			"claudet": {
				Disabled:        true,
				Driver:          "claude",
				Command:         "claude",
				PromptTransport: "argument",
				Credential: CredentialConfig{
					Mode:      "env",
					TargetEnv: "ANTHROPIC_AUTH_TOKEN",
				},
			},
		},
		Templates: map[string]PromptTemplate{
			"system": {
				Content: "run {{if eq .Driver \"codex\"}}$system-planner{{else}}/system-planner{{end}} for current workspace{{if .Prompt}}\n\n{{.Prompt}}{{end}}",
			},
			"auto-evolving": {
				Content: "run {{if eq .Driver \"codex\"}}$auto-evolving{{else}}/auto-evolving{{end}} for current workspace{{if .Prompt}}\n\n{{.Prompt}}{{end}}",
			},
			"codex-base": {
				File: "~/projects/cc-plugin/pkg/system-prompts/CL4R1T4S/OPENAI/Codex.md",
			},
		},
	}
}

func validateSettings(settings Settings) error {
	if strings.TrimSpace(settings.DefaultClient) == "" {
		return fmt.Errorf("default_client is required")
	}
	if len(settings.Clients) == 0 {
		return fmt.Errorf("at least one client is required")
	}
	for id, client := range settings.Clients {
		if !configIDPattern.MatchString(id) {
			return fmt.Errorf("client %q must use kebab-case", id)
		}
		if client.Disabled {
			continue
		}
		if err := validateClient(id, client); err != nil {
			return err
		}
	}
	for id, promptTemplate := range settings.Templates {
		if !configIDPattern.MatchString(id) {
			return fmt.Errorf("template %q must use kebab-case", id)
		}
		hasContent := promptTemplate.Content != ""
		hasFile := strings.TrimSpace(promptTemplate.File) != ""
		if hasContent == hasFile {
			return fmt.Errorf("template %q must set exactly one of content or file", id)
		}
	}
	if _, _, err := resolveClient(settings, settings.DefaultClient); err != nil {
		return fmt.Errorf("default_client: %w", err)
	}
	return nil
}

func validateClient(id string, client ClientConfig) error {
	switch client.Driver {
	case "agy", "claude", "codex":
	default:
		return fmt.Errorf("client %q has unsupported driver %q", id, client.Driver)
	}
	if strings.TrimSpace(client.Command) == "" {
		return fmt.Errorf("client %q command is required", id)
	}
	if strings.TrimSpace(client.Model) == "" {
		return fmt.Errorf("client %q model is required", id)
	}
	if strings.TrimSpace(client.Effort) == "" {
		return fmt.Errorf("client %q effort is required", id)
	}
	expectedTransport := map[string]string{
		"agy":    "argument",
		"claude": "argument",
		"codex":  "stdin",
	}[client.Driver]
	if client.PromptTransport != expectedTransport {
		return fmt.Errorf(
			"client %q prompt_transport must be %q for driver %q",
			id,
			expectedTransport,
			client.Driver,
		)
	}
	if client.Driver == "claude" && strings.TrimSpace(client.Settings) == "" {
		return fmt.Errorf("client %q settings is required for claude driver", id)
	}
	if err := validateCredential(id, client.Credential); err != nil {
		return err
	}
	if client.Driver == "agy" {
		allowedModels := client.Models
		if len(allowedModels) == 0 {
			allowedModels = defaultAgyModels
		}
		if !contains(allowedModels, client.Model) {
			return fmt.Errorf("client %q has unsupported agy model %q", id, client.Model)
		}
		if !contains([]string{"low", "medium", "high"}, client.Effort) {
			return fmt.Errorf("client %q has unsupported agy effort %q", id, client.Effort)
		}
	}
	if client.Driver == "claude" &&
		!contains([]string{"low", "medium", "high", "xhigh", "max"}, client.Effort) {
		return fmt.Errorf("client %q has unsupported claude effort %q", id, client.Effort)
	}
	return nil
}

func validateCredential(id string, credential CredentialConfig) error {
	switch credential.Mode {
	case "oauth":
		return nil
	case "env":
		if strings.TrimSpace(credential.SourceEnv) == "" {
			return fmt.Errorf("client %q credential source_env is required", id)
		}
		if strings.TrimSpace(credential.TargetEnv) == "" {
			return fmt.Errorf("client %q credential target_env is required", id)
		}
		return nil
	default:
		return fmt.Errorf("client %q has unsupported credential mode %q", id, credential.Mode)
	}
}

func resolveClient(settings Settings, requested string) (string, ClientConfig, error) {
	clientID := requested
	if clientID == "" {
		clientID = settings.DefaultClient
	}
	client, ok := settings.Clients[clientID]
	if !ok {
		return "", ClientConfig{}, fmt.Errorf("unknown client %q", clientID)
	}
	if client.Disabled {
		return "", ClientConfig{}, fmt.Errorf("client %q is disabled", clientID)
	}
	return clientID, client, nil
}

func applyClientOverrides(
	clientID string,
	client ClientConfig,
	model string,
	effort string,
	bypassPermission *bool,
) (ClientConfig, error) {
	if model != "" {
		if !contains(clientModelChoices(client), model) {
			return ClientConfig{}, fmt.Errorf("client %q does not support model %q", clientID, model)
		}
		client.Model = model
	}
	if effort != "" {
		if !contains(clientEffortChoices(client), effort) {
			return ClientConfig{}, fmt.Errorf("client %q does not support effort %q", clientID, effort)
		}
		client.Effort = effort
	}
	if bypassPermission != nil {
		client.AutoApprove = *bypassPermission
	}
	if err := validateClient(clientID, client); err != nil {
		return ClientConfig{}, err
	}
	return client, nil
}

func clientModelChoices(client ClientConfig) []string {
	configured := client.Models
	if len(configured) == 0 && client.Driver == "agy" {
		configured = defaultAgyModels
	}
	return prependUnique(client.Model, configured)
}

func clientEffortChoices(client ClientConfig) []string {
	configured := client.Efforts
	if len(configured) == 0 {
		switch client.Driver {
		case "agy":
			configured = []string{"high", "medium", "low"}
		case "claude":
			configured = []string{"max", "xhigh", "high", "medium", "low"}
		}
	}
	return orderEffortChoices(prependUnique(client.Effort, configured))
}

func orderEffortChoices(choices []string) []string {
	ordered := make([]string, 0, len(choices))
	for _, effort := range []string{"max", "xhigh", "high", "medium", "low"} {
		if contains(choices, effort) {
			ordered = append(ordered, effort)
		}
	}
	for _, effort := range choices {
		if !contains(ordered, effort) {
			ordered = append(ordered, effort)
		}
	}
	return ordered
}

func prependUnique(current string, configured []string) []string {
	choices := make([]string, 0, len(configured)+1)
	if current != "" {
		choices = append(choices, current)
	}
	for _, candidate := range configured {
		if candidate != "" && !contains(choices, candidate) {
			choices = append(choices, candidate)
		}
	}
	return choices
}

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
