package config

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// GetMachineID 获取当前设备的唯一识别短码
func GetMachineID() string {
	// 获取主机名作为简单标识（生产环境下建议结合 CPU/硬盘序列号）
	// 为了演示，我们先使用基础库，避免引入外部依赖
	hostname, _ := "LegalExtractor-User", error(nil)
	// 实际应用中可以获取更硬的标识
	hash := md5.Sum([]byte(hostname + "salt-for-legal"))
	return strings.ToUpper(fmt.Sprintf("%x", hash)[:8])
}

// VerifyLicense 校验授权码是否合法
// 规则：授权码 = MD5(MachineID + "SECRET_KEY") 的前 16 位，每 4 位加一个横杠
func VerifyLicense(machineID, licenseCode string) bool {
	expected := GenerateLicense(machineID)
	return strings.ToUpper(licenseCode) == expected
}

// GenerateLicense 生成授权码（供开发者使用）
func GenerateLicense(machineID string) string {
	raw := fmt.Sprintf("%x", md5.Sum([]byte(machineID + "legal-extractor-secret-2026")))
	code := strings.ToUpper(raw[:16])
	return fmt.Sprintf("%s-%s-%s-%s", code[0:4], code[4:8], code[8:12], code[12:16])
}

// IsActivated 检查是否已激活
func IsActivated() bool {
	if v == nil {
		return false
	}
	license := v.GetString("license_key")
	if license == "" {
		return false
	}
	return VerifyLicense(GetMachineID(), license)
}

// SaveLicense 保存授权码
func SaveLicense(code string) error {
	if v == nil {
		return fmt.Errorf("config system not initialized")
	}
	v.Set("license_key", code)
	return v.WriteConfig()
}
