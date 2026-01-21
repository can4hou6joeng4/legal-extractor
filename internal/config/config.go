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
	Baidu BaiduConfig `mapstructure:"baidu"`
}

// BaiduConfig 百度 AI OCR 配置
type BaiduConfig struct {
	AppID     string `mapstructure:"app_id"`
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
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
	v.SetDefault("baidu.app_id", "")
	v.SetDefault("baidu.api_key", "")
	v.SetDefault("baidu.secret_key", "")

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
# 例如: LEGAL_EXTRACTOR_BAIDU_API_KEY=xxx

baidu:
  app_id: ""     # 百度 AI AppID
  api_key: ""    # 百度 AI API Key
  secret_key: "" # 百度 AI Secret Key
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

// GetBaidu 获取百度配置
func GetBaidu() BaiduConfig {
	if cfg == nil {
		return BaiduConfig{}
	}
	return cfg.Baidu
}

// LoadConfig 兼容旧 API，内部调用 Init
func LoadConfig(path string) (*Config, error) {
	if err := Init(path); err != nil {
		return nil, err
	}
	return Get(), nil
}
