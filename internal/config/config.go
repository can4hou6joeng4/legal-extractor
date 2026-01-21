package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

//go:embed baked_conf.yaml
var bakedConfig []byte

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

	// 1. 获取可执行文件所在目录，确保生产环境下路径正确
	exePath, err := os.Executable()
	var baseDir string
	if err == nil {
		baseDir = filepath.Dir(exePath)
	} else {
		baseDir = "." // 回退到当前目录
	}

	// 设置默认值
	v.SetDefault("baidu.app_id", "")
	v.SetDefault("baidu.api_key", "")
	v.SetDefault("baidu.secret_key", "")

	// 绑定环境变量 (前缀 LEGAL_EXTRACTOR_)
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
		v.AddConfigPath(filepath.Join(baseDir, "config")) // 锁定可执行文件同级的 config 目录
		v.AddConfigPath(baseDir)
	}

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 磁盘上不存在配置文件，尝试加载内置的“烘焙”配置
			if len(bakedConfig) > 0 {
				v.SetConfigType("yaml")
				if err := v.ReadConfig(bytes.NewBuffer(bakedConfig)); err != nil {
					fmt.Printf("[⚠️ 警告] 加载内置配置失败: %v\n", err)
				} else {
					fmt.Println("[ℹ️ 提示] 正在使用内置预设配置运行")
				}
			}

			// 如果内置配置也为空，或者加载失败，则尝试创建默认配置模板
			if v.GetString("baidu.api_key") == "" {
				defaultPath := filepath.Join(baseDir, "config", "conf.yaml")
				if err := ensureConfigFile(defaultPath); err != nil {
					return fmt.Errorf("创建默认配置失败: %w", err)
				}
				v.SetConfigFile(defaultPath)
				_ = v.ReadInConfig()
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析到结构体
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 2. 检查关键配置是否为空，给出明确指引
	if cfg.Baidu.APIKey == "" || cfg.Baidu.SecretKey == "" {
		exePath, _ := os.Executable()
		absConfigPath := filepath.Join(filepath.Dir(exePath), "config", "conf.yaml")
		fmt.Printf("\n[⚠️ 配置提示] 未检测到有效的百度 API 密钥。\n")
		fmt.Printf("请编辑配置文件: %s\n", absConfigPath)
		fmt.Printf("申请教程详见文档: https://github.com/can4hou6joeng4/legal-extractor/blob/main/docs/user/CONFIG_GUIDE.md\n\n")
	}

	return nil
}

// ensureConfigFile 确保配置文件存在，不存在则创建默认配置
func ensureConfigFile(configPath string) error {
	// 如果传入的是空或相对路径，尝试将其转换为基于可执行文件目录的绝对路径
	if !filepath.IsAbs(configPath) {
		exePath, _ := os.Executable()
		baseDir := filepath.Dir(exePath)
		if configPath == "" || configPath == "config/conf.yaml" {
			configPath = filepath.Join(baseDir, "config", "conf.yaml")
		} else {
			configPath = filepath.Join(baseDir, configPath)
		}
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
