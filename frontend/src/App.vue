<script setup lang="ts">
import { ref, computed } from 'vue'
import { SelectFile, ExtractFromFile, PreviewData } from '../wailsjs/go/main/App'

interface Record {
  defendant: string
  idNumber: string
  request: string
  factsReason: string
}

interface ExtractResult {
  success: boolean
  recordCount: number
  outputPath: string
  errorMessage?: string
  records?: Record[]
}

const selectedFile = ref<string>('')
const fileName = computed(() => selectedFile.value ? selectedFile.value.split('/').pop() : '')
const isLoading = ref(false)
const result = ref<ExtractResult | null>(null)
const previewRecords = ref<Record[]>([])
const showPreview = ref(false)

async function handleSelectFile() {
  try {
    const file = await SelectFile()
    if (file) {
      selectedFile.value = file
      result.value = null
      previewRecords.value = []
      showPreview.value = false
    }
  } catch (e) {
    console.error('File selection failed:', e)
  }
}

async function handleExtract() {
  if (!selectedFile.value) return
  
  isLoading.value = true
  result.value = null
  
  try {
    const res = await ExtractFromFile(selectedFile.value)
    result.value = res
  } catch (e: any) {
    result.value = {
      success: false,
      recordCount: 0,
      outputPath: '',
      errorMessage: e.message || 'Unknown error'
    }
  } finally {
    isLoading.value = false
  }
}

async function handlePreview() {
  if (!selectedFile.value) return
  
  isLoading.value = true
  
  try {
    const res = await PreviewData(selectedFile.value)
    if (res.success && res.records) {
      previewRecords.value = res.records
      showPreview.value = true
    }
  } catch (e) {
    console.error('Preview failed:', e)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="app-container">
    <!-- Header -->
    <header class="header">
      <div class="logo">
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
          <line x1="16" y1="13" x2="8" y2="13"></line>
          <line x1="16" y1="17" x2="8" y2="17"></line>
          <polyline points="10 9 9 9 8 9"></polyline>
        </svg>
      </div>
      <h1>Legal Document Extractor</h1>
      <p class="subtitle">提取法律文书信息至 CSV</p>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <!-- File Selection Card -->
      <div class="card">
        <h2>📁 选择文件</h2>
        <div class="file-selector">
          <button class="btn btn-primary" @click="handleSelectFile">
            <span>选择 .docx 文件</span>
          </button>
          <div v-if="fileName" class="file-name">
            <span class="file-icon">📄</span>
            {{ fileName }}
          </div>
        </div>
      </div>

      <!-- Action Buttons Card -->
      <div class="card" v-if="selectedFile">
        <h2>⚡ 操作</h2>
        <div class="action-buttons">
          <button 
            class="btn btn-secondary" 
            @click="handlePreview"
            :disabled="isLoading"
          >
            👁️ 预览数据
          </button>
          <button 
            class="btn btn-success" 
            @click="handleExtract"
            :disabled="isLoading"
          >
            <span v-if="isLoading">⏳ 处理中...</span>
            <span v-else>🚀 提取并保存</span>
          </button>
        </div>
      </div>

      <!-- Result Card -->
      <div class="card result-card" v-if="result">
        <div v-if="result.success" class="result success">
          <h2>✅ 提取成功</h2>
          <p><strong>提取记录数：</strong>{{ result.recordCount }} 条</p>
          <p><strong>保存路径：</strong></p>
          <code class="output-path">{{ result.outputPath }}</code>
        </div>
        <div v-else class="result error">
          <h2>❌ 提取失败</h2>
          <p>{{ result.errorMessage }}</p>
        </div>
      </div>

      <!-- Preview Table -->
      <div class="card preview-card" v-if="showPreview && previewRecords.length > 0">
        <h2>📋 数据预览 (前 {{ Math.min(previewRecords.length, 10) }} 条)</h2>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                <th>被告</th>
                <th>身份证号码</th>
                <th>诉讼请求</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, index) in previewRecords.slice(0, 10)" :key="index">
                <td>{{ record.defendant }}</td>
                <td>{{ record.idNumber }}</td>
                <td class="truncate">{{ record.request?.substring(0, 50) }}...</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="preview-note">共 {{ previewRecords.length }} 条记录</p>
      </div>
    </main>

    <!-- Footer -->
    <footer class="footer">
      <p>Built with Wails + Vue 3 • © 2026</p>
    </footer>
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  color: #e8e8e8;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
}

.header {
  text-align: center;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo {
  display: inline-flex;
  padding: 0.75rem;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 12px;
  margin-bottom: 1rem;
}

.header h1 {
  font-size: 1.75rem;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  color: #888;
  margin-top: 0.5rem;
}

.main-content {
  flex: 1;
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.card {
  background: rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

.card h2 {
  font-size: 1.1rem;
  margin: 0 0 1rem 0;
  color: #b8b8b8;
}

.file-selector {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: rgba(102, 126, 234, 0.2);
  border-radius: 8px;
  font-family: monospace;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.15);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-success {
  background: linear-gradient(135deg, #11998e, #38ef7d);
  color: white;
}

.action-buttons {
  display: flex;
  gap: 1rem;
}

.result-card .success {
  color: #38ef7d;
}

.result-card .error {
  color: #ff6b6b;
}

.output-path {
  display: block;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  word-break: break-all;
  font-size: 0.875rem;
  margin-top: 0.5rem;
}

.table-container {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

th, td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

th {
  background: rgba(102, 126, 234, 0.2);
  font-weight: 600;
}

.truncate {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.preview-note {
  text-align: center;
  color: #888;
  margin-top: 1rem;
  font-size: 0.875rem;
}

.footer {
  text-align: center;
  padding: 1rem;
  color: #666;
  font-size: 0.875rem;
}
</style>
