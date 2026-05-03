package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func FindHandler(ctx context.Context, req mcp.CallToolRequest, cfg *Config) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	if allowed, reason := cfg.CheckPath(path); !allowed {
		return mcp.NewToolResultError("access denied: " + reason), nil
	}

	name := ""
	if v, ok := args["name"].(string); ok {
		name = v
	}

	fileType := ""
	if v, ok := args["type"].(string); ok {
		fileType = v
	}

	maxDepth := 0
	if v, ok := args["max_depth"].(float64); ok {
		maxDepth = int(v)
	}

	var results []string
	baseDepth := strings.Count(filepath.Clean(path), string(os.PathSeparator))

	err := filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !cfg.CheckPathForWalk(filePath) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if maxDepth > 0 {
			currentDepth := strings.Count(filepath.Clean(filePath), string(os.PathSeparator)) - baseDepth
			if currentDepth > maxDepth {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if name != "" {
			matched, err := filepath.Match(name, fi.Name())
			if err != nil || !matched {
				return nil
			}
		}

		switch fileType {
		case "file":
			if !fi.Mode().IsRegular() {
				return nil
			}
		case "dir":
			if !fi.IsDir() {
				return nil
			}
		}

		results = append(results, filePath)
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk path: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No matches found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d matches:\n%s", len(results), joinResults(results))), nil
}
