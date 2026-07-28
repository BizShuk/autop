package autop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var appsArrayPattern = regexp.MustCompile(`\bapps\s*:\s*\[`)

type ecosystemTask struct {
	ClientID              string
	TemplateID            string
	Prompt                string
	WorkspaceDir          string
	BypassPermission      bool
	Model                 string
	Effort                string
	Cron                  string
	IncludeRuntimeOptions bool
}

func installEcosystem(path string, clientID string, templateID string) error {
	return installEcosystemTask(path, ecosystemTask{
		ClientID:   clientID,
		TemplateID: templateID,
	})
}

func installEcosystemTask(path string, task ecosystemTask) error {
	if !configIDPattern.MatchString(task.ClientID) {
		return fmt.Errorf("client %q must use kebab-case", task.ClientID)
	}
	if task.TemplateID != "" && !configIDPattern.MatchString(task.TemplateID) {
		return fmt.Errorf("template %q must use kebab-case", task.TemplateID)
	}
	if task.IncludeRuntimeOptions && (task.Model == "" || task.Effort == "") {
		return fmt.Errorf("model and effort are required for wizard-managed tasks")
	}
	if task.IncludeRuntimeOptions && task.TemplateID == "" && strings.TrimSpace(task.Prompt) == "" {
		return fmt.Errorf("prompt is required when no template is selected")
	}

	taskID := "autop-" + task.ClientID
	if task.TemplateID != "" {
		taskID += "-" + task.TemplateID
	}
	workDir := task.WorkspaceDir
	if workDir == "" {
		workDir = filepath.Dir(path)
	}
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve workspace directory %q: %w", workDir, err)
	}
	displayName := "Autop " + task.ClientID + " " + filepath.Base(workDir)
	block := renderManagedApp(taskID, displayName, task, workDir, "        ")

	content, mode, err := readEcosystem(path)
	if err != nil {
		return err
	}
	if content == nil {
		content = []byte(
			"module.exports = {\n" +
				"    apps: [\n" +
				block + "\n" +
				"    ]\n" +
				"};\n",
		)
		return writeFileAtomic(path, content, mode)
	}

	updated, replaced, err := replaceManagedApp(string(content), taskID, block)
	if err != nil {
		return err
	}
	if !replaced {
		updated, err = appendManagedApp(string(content), block)
		if err != nil {
			return fmt.Errorf("update %s: %w", path, err)
		}
	}
	if updated == string(content) {
		return nil
	}
	return writeFileAtomic(path, []byte(updated), mode)
}

func readEcosystem(path string) ([]byte, os.FileMode, error) {
	content, err := os.ReadFile(path)
	if err == nil {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil, 0, fmt.Errorf("stat %s: %w", path, statErr)
		}
		return content, info.Mode().Perm(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0o644, nil
	}
	return nil, 0, fmt.Errorf("read %s: %w", path, err)
}

func renderManagedApp(
	taskID string,
	displayName string,
	task ecosystemTask,
	workDir string,
	indent string,
) string {
	args := ecosystemCommandArgs(task)
	encodedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		encodedArgs = append(encodedArgs, strconv.Quote(arg))
	}

	lines := []string{
		indent + "// autop:begin " + taskID,
		indent + "{",
		indent + "    name: " + strconv.Quote(displayName) + ",",
		indent + `    script: "autop",`,
		indent + "    args: [" + strings.Join(encodedArgs, ", ") + "],",
		indent + "    cwd: " + strconv.Quote(workDir) + ",",
		indent + `    namespace: "autop",`,
		indent + "    instances: 1,",
		indent + "    optional: true,",
	}
	if task.Cron != "" {
		lines = append(lines, indent+"    cron: "+strconv.Quote(task.Cron)+",")
	}
	lines = append(lines,
		indent+"    autorestart: false,",
		indent+"    watch: false",
		indent+"},",
		indent+"// autop:end "+taskID,
	)
	return strings.Join(lines, "\n")
}

func autopCommandArgs(task ecosystemTask) []string {
	args := []string{"-c", task.ClientID}
	if task.TemplateID != "" {
		args = append(args, "-t", task.TemplateID)
	}
	if task.IncludeRuntimeOptions {
		args = append(
			args,
			"--bypass-permission="+strconv.FormatBool(task.BypassPermission),
			"--model",
			task.Model,
			"--effort",
			task.Effort,
		)
	}
	if prompt := strings.TrimSpace(task.Prompt); prompt != "" {
		args = append(args, "--", prompt)
	}
	return args
}

func ecosystemCommandArgs(task ecosystemTask) []string {
	args := autopCommandArgs(task)
	if strings.TrimSpace(task.Prompt) != "" {
		args[len(args)-1] = quoteCommandPart(args[len(args)-1])
	}
	return args
}

func replaceManagedApp(content string, taskName string, replacement string) (string, bool, error) {
	beginMarker := "// autop:begin " + taskName
	endMarker := "// autop:end " + taskName
	begin := strings.Index(content, beginMarker)
	if begin < 0 {
		if strings.Contains(content, endMarker) {
			return "", false, fmt.Errorf("managed task %q has an end marker without a begin marker", taskName)
		}
		return content, false, nil
	}
	end := strings.Index(content[begin+len(beginMarker):], endMarker)
	if end < 0 {
		return "", false, fmt.Errorf("managed task %q is missing its end marker", taskName)
	}
	end += begin + len(beginMarker) + len(endMarker)

	lineStart := strings.LastIndex(content[:begin], "\n") + 1
	lineEnd := end
	if nextNewline := strings.Index(content[end:], "\n"); nextNewline >= 0 {
		lineEnd = end + nextNewline + 1
	}
	replacementWithNewline := replacement
	if lineEnd > end {
		replacementWithNewline += "\n"
	}
	return content[:lineStart] + replacementWithNewline + content[lineEnd:], true, nil
}

