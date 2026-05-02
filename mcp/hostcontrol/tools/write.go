package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func WriteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	content, ok := args["content"].(string)
	if !ok {
		return mcp.NewToolResultError("content is required"), nil
	}

	append := false
	if v, ok := args["append"].(bool); ok {
		append = v
	}

	var flags int
	if append {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open file: %v", err)), nil
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Wrote %d bytes to %s", len(content), path)), nil
}
