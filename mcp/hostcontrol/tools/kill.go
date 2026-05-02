package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
)

func KillHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	pidFloat, ok := args["pid"].(float64)
	if !ok {
		return mcp.NewToolResultError("pid is required"), nil
	}
	pid := int(pidFloat)

	signalName := "TERM"
	if v, ok := args["signal"].(string); ok && v != "" {
		signalName = v
	}

	sig, err := parseSignal(signalName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid signal: %v", err)), nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to find process: %v", err)), nil
	}

	if err := process.Signal(sig); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to send signal: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Sent %s to process %d", signalName, pid)), nil
}

func parseSignal(name string) (os.Signal, error) {
	switch name {
	case "HUP", "SIGHUP":
		return syscall.SIGHUP, nil
	case "INT", "SIGINT":
		return syscall.SIGINT, nil
	case "KILL", "SIGKILL":
		return syscall.SIGKILL, nil
	case "TERM", "SIGTERM":
		return syscall.SIGTERM, nil
	case "USR1", "SIGUSR1":
		return syscall.SIGUSR1, nil
	case "USR2", "SIGUSR2":
		return syscall.SIGUSR2, nil
	default:
		if num, err := strconv.Atoi(name); err == nil {
			return syscall.Signal(num), nil
		}
		return syscall.SIGTERM, fmt.Errorf("unknown signal: %s", name)
	}
}
