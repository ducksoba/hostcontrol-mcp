package hostcontrol

import (
	"context"

	"hostcontrol-mcp/mcp/accesscontrol"
	"hostcontrol-mcp/mcp/hostcontrol/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	policy    *accesscontrol.Policy
}

func NewServer(name, version string) *Server {
	return &Server{
		mcpServer: server.NewMCPServer(name, version),
	}
}

func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

func (s *Server) LoadPolicy(path string) error {
	policy, err := accesscontrol.LoadPolicyFromFile(path)
	if err != nil {
		return err
	}
	s.policy = policy
	return nil
}

func (s *Server) RegisterTools(ctx context.Context) error {
	s.registerReadTool()
	s.registerWriteTool()
	s.registerBashTool()
	s.registerGrepTool()
	s.registerLsTool()
	s.registerPsTool()
	s.registerKillTool()
	return nil
}

func (s *Server) registerReadTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("read",
			mcp.WithDescription("Read file contents"),
			mcp.WithString("path", mcp.Required(), mcp.Description("File path to read")),
			mcp.WithNumber("start_line", mcp.Description("Starting line number (1-indexed)")),
			mcp.WithNumber("end_line", mcp.Description("Ending line number (inclusive)")),
		),
		Handler: tools.ReadHandler,
	})
}

func (s *Server) registerWriteTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("write",
			mcp.WithDescription("Write or append to a file"),
			mcp.WithString("path", mcp.Required(), mcp.Description("File path to write to")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Content to write")),
			mcp.WithBoolean("append", mcp.Description("Append to file instead of overwriting")),
		),
		Handler: tools.WriteHandler,
	})
}

func (s *Server) registerBashTool() {
	policy := s.policy
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("bash",
			mcp.WithDescription("Execute a shell command"),
			mcp.WithString("command", mcp.Required(), mcp.Description("Shell command to execute")),
			mcp.WithString("cwd", mcp.Description("Working directory for the command")),
			mcp.WithNumber("timeout", mcp.Description("Timeout in seconds (default: 30)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return tools.BashHandler(ctx, req, policy)
		},
	})
}

func (s *Server) registerGrepTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("grep",
			mcp.WithDescription("Search files using regex pattern"),
			mcp.WithString("pattern", mcp.Required(), mcp.Description("Regex pattern to search for")),
			mcp.WithString("path", mcp.Required(), mcp.Description("File or directory to search in")),
			mcp.WithBoolean("recursive", mcp.Description("Search recursively in directories")),
		),
		Handler: tools.GrepHandler,
	})
}

func (s *Server) registerLsTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("ls",
			mcp.WithDescription("List directory contents"),
			mcp.WithString("path", mcp.Required(), mcp.Description("Directory path to list")),
			mcp.WithBoolean("long", mcp.Description("Show detailed listing")),
		),
		Handler: tools.LsHandler,
	})
}

func (s *Server) registerPsTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("ps",
			mcp.WithDescription("List running processes"),
			mcp.WithString("user", mcp.Description("Filter by user")),
			mcp.WithString("command", mcp.Description("Filter by command name")),
		),
		Handler: tools.PsHandler,
	})
}

func (s *Server) registerKillTool() {
	s.mcpServer.AddTools(server.ServerTool{
		Tool: mcp.NewTool("kill",
			mcp.WithDescription("Send signal to a process"),
			mcp.WithNumber("pid", mcp.Required(), mcp.Description("Process ID to signal")),
			mcp.WithString("signal", mcp.Description("Signal to send (default: TERM)")),
		),
		Handler: tools.KillHandler,
	})
}
