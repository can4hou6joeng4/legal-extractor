package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClient handles communication with an MCP Server
type MCPClient struct {
	cli *client.Client
}

// NewMCPClient creates a new client with the given configuration
func NewMCPClient(bin string, args []string) (*MCPClient, error) {
	if bin == "" {
		return nil, fmt.Errorf("MCP binary path not specified")
	}

	cli, err := client.NewStdioMCPClient(bin, args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Initialize
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "legal-extractor-client",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initReq)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	return &MCPClient{cli: cli}, nil
}

// ExtractText calls the 'ocr' or 'read_image' tool on the server
func (c *MCPClient) ExtractText(imagePath string) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// List tools
	listResp, err := c.cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to list tools: %w", err)
	}

	var toolName string
	for _, t := range listResp.Tools {
		if t.Name == "ocr" || t.Name == "read_image" || t.Name == "extract_text" {
			toolName = t.Name
			break
		}
	}

	if toolName == "" {
		return "", fmt.Errorf("remote server does not offer an 'ocr', 'read_image', or 'extract_text' tool")
	}

	args := map[string]interface{}{
		"image": imagePath,
		"path":  imagePath,
	}

	res, err := c.cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}

	var fullText string
	for _, content := range res.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			fullText += textContent.Text + "\n"
		}
	}

	return fullText, nil
}

// Close cleans up the process
func (c *MCPClient) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}
