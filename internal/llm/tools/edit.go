package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kujtimiihoxha/termai/internal/config"
	"github.com/kujtimiihoxha/termai/internal/lsp"
	"github.com/kujtimiihoxha/termai/internal/permission"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type EditParams struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

type EditPermissionsParams struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
	Diff      string `json:"diff"`
}

type editTool struct {
	lspClients  map[string]*lsp.Client
	permissions permission.Service
}

const (
	EditToolName    = "edit"
	editDescription = `Edits files by replacing text, creating new files, or deleting content. For moving or renaming files, use the Bash tool with the 'mv' command instead. For larger file edits, use the FileWrite tool to overwrite files.

Before using this tool:

1. Use the FileRead tool to understand the file's contents and context

2. Verify the directory path is correct (only applicable when creating new files):
   - Use the LS tool to verify the parent directory exists and is the correct location

To make a file edit, provide the following:
1. file_path: The absolute path to the file to modify (must be absolute, not relative)
2. old_string: The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)
3. new_string: The edited text to replace the old_string

Special cases:
- To create a new file: provide file_path and new_string, leave old_string empty
- To delete content: provide file_path and old_string, leave new_string empty

The tool will replace ONE occurrence of old_string with new_string in the specified file.

CRITICAL REQUIREMENTS FOR USING THIS TOOL:

1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:
   - Include AT LEAST 3-5 lines of context BEFORE the change point
   - Include AT LEAST 3-5 lines of context AFTER the change point
   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file

2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:
   - Make separate calls to this tool for each instance
   - Each call must uniquely identify its specific instance using extensive context

3. VERIFICATION: Before using this tool:
   - Check how many instances of the target text exist in the file
   - If multiple instances exist, gather enough context to uniquely identify each one
   - Plan separate tool calls for each instance

WARNING: If you do not follow these requirements:
   - The tool will fail if old_string matches multiple locations
   - The tool will fail if old_string doesn't match exactly (including whitespace)
   - You may change the wrong instance if you don't include enough context

When making edits:
   - Ensure the edit results in idiomatic, correct code
   - Do not leave the code in a broken state
   - Always use absolute file paths (starting with /)

Remember: when making multiple file edits in a row to the same file, you should prefer to send all edits in a single message with multiple calls to this tool, rather than multiple messages with a single call each.`
)

func NewEditTool(lspClients map[string]*lsp.Client, permissions permission.Service) BaseTool {
	return &editTool{
		lspClients:  lspClients,
		permissions: permissions,
	}
}

func (e *editTool) Info() ToolInfo {
	return ToolInfo{
		Name:        EditToolName,
		Description: editDescription,
		Parameters: map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "The absolute path to the file to modify",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The text to replace",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The text to replace it with",
			},
		},
		Required: []string{"file_path", "old_string", "new_string"},
	}
}

func (e *editTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params EditParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse("invalid parameters"), nil
	}

	if params.FilePath == "" {
		return NewTextErrorResponse("file_path is required"), nil
	}

	if !filepath.IsAbs(params.FilePath) {
		wd := config.WorkingDirectory()
		params.FilePath = filepath.Join(wd, params.FilePath)
	}

	if params.OldString == "" {
		result, err := e.createNewFile(params.FilePath, params.NewString)
		if err != nil {
			return NewTextErrorResponse(fmt.Sprintf("error creating file: %s", err)), nil
		}
		return NewTextResponse(result), nil
	}

	if params.NewString == "" {
		result, err := e.deleteContent(params.FilePath, params.OldString)
		if err != nil {
			return NewTextErrorResponse(fmt.Sprintf("error deleting content: %s", err)), nil
		}
		return NewTextResponse(result), nil
	}

	result, err := e.replaceContent(params.FilePath, params.OldString, params.NewString)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("error replacing content: %s", err)), nil
	}

	waitForLspDiagnostics(ctx, params.FilePath, e.lspClients)
	result = fmt.Sprintf("<result>\n%s\n</result>\n", result)
	result += appendDiagnostics(params.FilePath, e.lspClients)
	return NewTextResponse(result), nil
}

func (e *editTool) createNewFile(filePath, content string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		if fileInfo.IsDir() {
			return "", fmt.Errorf("path is a directory, not a file: %s", filePath)
		}
		return "", fmt.Errorf("file already exists: %s. Use the Replace tool to overwrite an existing file", filePath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to access file: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create parent directories: %w", err)
	}

	p := e.permissions.Request(
		permission.CreatePermissionRequest{
			Path:        filepath.Dir(filePath),
			ToolName:    EditToolName,
			Action:      "create",
			Description: fmt.Sprintf("Create file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:  filePath,
				OldString: "",
				NewString: content,
				Diff:      GenerateDiff("", content),
			},
		},
	)
	if !p {
		return "", fmt.Errorf("permission denied")
	}

	err = os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)

	return "File created: " + filePath, nil
}

