package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"strconv"
	"time"
)

var (
	// BuildTime is injected via -ldflags at build time (Unix timestamp)
	BuildTime string = ""
)

const TrialDurationDays = 7

// TrialStatus represents the current trial state
type TrialStatus struct {
	IsExpired bool          `json:"isExpired"`
	Remaining time.Duration `json:"remaining"` // Duration until expiry
	Days      int           `json:"days"`      // Remaining whole days
	Hours     int           `json:"hours"`     // Remaining hours (modulo days)
}

// GetTrialStatus calculates the remaining trial time
func GetTrialStatus() TrialStatus {
	if BuildTime == "" {
		// Dev mode or local run without injection: no trial limit
		return TrialStatus{IsExpired: false, Remaining: 999 * time.Hour}
	}

	bt, err := strconv.ParseInt(BuildTime, 10, 64)
	if err != nil {
		return TrialStatus{IsExpired: false}
	}

	expiryTime := time.Unix(bt, 0).AddDate(0, 0, TrialDurationDays)
	remaining := time.Until(expiryTime)

	if remaining <= 0 {
		return TrialStatus{IsExpired: true, Remaining: 0}
	}

	return TrialStatus{
		IsExpired: false,
		Remaining: remaining,
		Days:      int(remaining.Hours() / 24),
		Hours:     int(remaining.Hours()) % 24,
	}
}

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
	err = v.ReadInConfig()

	// 判断是否需要加载内置配置 (Baked Config)
	// 触发条件：
	// 1. 本地配置文件不存在 (ConfigFileNotFoundError)
	// 2. 或者本地文件存在，但 API Key 为空 (说明是初始模板)
	useBaked := false
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			useBaked = true
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		// 文件读取成功，检查是否为空配置
		if v.GetString("baidu.api_key") == "" {
			fmt.Println("[ℹ️ 提示] 本地配置未设置 API Key，尝试加载内置配置...")
			useBaked = true
		}
	}

	// 加载内置配置
	if useBaked && len(bakedConfig) > 0 {
		// 注意：如果之前ReadInConfig成功但内容为空，这里ReadConfig(baked)会覆盖它
		// 如果是混合场景（本地有AppID但没Key），建议使用MergeConfig，但为简单起见，
		// 这里假设BakedConfig是完整的备用方案。
		v.SetConfigType("yaml")
		if loadErr := v.MergeConfig(bytes.NewBuffer(bakedConfig)); loadErr != nil {
			fmt.Printf("[⚠️ 警告] 加载内置配置失败: %v\n", loadErr)
		} else {
			fmt.Println("[ℹ️ 提示] 已加载内置预设配置 (baked_conf.yaml)")
		}
	}

	// 如果最终 API Key 仍然为空，且之前是因为文件不存在才进来的，则创建默认模板
	if v.GetString("baidu.api_key") == "" {
		// 仅当本地文件确实不存在时才创建，避免覆盖用户已有的（虽然可能是空的）文件
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			defaultPath := filepath.Join(baseDir, "config", "conf.yaml")
			if createErr := ensureConfigFile(defaultPath); createErr != nil {
				return fmt.Errorf("创建默认配置失败: %w", createErr)
			}
			// 再次尝试读取刚创建的文件（虽然它是空的，但为了保持状态一致）
			v.SetConfigFile(defaultPath)
			_ = v.ReadInConfig()
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
