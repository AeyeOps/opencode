package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/permission"
	"github.com/opencode-ai/opencode/internal/version"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpTool struct {
	mcpName     string
	tool        mcp.Tool
	mcpConfig   config.MCPServer
	permissions permission.Service
}

type MCPClient interface {
	Initialize(
		ctx context.Context,
		request mcp.InitializeRequest,
	) (*mcp.InitializeResult, error)
	ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	Close() error
}

func (b *mcpTool) Info() tools.ToolInfo {
	required := b.tool.InputSchema.Required
	if required == nil {
		required = make([]string, 0)
	}
	return tools.ToolInfo{
		Name:        fmt.Sprintf("%s_%s", b.mcpName, b.tool.Name),
		Description: b.tool.Description,
		Parameters:  b.tool.InputSchema.Properties,
		Required:    required,
	}
}

func runTool(ctx context.Context, c MCPClient, toolName string, input string) (tools.ToolResponse, error) {
	defer c.Close()
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "OpenCode",
		Version: version.Version,
	}

	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		return tools.NewTextErrorResponse(err.Error()), nil
	}

	toolRequest := mcp.CallToolRequest{}
	toolRequest.Params.Name = toolName
	var args map[string]any
	if err = json.Unmarshal([]byte(input), &args); err != nil {
		return tools.NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
	}
	toolRequest.Params.Arguments = args
	result, err := c.CallTool(ctx, toolRequest)
	if err != nil {
		return tools.NewTextErrorResponse(err.Error()), nil
	}

	output := ""
	for _, v := range result.Content {
		if v, ok := v.(mcp.TextContent); ok {
			output = v.Text
		} else {
			output = fmt.Sprintf("%v", v)
		}
	}

	return tools.NewTextResponse(output), nil
}

func (b *mcpTool) Run(ctx context.Context, params tools.ToolCall) (tools.ToolResponse, error) {
	sessionID, messageID := tools.GetContextValues(ctx)
	if sessionID == "" || messageID == "" {
		return tools.ToolResponse{}, fmt.Errorf("session ID and message ID are required for creating a new file")
	}
	permissionDescription := fmt.Sprintf("execute %s with the following parameters: %s", b.Info().Name, params.Input)
	p := b.permissions.Request(
		permission.CreatePermissionRequest{
			SessionID:   sessionID,
			Path:        config.WorkingDirectory(),
			ToolName:    b.Info().Name,
			Action:      "execute",
			Description: permissionDescription,
			Params:      params.Input,
		},
	)
	if !p {
		return tools.NewTextErrorResponse("permission denied"), nil
	}

	switch b.mcpConfig.Type {
	case config.MCPStdio:
		c, err := client.NewStdioMCPClient(
			b.mcpConfig.Command,
			b.mcpConfig.Env,
			b.mcpConfig.Args...,
		)
		if err != nil {
			return tools.NewTextErrorResponse(err.Error()), nil
		}
		return runTool(ctx, c, b.tool.Name, params.Input)
	case config.MCPSse:
		c, err := client.NewSSEMCPClient(
			b.mcpConfig.URL,
			client.WithHeaders(b.mcpConfig.Headers),
		)
		if err != nil {
			return tools.NewTextErrorResponse(err.Error()), nil
		}
		return runTool(ctx, c, b.tool.Name, params.Input)
	}

	return tools.NewTextErrorResponse("invalid mcp type"), nil
}

func NewMcpTool(name string, tool mcp.Tool, permissions permission.Service, mcpConfig config.MCPServer) tools.BaseTool {
	return &mcpTool{
		mcpName:     name,
		tool:        tool,
		mcpConfig:   mcpConfig,
		permissions: permissions,
	}
}

var mcpTools []tools.BaseTool

func getTools(ctx context.Context, name string, m config.MCPServer, permissions permission.Service, c MCPClient) []tools.BaseTool {
	logging.Debug("getTools: Starting for MCP server", "name", name)
	var stdioTools []tools.BaseTool
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "OpenCode",
		Version: version.Version,
	}

	logging.Debug("getTools: Initializing MCP client", "name", name, "protocolVersion", initRequest.Params.ProtocolVersion)
	logging.Debug("getTools: Sending initialize request", "name", name, "clientInfo", initRequest.Params.ClientInfo)

	// Add timeout for initialization
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	initResult, err := c.Initialize(initCtx, initRequest)
	if err != nil {
		logging.Error("error initializing mcp client", "name", name, "error", err, "contextErr", initCtx.Err())
		return stdioTools
	}
	logging.Debug("getTools: MCP client initialized successfully", "name", name, "serverInfo", initResult.ServerInfo, "capabilities", initResult.Capabilities)

	toolsRequest := mcp.ListToolsRequest{}
	logging.Debug("getTools: Listing tools from MCP server", "name", name)
	tools, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		logging.Error("error listing tools", "name", name, "error", err)
		return stdioTools
	}
	logging.Debug("getTools: Retrieved tool list", "name", name, "count", len(tools.Tools))

	for _, t := range tools.Tools {
		logging.Debug("getTools: Adding tool", "name", name, "toolName", t.Name)
		stdioTools = append(stdioTools, NewMcpTool(name, t, permissions, m))
	}
	defer c.Close()
	logging.Debug("getTools: Completed", "name", name, "totalTools", len(stdioTools))
	return stdioTools
}

func GetMcpTools(ctx context.Context, permissions permission.Service) []tools.BaseTool {
	logging.Debug("GetMcpTools: Starting MCP tools initialization")
	if len(mcpTools) > 0 {
		logging.Debug("GetMcpTools: Returning cached MCP tools", "count", len(mcpTools))
		return mcpTools
	}

	mcpServers := config.Get().MCPServers
	logging.Debug("GetMcpTools: Found MCP servers", "count", len(mcpServers))

	for name, m := range mcpServers {
		logging.Debug("GetMcpTools: Processing MCP server", "name", name, "type", m.Type, "command", m.Command, "args", m.Args)
		switch m.Type {
		case config.MCPStdio:
			logging.Debug("GetMcpTools: Creating stdio MCP client", "name", name, "command", m.Command, "args", m.Args, "env", m.Env)

			// Log the full command that will be executed
			logging.Debug("GetMcpTools: Executing command", "name", name, "fullCommand", fmt.Sprintf("%s %v", m.Command, m.Args))

			c, err := client.NewStdioMCPClient(
				m.Command,
				m.Env,
				m.Args...,
			)
			if err != nil {
				logging.Error("error creating mcp client", "name", name, "error", err, "command", m.Command, "args", m.Args)
				continue
			}
			logging.Debug("GetMcpTools: Stdio client created successfully", "name", name)

			tools := getTools(ctx, name, m, permissions, c)
			logging.Debug("GetMcpTools: Retrieved tools from MCP server", "name", name, "toolCount", len(tools))
			mcpTools = append(mcpTools, tools...)
		case config.MCPSse:
			logging.Debug("GetMcpTools: Creating SSE MCP client", "name", name, "url", m.URL)
			c, err := client.NewSSEMCPClient(
				m.URL,
				client.WithHeaders(m.Headers),
			)
			if err != nil {
				logging.Error("error creating mcp client", "name", name, "error", err)
				continue
			}
			logging.Debug("GetMcpTools: SSE client created successfully", "name", name)
			tools := getTools(ctx, name, m, permissions, c)
			logging.Debug("GetMcpTools: Retrieved tools from MCP server", "name", name, "toolCount", len(tools))
			mcpTools = append(mcpTools, tools...)
		}
	}

	return mcpTools
}