func (e *editTool) deleteContent(filePath, oldString string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", filePath)
		}
		return "", fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	if getLastReadTime(filePath).IsZero() {
		return "", fmt.Errorf("you must read the file before editing it. Use the View tool first")
	}

	modTime := fileInfo.ModTime()
	lastRead := getLastReadTime(filePath)
	if modTime.After(lastRead) {
		return "", fmt.Errorf("file %s has been modified since it was last read (mod time: %s, last read: %s)",
			filePath, modTime.Format(time.RFC3339), lastRead.Format(time.RFC3339))
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)

	index := strings.Index(oldContent, oldString)
	if index == -1 {
		return "", fmt.Errorf("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks")
	}

	lastIndex := strings.LastIndex(oldContent, oldString)
	if index != lastIndex {
		return "", fmt.Errorf("old_string appears multiple times in the file. Please provide more context to ensure a unique match")
	}

	newContent := oldContent[:index] + oldContent[index+len(oldString):]

	p := e.permissions.Request(
		permission.CreatePermissionRequest{
			Path:        filepath.Dir(filePath),
			ToolName:    EditToolName,
			Action:      "delete",
			Description: fmt.Sprintf("Delete content from file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:  filePath,
				OldString: oldString,
				NewString: "",
				Diff:      GenerateDiff(oldContent, newContent),
			},
		},
	)
	if !p {
		return "", fmt.Errorf("permission denied")
	}

	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)

	return "Content deleted from file: " + filePath, nil
}

func (e *editTool) replaceContent(filePath, oldString, newString string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", filePath)
		}
		return "", fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	if getLastReadTime(filePath).IsZero() {
		return "", fmt.Errorf("you must read the file before editing it. Use the View tool first")
	}

	modTime := fileInfo.ModTime()
	lastRead := getLastReadTime(filePath)
	if modTime.After(lastRead) {
		return "", fmt.Errorf("file %s has been modified since it was last read (mod time: %s, last read: %s)",
			filePath, modTime.Format(time.RFC3339), lastRead.Format(time.RFC3339))
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)

	index := strings.Index(oldContent, oldString)
	if index == -1 {
		return "", fmt.Errorf("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks")
	}

	lastIndex := strings.LastIndex(oldContent, oldString)
	if index != lastIndex {
		return "", fmt.Errorf("old_string appears multiple times in the file. Please provide more context to ensure a unique match")
	}

	newContent := oldContent[:index] + newString + oldContent[index+len(oldString):]

	startIndex := max(0, index-3)
	oldEndIndex := min(len(oldContent), index+len(oldString)+3)
	newEndIndex := min(len(newContent), index+len(newString)+3)

	diff := GenerateDiff(oldContent[startIndex:oldEndIndex], newContent[startIndex:newEndIndex])

	p := e.permissions.Request(
		permission.CreatePermissionRequest{
			Path:        filepath.Dir(filePath),
			ToolName:    EditToolName,
			Action:      "replace",
			Description: fmt.Sprintf("Replace content in file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:  filePath,
				OldString: oldString,
				NewString: newString,
				Diff:      diff,
			},
		},
	)
	if !p {
		return "", fmt.Errorf("permission denied")
	}

	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)

	return "Content replaced in file: " + filePath, nil
}

func GenerateDiff(oldContent, newContent string) string {
	dmp := diffmatchpatch.New()
	fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(oldContent, newContent)
	diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
	diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
	diffs = dmp.DiffCleanupSemantic(diffs)
	buff := strings.Builder{}

	buff.WriteString("Changes:\n")

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			for line := range strings.SplitSeq(text, "\n") {
				if line == "" {
					continue
				}
				_, _ = buff.WriteString("+ " + line + "\n")
			}
		case diffmatchpatch.DiffDelete:
			for line := range strings.SplitSeq(text, "\n") {
				if line == "" {
					continue
				}
				_, _ = buff.WriteString("- " + line + "\n")
			}
		case diffmatchpatch.DiffEqual:
			lines := strings.Split(text, "\n")
			if len(lines) > 3 {
				if lines[0] != "" {
					_, _ = buff.WriteString("  " + lines[0] + "\n")
				}
				_, _ = buff.WriteString("  ...\n")
				if lines[len(lines)-1] != "" {
					_, _ = buff.WriteString("  " + lines[len(lines)-1] + "\n")
				}
			} else {
				for _, line := range lines {
					if line == "" {
						continue
					}
					_, _ = buff.WriteString("  " + line + "\n")
				}
			}
		}
	}
	return buff.String()
}
