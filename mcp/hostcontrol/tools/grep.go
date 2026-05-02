package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
)

func GrepHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return mcp.NewToolResultError("pattern is required"), nil
	}

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	recursive := false
	if v, ok := args["recursive"].(bool); ok {
		recursive = v
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid regex pattern: %v", err)), nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to access path: %v", err)), nil
	}

	var results []string

	if info.IsDir() {
		if recursive {
			err = filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if fi.IsDir() {
					return nil
				}
				matches, err := grepFile(filePath, re)
				if err != nil {
					return nil
				}
				results = append(results, matches...)
				return nil
			})
		} else {
			entries, err := os.ReadDir(path)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to read directory: %v", err)), nil
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				filePath := filepath.Join(path, entry.Name())
				matches, err := grepFile(filePath, re)
				if err != nil {
					continue
				}
				results = append(results, matches...)
			}
		}
	} else {
		matches, err := grepFile(path, re)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search file: %v", err)), nil
		}
		results = matches
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No matches found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d matches:\n%s", len(results), joinResults(results))), nil
}

func grepFile(path string, re *regexp.Regexp) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, fmt.Sprintf("%s:%d: %s", path, lineNum, line))
		}
	}
	return matches, scanner.Err()
}

func joinResults(results []string) string {
	out := ""
	for _, r := range results {
		out += r + "\n"
	}
	return out
}
