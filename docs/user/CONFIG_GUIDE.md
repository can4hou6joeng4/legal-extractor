# 腾讯云 OCR API 配置指南

Legal Extractor 使用腾讯云的 SmartStructuralOCRV2（智能结构化识别 V2）技术来处理 PDF 文档和图片扫描件。为了使用此功能，您需要申请并配置腾讯云的 API 密钥。

> **注意**：处理 .docx (Word) 文件不需要此配置，可以直接使用。

## 步骤 1：注册/登录腾讯云

1. 访问 [腾讯云控制台](https://console.cloud.tencent.com/)。
2. 登录您的腾讯云账号（支持微信扫码登录）。

## 步骤 2：开通文字识别服务

1. 在控制台搜索 **"文字识别"** 或直接访问 [文字识别控制台](https://console.cloud.tencent.com/ocr/overview)。
2. 首次使用需点击 **"立即开通"**。
3. 开通后，系统会赠送一定的免费调用额度。

## 步骤 3：获取 API 密钥

1. 访问 [API 密钥管理](https://console.cloud.tencent.com/cam/capi)。
2. 点击 **"新建密钥"**（如果没有现有密钥）。
3. 您将获得以下两个关键信息：
   - **SecretId**：用于标识 API 调用者身份
   - **SecretKey**：用于加密签名字符串

> **安全提示**：请妥善保管您的 SecretKey，不要泄露给他人。

## 步骤 4：配置软件

### 方式一：修改配置文件 (桌面版推荐)

1. 打开软件所在的目录。
2. 进入 `config` 文件夹（如果不存在请手动创建）。
3. 创建或编辑 `conf.yaml` 文件，填入您的密钥：

```yaml
tencent:
  secret_id: "您的SecretId"
  secret_key: "您的SecretKey"
```

### 方式二：环境变量 (Docker/服务器版推荐)

如果您使用 Docker 部署，可以在 `docker-compose.yml` 中设置环境变量：

```yaml
environment:
  - LEGAL_EXTRACTOR_TENCENT_SECRET_ID=您的SecretId
  - LEGAL_EXTRACTOR_TENCENT_SECRET_KEY=您的SecretKey
```

或者在命令行中设置：

```bash
export LEGAL_EXTRACTOR_TENCENT_SECRET_ID="您的SecretId"
export LEGAL_EXTRACTOR_TENCENT_SECRET_KEY="您的SecretKey"
```

## 常见问题

**Q: 只有 PDF 文件需要配置吗？**
A: 是的。只有当您导入 PDF 或图片文件时，程序才会调用腾讯云 OCR。处理 .docx 文件使用本地解析，无需配置。

**Q: 为什么提示 "签名验证失败"？**
A: 请检查 SecretId 和 SecretKey 是否复制完整，注意不要包含多余的空格。

**Q: 为什么提示 "SecretId 不存在"？**
A: 请确认密钥是否已被禁用或删除，可在 [API 密钥管理](https://console.cloud.tencent.com/cam/capi) 页面查看状态。

**Q: 是否收费？**
A: 腾讯云文字识别服务提供一定的免费额度：
- 智能结构化识别：每月 1000 次免费调用
- 超出后按量计费，具体价格请参考 [腾讯云定价](https://cloud.tencent.com/document/product/866/17619)

**Q: 如何查看剩余调用次数？**
A: 访问 [文字识别控制台](https://console.cloud.tencent.com/ocr/overview)，可查看资源包使用情况和调用统计。

## 技术支持

如遇到问题，请通过以下方式获取帮助：
- GitHub Issues: https://github.com/can4hou6joeng4/legal-extractor/issues
- 腾讯云文档: https://cloud.tencent.com/document/product/866