func appendManagedApp(content string, block string) (string, error) {
	location := appsArrayPattern.FindStringIndex(content)
	if location == nil {
		return "", fmt.Errorf("apps array not found")
	}
	openBracket := location[1] - 1
	closeBracket, err := matchingArrayClose(content, openBracket)
	if err != nil {
		return "", err
	}

	body := content[openBracket+1 : closeBracket]
	bodyWithoutTrailingSpace := strings.TrimRightFunc(body, unicode.IsSpace)
	closeIndent := lineIndent(content, closeBracket)
	var newBody string
	if strings.TrimSpace(bodyWithoutTrailingSpace) == "" {
		newBody = "\n" + block + "\n" + closeIndent
	} else {
		lastCodeIndex, lastCode, ok := lastSignificantCode(bodyWithoutTrailingSpace)
		if !ok {
			return "", fmt.Errorf("apps array contains no JavaScript value")
		}
		if lastCode != ',' {
			bodyWithoutTrailingSpace = bodyWithoutTrailingSpace[:lastCodeIndex+1] +
				"," +
				bodyWithoutTrailingSpace[lastCodeIndex+1:]
		}
		newBody = bodyWithoutTrailingSpace + "\n" + block + "\n" + closeIndent
	}
	return content[:openBracket+1] + newBody + content[closeBracket:], nil
}

func lastSignificantCode(content string) (int, byte, bool) {
	const (
		stateNormal = iota
		stateSingleQuote
		stateDoubleQuote
		stateTemplateQuote
		stateLineComment
		stateBlockComment
	)
	state := stateNormal
	escaped := false
	lastIndex := -1

	for i := 0; i < len(content); i++ {
		current := content[i]
		next := byte(0)
		if i+1 < len(content) {
			next = content[i+1]
		}

		switch state {
		case stateSingleQuote, stateDoubleQuote, stateTemplateQuote:
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if (state == stateSingleQuote && current == '\'') ||
				(state == stateDoubleQuote && current == '"') ||
				(state == stateTemplateQuote && current == '`') {
				state = stateNormal
				lastIndex = i
			}
			continue
		case stateLineComment:
			if current == '\n' {
				state = stateNormal
			}
			continue
		case stateBlockComment:
			if current == '*' && next == '/' {
				state = stateNormal
				i++
			}
			continue
		}

		switch {
		case current == '/' && next == '/':
			state = stateLineComment
			i++
		case current == '/' && next == '*':
			state = stateBlockComment
			i++
		case current == '\'':
			state = stateSingleQuote
			lastIndex = i
		case current == '"':
			state = stateDoubleQuote
			lastIndex = i
		case current == '`':
			state = stateTemplateQuote
			lastIndex = i
		case !unicode.IsSpace(rune(current)):
			lastIndex = i
		}
	}
	if lastIndex < 0 {
		return 0, 0, false
	}
	return lastIndex, content[lastIndex], true
}

func matchingArrayClose(content string, openBracket int) (int, error) {
	const (
		stateNormal = iota
		stateSingleQuote
		stateDoubleQuote
		stateTemplateQuote
		stateLineComment
		stateBlockComment
	)
	state := stateNormal
	depth := 0
	escaped := false

	for i := openBracket; i < len(content); i++ {
		current := content[i]
		next := byte(0)
		if i+1 < len(content) {
			next = content[i+1]
		}

		switch state {
		case stateSingleQuote, stateDoubleQuote, stateTemplateQuote:
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if (state == stateSingleQuote && current == '\'') ||
				(state == stateDoubleQuote && current == '"') ||
				(state == stateTemplateQuote && current == '`') {
				state = stateNormal
			}
			continue
		case stateLineComment:
			if current == '\n' {
				state = stateNormal
			}
			continue
		case stateBlockComment:
			if current == '*' && next == '/' {
				state = stateNormal
				i++
			}
			continue
		}

		switch {
		case current == '\'':
			state = stateSingleQuote
		case current == '"':
			state = stateDoubleQuote
		case current == '`':
			state = stateTemplateQuote
		case current == '/' && next == '/':
			state = stateLineComment
			i++
		case current == '/' && next == '*':
			state = stateBlockComment
			i++
		case current == '[':
			depth++
		case current == ']':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("apps array is not closed")
}

func lineIndent(content string, offset int) string {
	lineStart := strings.LastIndex(content[:offset], "\n") + 1
	indent := content[lineStart:offset]
	if strings.TrimSpace(indent) == "" {
		return indent
	}
	return "    "
}

func writeFileAtomic(path string, content []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	tempFile, err := os.CreateTemp(directory, ".ecosystem.config.js.*")
	if err != nil {
		return fmt.Errorf("create temporary ecosystem config: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if err := tempFile.Chmod(mode); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("set temporary ecosystem config permissions: %w", err)
	}
	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write temporary ecosystem config: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("sync temporary ecosystem config: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary ecosystem config: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
