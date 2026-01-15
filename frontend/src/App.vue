<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import {
  SelectFile,
  SelectOutputPath,
  ExtractToPath,
  PreviewData,
} from "../wailsjs/go/main/App";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";

interface Record {
  defendant: string;
  idNumber: string;
  request: string;
  factsReason: string;
}

interface ExtractResult {
  success: boolean;
  recordCount: number;
  outputPath: string;
  errorMessage?: string;
  records?: Record[];
}

const selectedFile = ref<string>("");
const outputOutputPath = ref<string>("");
const fileName = computed(() =>
  selectedFile.value ? selectedFile.value.split("/").pop() : ""
);
const isLoading = ref(false);
const result = ref<ExtractResult | null>(null);
const previewRecords = ref<Record[]>([]);
const showPreview = ref(false);
const isDragging = ref(false);
const notification = ref<{ message: string; type: 'success' | 'error' | 'info' } | null>(null);

function showNotification(message: string, type: 'success' | 'error' | 'info' = 'info') {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 3000);
}

// Wails 原生拖拽处理
onMounted(() => {
  OnFileDrop((x: number, y: number, paths: string[]) => {
    console.log("OnFileDrop triggered:", { x, y, paths });
    isDragging.value = false;
    if (paths && paths.length > 0) {
      const filePath = paths[0];
      if (filePath.toLowerCase().endsWith(".docx")) {
        console.log("Setting file:", filePath);
        setFile(filePath);
        showNotification("文件已加载", "success");
      } else {
        console.warn("请拖拽 .docx 文件");
        showNotification("不支持的文件格式，请使用 .docx 文件", "error");
      }
    }
  }, true);
});

onUnmounted(() => {
  OnFileDropOff();
});

async function handleSelectFile() {
  try {
    const file = await SelectFile();
    if (file) {
      setFile(file);
    }
  } catch (e) {
    console.error("File selection failed:", e);
  }
}

function setFile(file: string) {
  selectedFile.value = file;
  outputOutputPath.value = ""; // Reset output path when file changes
  result.value = null;
  previewRecords.value = [];
  showPreview.value = false;
}

function onDrop(e: DragEvent) {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    // Note: Wails drag and drop might need specific handling if not using the system dialog,
    // but for web-view based dropped files we can explicitly check if we can get the path.
    // However, in Wails v2 default drag/drop often gives the path if configured,
    // or we might strictly rely on the button if the browser environment forbids path access.
    // For this "Pro Max" UI, we'll assume the user might drag a file here.
    // If the browser security model blocks full path, we might need a workaround.
    // For now, let's keep the click-to-select as primary, but styling the area as drop-zone.
    // If we can't get the full path from a drop event in standard webview without extra setup,
    // we will trigger the select file dialog even on click of the drop zone.

    // Actually, usually in Wails one uses the system dialog.
    // Let's treat the drop zone mainly as a big click trigger for safety unless we're sure about drop handling.
    // But we'll add the visual feedback for dragging.
    handleSelectFile();
  }
}

async function handleSelectOutput() {
  if (!selectedFile.value) return;
  
  // Suggest a default name based on input file
  const baseName = (fileName.value || "document.docx").replace(/\.[^/.]+$/, "") + "_extracted.csv";
  
  try {
    const path = await SelectOutputPath(baseName);
    if (path) {
      outputOutputPath.value = path;
    }
  } catch (e) {
    console.error("Output selection failed:", e);
  }
}

async function handleExtract() {
  if (!selectedFile.value || !outputOutputPath.value) return;

  isLoading.value = true;
  result.value = null;

  try {
    const res = await ExtractToPath(selectedFile.value, outputOutputPath.value);
    
    result.value = res;
    if (res.success) {
      showNotification("提取成功！已保存至 " + res.outputPath, "success");
    } else {
      showNotification(res.errorMessage || "提取失败", "error");
    }
  } catch (e: any) {
    result.value = {
      success: false,
      recordCount: 0,
      outputPath: "",
      errorMessage: e.message || "Unknown error",
    };
  } finally {
    isLoading.value = false;
  }
}

