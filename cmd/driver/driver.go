// Package driver maps Autop client profiles to executable process invocations.
package driver

import (
	"errors"
	"fmt"
	"os"
	"strings"

	gosdkconfig "github.com/bizshuk/gosdk/config"
)

// ClientConfig describes one selectable local LLM CLI profile.
type ClientConfig struct {
	Disabled        bool             `json:"disabled" mapstructure:"disabled"`
	Driver          string           `json:"driver" mapstructure:"driver"`
	Command         string           `json:"command" mapstructure:"command"`
	AutoApprove     bool             `json:"auto_approve" mapstructure:"auto_approve"`
	Model           string           `json:"model" mapstructure:"model"`
	Models          []string         `json:"models,omitempty" mapstructure:"models"`
	Effort          string           `json:"effort" mapstructure:"effort"`
	Efforts         []string         `json:"efforts,omitempty" mapstructure:"efforts"`
	Settings        string           `json:"settings,omitempty" mapstructure:"settings"`
	PromptTransport string           `json:"prompt_transport" mapstructure:"prompt_transport"`
	Credential      CredentialConfig `json:"credential" mapstructure:"credential"`
}

// CredentialConfig describes how an existing CLI obtains authentication.
type CredentialConfig struct {
	Mode      string `json:"mode" mapstructure:"mode"`
	SourceEnv string `json:"source_env,omitempty" mapstructure:"source_env"`
	TargetEnv string `json:"target_env,omitempty" mapstructure:"target_env"`
}

// Process contains an executable invocation without shell parsing.
type Process struct {
	Name     string
	Args     []string
	Stdin    string
	ExtraEnv []string
	// EnvPreview is the shell-safe environment prefix used when the command is
	// displayed or logged. It references the source variable instead of the
	// resolved secret, so a rendered command line stays copy-pasteable without
	// ever exposing a credential value.
	EnvPreview []string
}

// LookupEnv reads one environment variable without exposing the whole environment.
type LookupEnv func(string) (string, bool)

// Prepare maps one client profile and prompt to an executable process.
func Prepare(
	clientID string,
	client ClientConfig,
	prompt string,
	workDir string,
	lookupEnv LookupEnv,
) (Process, error) {
	if err := checkSettingsFile(clientID, client.Settings); err != nil {
		return Process{}, err
	}
	extraEnv, err := prepareCredentialEnv(clientID, client.Credential, lookupEnv)
	if err != nil {
		return Process{}, err
	}

	process, err := BuildProcess(clientID, client, prompt, workDir)
	if err != nil {
		return Process{}, err
	}
	process.ExtraEnv = extraEnv
	return process, nil
}

// BuildProcess maps a client profile to an executable process without resolving credentials.
func BuildProcess(
	clientID string,
	client ClientConfig,
	prompt string,
	workDir string,
) (Process, error) {
	var process Process
	switch client.Driver {
	case "agy":
		process = prepareAgyProcess(client, prompt, workDir)
	case "claude":
		process = prepareClaudeProcess(client, prompt, workDir)
	case "codex":
		process = prepareCodexProcess(client, prompt, workDir)
	case "grok":
		process = prepareGrokProcess(client, prompt, workDir)
	default:
		return Process{}, fmt.Errorf(
			"client %q has unsupported driver %q",
			clientID,
			client.Driver,
		)
	}
	process.EnvPreview = credentialEnvPreview(client.Credential)
	return process, nil
}

// checkSettingsFile fails before execution when a profile names a settings file
// that is not readable. Profiles whose driver takes no settings file leave the
// field empty and are skipped.
func checkSettingsFile(clientID string, settings string) error {
	if strings.TrimSpace(settings) == "" {
		return nil
	}
	path := gosdkconfig.ExpandHome(settings)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("client %q settings file %s does not exist", clientID, path)
		}
		return fmt.Errorf("client %q settings file %s: %w", clientID, path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("client %q settings path %s is a directory", clientID, path)
	}
	return nil
}

func credentialEnvPreview(credential CredentialConfig) []string {
	if credential.Mode == "oauth" ||
		credential.SourceEnv == "" ||
		credential.TargetEnv == "" {
		return nil
	}
	return []string{credential.TargetEnv + `="$` + credential.SourceEnv + `"`}
}

func prepareCredentialEnv(
	clientID string,
	credential CredentialConfig,
	lookupEnv LookupEnv,
) ([]string, error) {
	if credential.Mode == "oauth" {
		return nil, nil
	}
	value, ok := lookupEnv(credential.SourceEnv)
	if !ok || value == "" {
		return nil, fmt.Errorf(
			"client %q requires environment variable %s",
			clientID,
			credential.SourceEnv,
		)
	}
	return []string{credential.TargetEnv + "=" + value}, nil
}
