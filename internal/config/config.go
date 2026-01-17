package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	MCP MCPConfig `mapstructure:"mcp"`
}

// MCPConfig MCP OCR 服务配置
type MCPConfig struct {
	Bin  string   `mapstructure:"bin"`
	Args []string `mapstructure:"args"`
}

var (
	// 全局配置实例
	cfg *Config
	v   *viper.Viper
)

// Init 初始化配置系统
func Init(configPath string) error {
	v = viper.New()

	// 设置默认值
	v.SetDefault("mcp.bin", "")
	v.SetDefault("mcp.args", []string{})

	// 绑定环境变量 (前缀 LEGAL_EXTRACTOR_)
	// 例如: LEGAL_EXTRACTOR_MCP_BIN=npx
	v.SetEnvPrefix("LEGAL_EXTRACTOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 配置文件设置
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认查找路径
		v.SetConfigName("conf")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，尝试创建默认配置
			if err := ensureConfigFile(configPath); err != nil {
				return fmt.Errorf("创建默认配置失败: %w", err)
			}
			// 重新读取
			_ = v.ReadInConfig()
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析到结构体
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

// ensureConfigFile 确保配置文件存在，不存在则创建默认配置
func ensureConfigFile(configPath string) error {
	if configPath == "" {
		configPath = "config/conf.yaml"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入默认配置
	defaultConfig := `# Legal Extractor 配置文件
# 支持通过环境变量覆盖，前缀为 LEGAL_EXTRACTOR_
# 例如: LEGAL_EXTRACTOR_MCP_BIN=npx

mcp:
  bin: ""    # MCP 服务启动命令，例如 "npx"
  args: []   # 命令参数，例如 ["-y", "@modelcontextprotocol/server-ocr"]
`
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

// Get 获取当前配置
func Get() *Config {
	if cfg == nil {
		return &Config{}
	}
	return cfg
}

// GetMCP 获取 MCP 配置
func GetMCP() MCPConfig {
	if cfg == nil {
		return MCPConfig{}
	}
	return cfg.MCP
}

// LoadConfig 兼容旧 API，内部调用 Init
func LoadConfig(path string) (*Config, error) {
	if err := Init(path); err != nil {
		return nil, err
	}
	return Get(), nil
}
