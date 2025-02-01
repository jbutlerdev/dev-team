package handlers

import (
	"context"
	"dev-team/internal/settings"
	"dev-team/internal/state"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func CallMcpServer(serverName string, toolName string, args map[string]interface{}) (string, error) {
	state.State.Mu.RLock()
	defer state.State.Mu.RUnlock()

	mcpServer, ok := state.State.Settings.McpServers[serverName]
	if !ok {
		return "", fmt.Errorf("mcp server %s not found", serverName)
	}

	c, err := client.NewStdioMCPClient(
		mcpServer.Path,
		mcpServer.Args,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "dev-team-client",
		Version: "1.0.0",
	}

	initResult, err := c.Initialize(ctx, initRequest)
	if err != nil {
		return "", fmt.Errorf("failed to initialize: %w", err)
	}
	log.Printf("Initialized with server: %s %s", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	callToolRequest := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}
	callToolRequest.Params.Name = toolName
	callToolRequest.Params.Arguments = args


	result, err := c.CallTool(ctx, callToolRequest)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %w", err)
	}

	var output string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			output += textContent.Text + "\n"
		} else {
			jsonBytes, _ := json.MarshalIndent(content, "", "  ")
			output += string(jsonBytes) + "\n"
		}
	}

	return output, nil
}

func ListMcpServers() map[string]settings.McpServer{
	state.State.Mu.RLock()
	defer state.State.Mu.RUnlock()

	return state.State.Settings.McpServers
}

func AddMcpServer(name string, server settings.McpServer) error {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()
	if state.State.Settings.McpServers == nil {
		state.State.Settings.McpServers = make(map[string]settings.McpServer)
	}
	state.State.Settings.McpServers[name] = server
	err := state.SaveConfig()
	return err
}

func RemoveMcpServer(name string) error {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	delete(state.State.Settings.McpServers, name)
	err := state.SaveConfig()
	return err
}
