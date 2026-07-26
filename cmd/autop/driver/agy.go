package driver

func prepareAgyProcess(client ClientConfig, prompt string, workDir string) Process {
	args := []string{"--print"}
	if client.AutoApprove {
		args = append(args, "--dangerously-skip-permissions")
	}
	args = append(
		args,
		"--model="+client.Model,
		"--effort="+client.Effort,
		"--add-dir",
		workDir,
		prompt,
	)
	return Process{
		Name: client.Command,
		Args: args,
	}
}
