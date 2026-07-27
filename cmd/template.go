package autop

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	gosdkconfig "github.com/bizshuk/gosdk/config"
)

type promptTemplateData struct {
	Prompt  string
	WorkDir string
	Client  string
	Driver  string
}

func renderPrompt(
	settings Settings,
	templateID string,
	rawPrompt string,
	workDir string,
	clientID string,
) (string, error) {
	if templateID == "" {
		if strings.TrimSpace(rawPrompt) == "" {
			return "", fmt.Errorf("prompt is required when --template is not set")
		}
		return rawPrompt, nil
	}

	promptTemplate, ok := settings.Templates[templateID]
	if !ok {
		return "", fmt.Errorf("unknown template %q", templateID)
	}
	source := promptTemplate.Content
	if promptTemplate.File != "" {
		templatePath := gosdkconfig.ExpandHome(promptTemplate.File)
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return "", fmt.Errorf("read template %q from %q: %w", templateID, templatePath, err)
		}
		source = string(content)
	}

	parsed, err := template.New(templateID).Option("missingkey=error").Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", templateID, err)
	}
	var rendered bytes.Buffer
	driver := ""
	if client, ok := settings.Clients[clientID]; ok {
		driver = client.Driver
	}
	data := promptTemplateData{
		Prompt:  rawPrompt,
		WorkDir: workDir,
		Client:  clientID,
		Driver:  driver,
	}
	if err := parsed.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("render template %q: %w", templateID, err)
	}
	if strings.TrimSpace(rendered.String()) == "" {
		return "", fmt.Errorf("template %q rendered an empty prompt", templateID)
	}
	return rendered.String(), nil
}
