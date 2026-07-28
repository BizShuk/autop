package driver

import (
	"strings"
	"unicode"
)

func prepareCodexProcess(client ClientConfig, prompt string, workDir string) Process {
	args := []string{"exec"}
	if client.AutoApprove {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	args = append(
		args,
		"--model",
		client.Model,
		"-c",
		`model_reasoning_effort="`+client.Effort+`"`,
		"-C",
		workDir,
		"-",
	)
	return Process{
		Name:  client.Command,
		Args:  args,
		Stdin: normalizeCodexSkillPrompt(prompt),
	}
}

func normalizeCodexSkillPrompt(prompt string) string {
	trimmedPrompt := strings.TrimLeftFunc(prompt, unicode.IsSpace)
	if !strings.HasPrefix(trimmedPrompt, "/") {
		return prompt
	}

	skillEnd := strings.IndexFunc(trimmedPrompt, unicode.IsSpace)
	if skillEnd == -1 {
		skillEnd = len(trimmedPrompt)
	}
	if skillEnd == 1 || strings.Contains(trimmedPrompt[1:skillEnd], "/") {
		return prompt
	}

	leadingWhitespace := prompt[:len(prompt)-len(trimmedPrompt)]
	return leadingWhitespace + "$" + trimmedPrompt[1:]
}
