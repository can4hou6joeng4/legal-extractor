# 常见问题与故障排查 (FAQ & Troubleshooting)

本文档涵盖 LegalExtractor 使用过程中的常见问题及解决方案。

---

## 🔴 PDF 解析相关问题

### 问题：提示 "图片为空" 或 "parse document task failed"

**错误信息：**
```
百度解析任务执行失败: parse document task failed
错误代码: 216200 - 图片为空
```

**可能原因：**
1. PDF 文件已损坏或加密
2. 文件内容为空白页
3. API 密钥配置错误

**解决方案：**
1. 用其他 PDF 阅读器打开文件，确认内容正常显示
2. 检查文件是否设置了密码保护
3. 验证 API 密钥是否正确配置

---

### 问题：提示 "Open api daily request limit reached"

**错误信息：**
```
提交任务业务错误: [17] Open api daily request limit reached
```

**原因：** 百度 AI 免费版 API 调用次数已达到每日上限。

**解决方案：**
1. 等待次日额度重置（北京时间 0:00）
2. 升级百度 AI 套餐获取更多配额
3. 使用多个 API Key 轮换

---

### 问题：提示 "百度 API Key 或 Secret Key 未配置"

**解决方案：**

**桌面版用户：**
最新版本（v2.1.1+）已内置 API 密钥，无需手动配置。如果仍有问题，请下载最新版本。

**Web/Docker 用户：**
设置环境变量后重启容器：
```bash
export LEGAL_EXTRACTOR_BAIDU_API_KEY="your_key"
export LEGAL_EXTRACTOR_BAIDU_SECRET_KEY="your_secret"
docker-compose down && docker-compose up -d
```

---

## 🟡 文件上传问题

### 问题：上传文件后提示 "不支持的文件格式"

**支持的格式：**
- `.docx` - Word 文档
- `.pdf` - PDF 文档
- `.jpg` / `.jpeg` / `.png` - 图片

**不支持的格式：**
- `.doc` (旧版 Word) - 请另存为 `.docx`
- `.txt` - 请复制内容到 Word 后保存

---

### 问题：文件上传后长时间无响应

**可能原因：**
1. 文件过大（建议单文件 < 10MB）
2. 网络连接不稳定
3. 百度 API 服务繁忙

**解决方案：**
1. 压缩 PDF 文件大小后重试
2. 检查网络连接
3. 等待 1-2 分钟后重试

---

## 🟢 导出相关问题

### 问题：导出的 Excel 文件中文显示乱码

**解决方案：**
使用 Excel 打开时选择 **UTF-8** 编码：
1. 打开 Excel → 数据 → 从文本/CSV
2. 选择文件后，在编码下拉菜单选择 **65001: Unicode (UTF-8)**

或直接使用 `.xlsx` 格式导出（推荐）。

---

### 问题：提取结果中某些字段为空

**可能原因：**
1. 原文档中不包含该字段信息
2. 文档格式与预期模板不匹配
3. OCR 识别精度问题（扫描件）

**解决方案：**
1. 使用「预览」功能检查原始识别结果
2. 确保文档包含标准法律文书关键词（如"被告："、"诉讼请求"）
3. 提高扫描件清晰度后重试

---

## 🔧 开发者问题

### 问题：`wails dev` 启动失败

**常见错误与解决方案：**

| 错误 | 解决方案 |
|------|----------|
| `wails: command not found` | 运行 `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| `npm ERR! missing script: build` | 运行 `cd frontend && npm install` |
| `baked_conf.yaml not found` | 创建 `internal/config/baked_conf.yaml` 文件（可为空） |

---

### 问题：测试失败

运行测试命令：
```bash
go test ./internal/... -v
```

如果测试失败，请检查：
1. Go 版本是否为 1.24+
2. 是否已安装所有依赖（`go mod tidy`）

---

## 📞 获取帮助

如果以上方案未能解决您的问题：

1. **查看日志**：桌面版在控制台输出错误详情
2. **提交 Issue**：[GitHub Issues](https://github.com/can4hou6joeng4/legal-extractor/issues)
3. **提供信息**：
   - 操作系统版本
   - 应用版本号
   - 完整错误信息
   - 复现步骤
