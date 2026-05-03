package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hostcontrol-mcp/mcp/hostcontrol"
	"hostcontrol-mcp/mcp/hostcontrol/tools"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio or http")
	listen := flag.String("listen", "127.0.0.1:3000", "Listen address for HTTP mode")
	configPath := flag.String("config", "", "Path to access control config file")
	flag.Parse()

	srv := hostcontrol.NewServer("hostcontrol-mcp", "0.1.0")

	if *configPath != "" {
		cfg, err := tools.LoadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
		srv.SetConfig(cfg)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv.SetContext(ctx)

	if err := srv.RegisterTools(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register tools: %v\n", err)
		os.Exit(1)
	}

	switch *transport {
	case "http":
		httpServer := server.NewStreamableHTTPServer(srv.MCPServer())
		fmt.Fprintf(os.Stderr, "hostcontrol-mcp server running on http://%s\n", *listen)
		if err := httpServer.Start(*listen); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	default:
		stdioServer := server.NewStdioServer(srv.MCPServer())
		fmt.Fprintln(os.Stderr, "hostcontrol-mcp server running on stdio")
		if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}
