# 百度 OCR 配置指南 (v3.0.0)

Legal Extractor v3.0 已全面迁移至百度 AI 引擎。为了处理 PDF 文档和图片扫描件，您需要配置百度云 Access Token。

> **注意**：处理 .docx (Word) 原生文件不需要此配置。

## 步骤 1：获取百度 Access Token

1. 登录 [百度智能云控制台](https://console.bce.baidu.com/)。
2. 在搜索栏输入 **"文字识别"** 并进入。
3. 创建应用并获取 `API Key` 和 `Secret Key`。
4. 按照百度官方文档获取 `Access Token`。

## 步骤 2：配置软件

### 方式一：修改配置文件 (推荐)

在项目目录下找到或创建 `internal/config/baked_conf.yaml`：

```yaml
baidu:
  token: "您的百度AccessToken"
```

### 方式二：环境变量 (Docker/开发环境)

设置环境变量：

```bash
export BAIDU_TOKEN="您的AccessToken"
```

## 常见问题

**Q: 为什么提示 "OCR 失败"？**
A: 请确认您的 Token 是否过期（通常有效期为 30 天），或者确认对应的 API（通用文字识别/高精度版）是否已在控制台领用免费额度。

**Q: 是否支持超长 PDF？**
A: v3.0 引入了物理切片技术，会自动将超长文档分割后分次调用 API，彻底解决单次请求体积限制。
