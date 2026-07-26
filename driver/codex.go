package driver

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
		Stdin: prompt,
	}
}
