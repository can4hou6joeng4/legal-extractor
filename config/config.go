package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MCP MCPConfig `yaml:"mcp"`
}

type MCPConfig struct {
	Bin  string   `yaml:"bin"`
	Args []string `yaml:"args"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
