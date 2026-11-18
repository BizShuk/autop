package autop

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	autopdriver "github.com/bizshuk/autop/cmd/driver"
	"github.com/bizshuk/autop/config/example"
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
	gosdkconfig.Default(gosdkconfig.WithAppName(appName), gosdkconfig.WithDefaultValue(string(example.Settings)))

	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return Settings{}, fmt.Errorf("decode autop settings: %w", err)
	}
	if err := validateSettings(settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

// defaultSettings decodes config/example/settings.example.json, the single
// source of autop's defaults (also the WithDefaultValue layer and the
// `autop config default` seed).
func defaultSettings() Settings {
	var settings Settings
	if err := json.Unmarshal(example.Settings, &settings); err != nil {
		panic(fmt.Errorf("config/example/settings.example.json: %w", err))
	}
	return settings
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
	case "agy", "claude", "codex", "grok":
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
		"grok":   "argument",
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