async function handlePreview() {
  if (!selectedFile.value) return;

  isLoading.value = true;

  try {
    const res = await PreviewData(selectedFile.value);
    if (res.success && res.records) {
      previewRecords.value = res.records;
      showPreview.value = true;
    }
  } catch (e) {
    console.error("Preview failed:", e);
  } finally {
    isLoading.value = false;
  }
}
</script>

<template>
  <div class="app-container">
    <!-- Background Decor -->
    <div class="bg-blur-1"></div>
    <div class="bg-blur-2"></div>

    <!-- Notification Toast -->
    <Transition name="toast">
      <div v-if="notification" class="toast" :class="notification.type">
        <span class="toast-icon">{{ notification.type === 'error' ? '⚠️' : notification.type === 'success' ? '✅' : 'ℹ️' }}</span>
        <span class="toast-message">{{ notification.message }}</span>
      </div>
    </Transition>

    <main class="main-content">
      <!-- Header -->
      <header class="header">
        <div class="logo-container">
          <div class="logo-icon">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path
                d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
              ></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
          </div>
          <span class="logo-text text-gradient-brand">LegalExtractor</span>
        </div>
        <h1 class="title">
          法律文书<span class="text-gradient-brand">智能提取</span>
        </h1>
        <p class="subtitle">高效、精准的 .docx 数据提取工具</p>
      </header>

      <!-- Main Action Area -->
      <div class="main-card glass-panel">
        <!-- Drop Zone / File Selector -->
        <div
          class="drop-zone"
          :class="{ 'is-dragging': isDragging, 'has-file': !!selectedFile }"
          style="--wails-drop-target: drop"
          @click="handleSelectFile"
        >
          <div class="drop-content">
            <div class="icon-wrapper">
              <span v-if="!selectedFile" class="icon">📂</span>
              <span v-else class="icon">📄</span>
            </div>
            <div class="text-content">
              <h3 v-if="!selectedFile">点击或拖拽上传文件</h3>
              <h3 v-else>{{ fileName }}</h3>
              <p v-if="!selectedFile" class="hint">支持 .docx 格式法律文书</p>
              <p v-else class="hint file-path">{{ selectedFile }}</p>
            </div>
            <div v-if="selectedFile" class="change-file-btn">更换</div>
          </div>
        </div>

        <!-- Output Configuration -->
        <div class="output-config glass-panel" v-if="selectedFile">
          <div class="config-row">
            <span class="config-label">导出位置：</span>
            <div class="path-display" :class="{ 'placeholder': !outputOutputPath }">
              {{ outputOutputPath || '请选择保存位置...' }}
            </div>
            <button class="btn btn-sm btn-secondary" @click="handleSelectOutput">
              {{ outputOutputPath ? '更改' : '选择' }}
            </button>
          </div>
        </div>

        <!-- Actions -->
        <div class="actions" v-if="selectedFile">
          <button
            class="btn btn-secondary"
            @click.stop="handlePreview"
            :disabled="isLoading"
          >
            <span>👁️ 预览数据</span>
          </button>

          <button
            class="btn btn-primary"
            @click.stop="handleExtract"
            :disabled="isLoading || !outputOutputPath"
            :title="!outputOutputPath ? '请先选择保存位置' : ''"
          >
            <span v-if="isLoading" class="loader"></span>
            <span v-else>🚀 开始提取</span>
          </button>
        </div>
      </div>

      <!-- Result Section -->
      <Transition name="fade">
        <div
          v-if="result"
          class="result-card glass-panel"
          :class="{ error: !result.success }"
        >
          <div class="result-header">
            <span class="status-icon">{{ result.success ? "✅" : "❌" }}</span>
            <h3>{{ result.success ? "提取成功" : "提取失败" }}</h3>
          </div>

          <div v-if="result.success" class="result-body">
            <div class="stat-item">
              <span class="label">提取记录</span>
              <span class="value">{{ result.recordCount }}</span>
            </div>
            <div class="path-box">
              <span class="label">保存至：</span>
              <code>{{ result.outputPath }}</code>
            </div>
          </div>
          <div v-else class="result-body">
            <p class="error-msg">{{ result.errorMessage }}</p>
          </div>
        </div>
      </Transition>

      <!-- Preview Table -->
      <Transition name="slide-up">
        <div
          v-if="showPreview && previewRecords.length > 0"
          class="preview-section glass-panel"
        >
          <div class="preview-header">
            <h3>数据预览</h3>
            <span class="badge">{{ previewRecords.length }} 条记录</span>
          </div>
          <div class="table-wrapper">
            <table>
              <colgroup>
                <col style="width: 100px" />
                <col style="width: 180px" />
                <col style="width: auto" />
                <col style="width: auto" />
              </colgroup>
              <thead>
                <tr>
                  <th>被告</th>
                  <th>身份证号码</th>
                  <th>诉讼请求</th>
                  <th>事实与理由</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(record, index) in previewRecords.slice(0, 50)"
                  :key="index"
                >
                  <td>
                    <div class="cell-content fixed-text" :title="record.defendant">{{ record.defendant }}</div>
                  </td>
                  <td>
                    <div class="cell-content fixed-text" :title="record.idNumber">{{ record.idNumber }}</div>
                  </td>
                  <td>
                    <div class="cell-content truncate" :title="record.request">
                      {{ record.request }}
                    </div>
                  </td>
                  <td>
                    <div class="cell-content truncate" :title="record.factsReason">
                      {{ record.factsReason }}
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </Transition>
    </main>

    <footer class="footer">
      <p>Powered by Wails & Vue 3</p>
    </footer>
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  /* padding: var(--spacing-lg); */
  padding: 40px 20px;
  position: relative;
  overflow-x: hidden;
}

