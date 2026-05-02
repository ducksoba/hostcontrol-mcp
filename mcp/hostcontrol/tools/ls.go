package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func LsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	long := false
	if v, ok := args["long"].(bool); ok {
		long = v
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read directory: %v", err)), nil
	}

	var output string
	for _, entry := range entries {
		if long {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			mode := info.Mode().String()
			size := info.Size()
			modTime := info.ModTime().Format(time.RFC3339)
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			output += fmt.Sprintf("%s %10d %s %s\n", mode, size, modTime, name)
		} else {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			output += name + "\n"
		}
	}

	absPath, _ := filepath.Abs(path)
	header := fmt.Sprintf("Contents of %s:\n", absPath)
	return mcp.NewToolResultText(header + output), nil
}
