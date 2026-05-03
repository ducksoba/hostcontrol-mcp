package tools

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func HostnameHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return mcp.NewToolResultError("failed to get hostname: " + err.Error()), nil
	}

	return mcp.NewToolResultText(hostname), nil
}
