package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ducksoba/hostcontrol-mcp/mcp/hostcontrol"
	"github.com/ducksoba/hostcontrol-mcp/mcp/hostcontrol/tools"

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
		httpSrv := &http.Server{Addr: *listen, Handler: httpServer}

		go func() {
			fmt.Fprintf(os.Stderr, "hostcontrol-mcp server running on http://%s\n", *listen)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()
		httpSrv.Shutdown(shutdownCtx)
	default:
		stdioServer := server.NewStdioServer(srv.MCPServer())
		fmt.Fprintln(os.Stderr, "hostcontrol-mcp server running on stdio")
		if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}
