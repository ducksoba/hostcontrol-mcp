package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"hostcontrol-mcp/mcp/accesscontrol"

	"github.com/mark3labs/mcp-go/mcp"
)

const defaultBashTimeout = 30 * time.Second

func BashHandler(ctx context.Context, req mcp.CallToolRequest, policy *accesscontrol.Policy) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	if policy != nil {
		baseCmd := extractBaseCommand(command)
		checkArgs := map[string]interface{}{"command": command}
		allowed, reason := policy.CheckTool("bash", checkArgs)
		if !allowed {
			return mcp.NewToolResultError(fmt.Sprintf("access denied: %s", reason)), nil
		}
		_ = baseCmd
	}

	cwd := "."
	if v, ok := args["cwd"].(string); ok && v != "" {
		cwd = v
	}

	timeout := defaultBashTimeout
	if v, ok := args["timeout"].(float64); ok && v > 0 {
		timeout = time.Duration(v) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError(fmt.Sprintf("command timed out after %v", timeout)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Exit error: %v\n%s", err, string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

func extractBaseCommand(command string) string {
	command = strings.TrimSpace(command)
	parts := strings.Fields(command)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