/* Background Blurs */
.bg-blur-1 {
  position: absolute;
  top: -10%;
  left: -10%;
  width: 50vw;
  height: 50vw;
  background: radial-gradient(
    circle,
    rgba(59, 130, 246, 0.15) 0%,
    transparent 70%
  );
  filter: blur(80px);
  z-index: -1;
  pointer-events: none;
}

.bg-blur-2 {
  position: absolute;
  bottom: -10%;
  right: -10%;
  width: 60vw;
  height: 60vw;
  background: radial-gradient(
    circle,
    rgba(139, 92, 246, 0.15) 0%,
    transparent 70%
  );
  filter: blur(80px);
  z-index: -1;
  pointer-events: none;
}

/* Main Content */
.main-content {
  width: 100%;
  max-width: 800px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  z-index: 1;
}

/* Header */
.header {
  text-align: center;
  margin-bottom: var(--spacing-md);
}

.logo-container {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  margin-bottom: var(--spacing-sm);
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--radius-full);
  border: 1px solid var(--surface-border);
}

.logo-icon {
  color: var(--accent-primary);
  display: flex;
}

.logo-text {
  font-weight: 600;
  font-size: 0.9rem;
  letter-spacing: 0.5px;
}

.title {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: var(--spacing-xs);
  line-height: 1.2;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
}

/* Main Card */
.main-card {
  padding: var(--spacing-md);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.main-card:hover {
  box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.3);
}

/* Drop Zone */
.drop-zone {
  border: 2px dashed rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-md);
  padding: var(--spacing-xl) var(--spacing-md);
  text-align: center;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: rgba(255, 255, 255, 0.01);
  position: relative;
  overflow: hidden;
}

.drop-zone:hover,
.drop-zone.is-dragging,
.drop-zone.wails-drop-target-active {
  border-color: var(--accent-primary);
  background: rgba(59, 130, 246, 0.05);
  transform: scale(1.01);
}

.drop-zone.has-file {
  border-style: solid;
  background: rgba(59, 130, 246, 0.08);
  border-color: rgba(59, 130, 246, 0.3);
}

.drop-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-sm);
}

.icon-wrapper {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 50%;
  font-size: 2rem;
  margin-bottom: var(--spacing-xs);
  transition: transform 0.3s ease;
}

.drop-zone:hover .icon-wrapper {
  transform: translateY(-5px) scale(1.1);
  background: rgba(59, 130, 246, 0.2);
}

.text-content h3 {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
}

