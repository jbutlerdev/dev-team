package settings

type Settings struct {
	Model       string            `json:"model"`
	Provider    string            `json:"provider"`
	APIKey      string            `json:"apiKey"`
	Server      string            `json:"server"`
	GitHubToken string            `json:"githubToken"`
	McpServers  map[string]McpServer `json:"mcpServers"`
}

type McpServer struct {
    Path string `json:"path"`
    Args []string `json:"args"`
}
