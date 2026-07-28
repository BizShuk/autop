package driver

func prepareGrokProcess(client ClientConfig, prompt string, workDir string) Process {
	args := make([]string, 0, 10)
	if client.AutoApprove {
		args = append(args, "--always-approve", "--permission-mode", "auto")
	}
	args = append(
		args,
		"--model",
		client.Model,
		"--reasoning-effort",
		client.Effort,
		"--cwd",
		workDir,
		"--single",
		prompt,
	)
	return Process{
		Name: client.Command,
		Args: args,
	}
}