.hint {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.file-path {
  font-family: monospace;
  background: rgba(0, 0, 0, 0.2);
  padding: 4px 8px;
  border-radius: 4px;
  max-width: 100%;
  word-break: break-all;
}

.change-file-btn {
  margin-top: var(--spacing-sm);
  font-size: 0.8rem;
  color: var(--accent-primary);
  text-decoration: underline;
  opacity: 0.8;
}

.change-file-btn:hover {
  opacity: 1;
}

/* Actions */
.actions {
  display: flex;
  gap: var(--spacing-sm);
  justify-content: center;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.8rem 1.8rem;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 140px;
}

.btn-primary {
  background: linear-gradient(
    135deg,
    var(--accent-primary),
    var(--accent-secondary)
  );
  color: white;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.4);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(59, 130, 246, 0.5);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
  border: 1px solid var(--surface-border);
}

.btn-secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* Result Card */
.result-card {
  border-radius: var(--radius-lg);
  padding: var(--spacing-md);
  border-left: 4px solid var(--success);
}

.result-card.error {
  border-left-color: var(--error);
}

.result-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.status-icon {
  font-size: 1.5rem;
}

.result-body {
  margin-left: calc(1.5rem + var(--spacing-sm));
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xs);
}

.stat-item .value {
  font-weight: 700;
  font-size: 1.2rem;
  color: var(--success);
}

.path-box {
  background: rgba(0, 0, 0, 0.2);
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  margin-top: var(--spacing-sm);
  font-family: monospace;
  font-size: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.path-box code {
  color: var(--accent-primary);
  word-break: break-all;
}

.error-msg {
  color: var(--error);
}

/* Preview Section */
.preview-section {
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 500px;
}

.preview-header {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--surface-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 0, 0, 0.1);
}

.badge {
  background: rgba(255, 255, 255, 0.1);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font-size: 0.8rem;
  font-weight: 500;
}

.table-wrapper {
  overflow-y: auto;
  flex: 1;
}

table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed; /* Enforce fixed column widths */
}

th {
  background: rgba(255, 255, 255, 0.02);
  padding: var(--spacing-sm);
  text-align: left;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-secondary);
  position: sticky;
  top: 0;
  backdrop-filter: blur(10px);
  z-index: 10;
  white-space: nowrap;
}

td {
  padding: var(--spacing-sm);
  border-bottom: 1px solid var(--surface-border);
  font-size: 0.9rem;
  vertical-align: middle;
}

tr:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.cell-content {
  width: 100%;
}

.fixed-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.truncate {
  width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Footer */
.footer {
  margin-top: auto;
  padding: var(--spacing-lg);
  color: var(--text-muted);
  font-size: 0.8rem;
}

/* Animations */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.4s cubic-bezier(0.25, 1, 0.5, 1);
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.loader {
  width: 16px;
  height: 16px;
  border: 2px solid #fff;
  border-bottom-color: transparent;
  border-radius: 50%;
  display: inline-block;
  box-sizing: border-box;
  animation: rotation 1s linear infinite;
}

@keyframes rotation {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* Toast Notification */
.toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(30, 41, 59, 0.9);
  backdrop-filter: blur(12px);
  border: 1px solid var(--surface-border);
  padding: 12px 24px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  min-width: 300px;
  justify-content: center;
}

.toast.success {
  border-color: var(--success);
}

.toast.error {
  border-color: var(--error);
}

.toast-icon {
  font-size: 1.2rem;
}

.toast-message {
  font-size: 0.95rem;
  font-weight: 500;
}

/* Output Config */
.output-config {
  padding: var(--spacing-md);
  border-radius: var(--radius-lg);
  margin-top: -10px; /* Slight overlap for visual grouping or separate it if preferred */
  margin-bottom: var(--spacing-sm);
  background: rgba(255, 255, 255, 0.03);
}

.config-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.config-label {
  color: var(--text-secondary);
  font-size: 0.95rem;
  font-weight: 500;
  white-space: nowrap;
}

.path-display {
  flex: 1;
  background: rgba(0, 0, 0, 0.2);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border: 1px solid transparent;
}

.path-display.placeholder {
  color: var(--text-muted);
  font-style: italic;
}

.btn-sm {
  padding: 0.4rem 1rem;
  font-size: 0.9rem;
  min-width: unset;
}


/* Toast Animation */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>
