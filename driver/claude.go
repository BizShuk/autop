package driver

import gosdkconfig "github.com/bizshuk/gosdk/config"

func prepareClaudeProcess(client ClientConfig, prompt string, workDir string) Process {
	args := make([]string, 0, 14)
	if client.AutoApprove {
		args = append(args, "--dangerously-skip-permissions")
	}
	args = append(
		args,
		"--settings",
		gosdkconfig.ExpandHome(client.Settings),
		"--model",
		client.Model,
		"--effort",
		client.Effort,
		"--add-dir",
		workDir,
		"--output-format",
		"text",
		"-p",
		prompt,
	)
	return Process{
		Name: client.Command,
		Args: args,
	}
}
