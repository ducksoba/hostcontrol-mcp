package tools

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func ReadHandler(ctx context.Context, req mcp.CallToolRequest, cfg *Config) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	if allowed, reason := cfg.CheckPath(path); !allowed {
		return mcp.NewToolResultError("access denied: " + reason), nil
	}

	startLine := 1
	if v, ok := args["start_line"].(float64); ok {
		startLine = int(v)
	}

	endLine := math.MaxInt
	if v, ok := args["end_line"].(float64); ok {
		endLine = int(v)
	}

	file, err := os.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open file: %v", err)), nil
	}
	defer file.Close()

	var output string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum < startLine {
			continue
		}
		if lineNum > endLine {
			break
		}
		output += fmt.Sprintf("%d: %s\n", lineNum, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error reading file: %v", err)), nil
	}

	return mcp.NewToolResultText(output), nil
}
