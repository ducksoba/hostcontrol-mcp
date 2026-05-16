package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const defaultBashTimeout = 30 * time.Second

func BashHandler(signalCtx context.Context, ctx context.Context, req mcp.CallToolRequest, cfg *Config) (*mcp.CallToolResult, error) {
	if cfg != nil && !cfg.AllowBash {
		return mcp.NewToolResultError("bash execution is not allowed"), nil
	}

	args := req.GetArguments()

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	if allowed, reason := cfg.CheckBashCommand(command); !allowed {
		return mcp.NewToolResultError("access denied: " + reason), nil
	}

	cwd := "."
	if v, ok := args["cwd"].(string); ok && v != "" {
		cwd = v
	}

	timeout := defaultBashTimeout
	if v, ok := args["timeout"].(float64); ok && v > 0 {
		timeout = time.Duration(cfg.CapTimeout(int(v))) * time.Second
	}

	execCtx, cancel := context.WithTimeout(signalCtx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError(fmt.Sprintf("command timed out after %v", timeout)), nil
		}
		if signalCtx.Err() != nil {
			return mcp.NewToolResultError("command cancelled: server shutting down"), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Exit error: %v\n%s", err, string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}
