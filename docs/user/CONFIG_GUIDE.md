# 🔧 Legal Extractor 配置指南

LegalExtractor V2.0 采用百度智能云提供的 **PaddleOCR-VL** 视觉大模型作为核心解析引擎。为了使用本软件，您需要配置自己的百度 AI API Key。

不用担心，百度云提供每月 **1000 次免费额度**，对于个人和小型团队使用绰绰有余。

## 第一步：获取 API Key

1.  **注册/登录**：访问 [百度智能云控制台](https://console.bce.baidu.com/) 并登录。
2.  **进入文字识别控制台**：在左侧菜单或产品列表中找到 **“人工智能” -> “文字识别”**。
3.  **创建应用**：
    *   点击“创建应用”。
    *   应用名称任意填写（如 `LegalOCR`）。
    *   接口选择：确保勾选 **“智能文档分析 (PP-Structure)”** 或 **“通用文字识别”** 相关权限（通常默认已选）。
    *   提交创建。
4.  **获取密钥**：
    *   在应用列表中找到刚才创建的应用。
    *   复制 **`API Key`** 和 **`Secret Key`**。

## 第二步：配置软件

1.  运行一次 `legal-extractor`，软件会自动在同级目录下生成 `config` 文件夹。
2.  打开 `config/conf.yaml` 文件（推荐使用记事本或 VS Code）。
3.  填入您的密钥：

```yaml
baidu:
  app_id: "您的 AppID (可选)"
  api_key: "您的 API Key 粘贴在这里"
  secret_key: "您的 Secret Key 粘贴在这里"
```

4.  保存文件并重启软件。

## 常见问题

**Q: 为什么提示“Unsupported openapi method”？**
A: 请确保您填写的 Key 正确，并且该应用已在百度云控制台开通了对应权限。

**Q: 软件提示“网络错误”？**
A: 请检查您的网络连接，并确认防火墙没有拦截软件对 `aip.baidubce.com` 的访问。
