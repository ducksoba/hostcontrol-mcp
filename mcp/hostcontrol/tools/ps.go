package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func PsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	filterUser := ""
	if v, ok := args["user"].(string); ok {
		filterUser = v
	}

	filterCmd := ""
	if v, ok := args["command"].(string); ok {
		filterCmd = v
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read /proc: %v", err)), nil
	}

	var output string
	count := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		procPath := filepath.Join("/proc", entry.Name())

		stat, err := os.ReadFile(filepath.Join(procPath, "stat"))
		if err != nil {
			continue
		}

		status, err := os.ReadFile(filepath.Join(procPath, "status"))
		if err != nil {
			continue
		}

		uid := extractUID(string(status))
		if filterUser != "" && uid != filterUser {
			continue
		}

		cmdline, err := os.ReadFile(filepath.Join(procPath, "cmdline"))
		if err != nil {
			continue
		}
		cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
		cmdStr = strings.TrimSpace(cmdStr)

		if filterCmd != "" && !strings.Contains(cmdStr, filterCmd) {
			continue
		}

		statParts := strings.Fields(string(stat))
		if len(statParts) < 4 {
			continue
		}
		comm := strings.Trim(statParts[1], "()")
		state := statParts[2]

		output += fmt.Sprintf("%-8d %-5s %-6s %s\n", pid, state, uid, comm)
		count++
	}

	if count == 0 {
		return mcp.NewToolResultText("No processes found matching filters"), nil
	}

	header := fmt.Sprintf("%-8s %-5s %-6s %s\n", "PID", "STATE", "UID", "COMMAND")
	return mcp.NewToolResultText(header + strings.Repeat("-", 60) + "\n" + output), nil
}

func extractUID(status string) string {
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "?"
}
