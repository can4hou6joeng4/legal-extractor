<script setup lang="ts">
import { ref, computed } from "vue";
import { api, downloadBlob, type Record, type ExtractResult } from "./services";
import MainDropZone from "./components/MainDropZone.vue";
import ConfigPanel from "./components/ConfigPanel.vue";
import ResultCard from "./components/ResultCard.vue";
import PreviewTable from "./components/PreviewTable.vue";

// State
const selectedFile = ref<string | File | null>(null);
const selectedFields = ref<string[]>([]);
const fieldLabels = ref<{ [key: string]: string }>({});
const selectedFormat = ref<"xlsx" | "csv" | "json">("xlsx");
const outputOutputPath = ref<string>("");

const fileName = computed(() => {
  if (!selectedFile.value) return "";
  if (typeof selectedFile.value === "string") {
    // Desktop path
    return selectedFile.value.split(/[\\/]/).pop() || "";
  } else {
    // Web File object
    return selectedFile.value.name;
  }
});

const isLoading = ref(false);
const result = ref<ExtractResult | null>(null);
const previewRecords = ref<Record[]>([]);
const showPreview = ref(false);

const notification = ref<{
  message: string;
  type: "success" | "error" | "info";
} | null>(null);

// Actions
function showNotification(
  message: string,
  type: "success" | "error" | "info" = "info",
) {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 3000);
}

function handleFileUpdate(file: string | File) {
  selectedFile.value = file;
  // Reset state when file changes
  outputOutputPath.value = "";
  selectedFields.value = []; // Clear previous selection
  result.value = null;
  previewRecords.value = [];
  showPreview.value = false;
}

async function handlePreview() {
  if (!selectedFile.value) return;

  isLoading.value = true;
  try {
    const res = await api.service.previewData(
      selectedFile.value,
      selectedFields.value,
    );
    if (res.success && res.records) {
      previewRecords.value = res.records;
      fieldLabels.value = res.fieldLabels || {};
      showPreview.value = true;
    } else if (res.errorMessage) {
      showNotification(res.errorMessage, "error");
    }
  } catch (e) {
    console.error("Preview failed:", e);
    showNotification("预览失败: " + (e as Error).message, "error");
  } finally {
    isLoading.value = false;
  }
}

async function handleExtract() {
  if (!selectedFile.value) return;

  isLoading.value = true;
  result.value = null;

  try {
    const defaultExt = selectedFormat.value;
    const baseName = fileName.value.includes('.')
      ? fileName.value.substring(0, fileName.value.lastIndexOf('.'))
      : fileName.value;
    const defaultName = `提取结果_${baseName}.${defaultExt}`;

    let finalOutputPath = "";

    // 仅 Desktop 模式需要选择输出路径
    if (api.isDesktop) {
      // 逻辑优化：优先使用用户在界面上选择的路径
      // 只有当 outputOutputPath 为空时，才弹出选择框
      finalOutputPath = outputOutputPath.value;

      if (!finalOutputPath) {
        try {
          finalOutputPath = await api.service.selectOutputPath(defaultName);
          if (finalOutputPath) {
            outputOutputPath.value = finalOutputPath;
          } else {
            // 用户取消选择
            isLoading.value = false;
            return;
          }
        } catch (e) {
          console.error("Select output path failed:", e);
          showNotification("选择保存路径失败", "error");
          isLoading.value = false;
          return;
        }
      }
    }

    // 调用提取接口
    // Web 模式下 finalOutputPath 为空，后端仅提取数据，前端负责导出
    const res = await api.service.extractToPath(
      selectedFile.value,
      finalOutputPath,
      selectedFields.value,
    );

    if (res.success) {
      // Desktop 模式：文件已保存
      // Web 模式：需要触发下载
      if (api.isWeb && res.records) {
        try {
          const blob = await api.service.exportData(res.records, selectedFormat.value) as Blob;
          downloadBlob(blob, defaultName);
          res.outputPath = "浏览器下载目录"; // 更新 UI 显示
        } catch (err) {
          showNotification("导出文件失败: " + (err as Error).message, "error");
        }
      }

      result.value = res;
      showNotification(`提取成功！共 ${res.recordCount} 条记录`, "success");
    } else {
      result.value = {
        success: false,
        recordCount: 0,
        outputPath: "",
        errorMessage: res.errorMessage || "未知错误",
      };
      showNotification(res.errorMessage || "提取失败", "error");
    }
  } catch (e) {
    console.error("Extraction failed:", e);
    result.value = {
      success: false,
      recordCount: 0,
      outputPath: "",
      errorMessage: (e as Error).message,
    };
    showNotification("提取过程发生错误", "error");
  } finally {
    isLoading.value = false;
  }
}

async function handleSelectOutputPath() {
  // Web 模式不支持选择输出路径
  if (api.isWeb) return;

  try {
    const path = await api.service.selectOutputPath("");
    if (path) {
      outputOutputPath.value = path;
    }
  } catch (e) {
    console.error("Select output path failed:", e);
  }
}

async function handleOpenFile(path: string) {
  // Web 模式不支持打开本地文件
  if (api.isWeb) return;

  try {
    await api.service.openFile(path);
  } catch (e) {
    console.error("Open file failed:", e);
    showNotification("打开文件失败", "error");
  }
}

function handleFieldsChange(fields: string[]) {
  selectedFields.value = fields;
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
        <span class="toast-icon">
            <svg v-if="notification.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
            <svg v-else-if="notification.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="M22 4 12 14.01l-3-3"/></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
        </span>
        <span class="toast-message">{{ notification.message }}</span>
      </div>
    </Transition>

    <!-- Loading Overlay -->
    <Transition name="fade">
      <div v-if="isLoading" class="loading-overlay">
        <div class="loading-spinner">
          <svg class="animate-spin" xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
          </svg>
        </div>
        <div class="loading-content">
          <h3 class="loading-title">正在处理中...</h3>
          <p class="loading-desc" v-if="selectedFile.toLowerCase().endsWith('.pdf')">
            正在进行文档智能解析，请稍候...
          </p>
          <p class="loading-desc" v-else>正在解析本地文档结构...</p>
        </div>
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
          <span class="logo-text font-heading text-gradient-brand">LegalExtractor</span>
        </div>
        <h1 class="title font-heading">
          法律文书<span class="text-gradient-brand">智能提取</span>
        </h1>
        <p class="subtitle">高效、精准的 .docx / .pdf 数据提取工具</p>
      </header>

      <!-- Main Action Area -->
      <div class="main-card glass-panel">
        <MainDropZone
          :selectedFile="selectedFile"
          :fileName="fileName"
          @update:selectedFile="handleFileUpdate"
          @notification="showNotification"
        />

        <ConfigPanel
          v-if="selectedFile"
          :selectedFile="selectedFile"
          :fileName="fileName"
          v-model:selectedFormat="selectedFormat"
          v-model:outputOutputPath="outputOutputPath"
          v-model:selectedFields="selectedFields"
          :isLoading="isLoading"
          @preview="handlePreview"
          @extract="handleExtract"
        />
      </div>

      <!-- Result Section -->
      <ResultCard :result="result" @notification="showNotification" />

      <!-- Preview Table -->
      <Transition name="slide-up">
        <PreviewTable
          v-if="showPreview && previewRecords.length > 0"
          :records="previewRecords"
          :fieldLabels="fieldLabels"
        />
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
  height: 100vh; /* Fixed height to viewport */
  overflow-y: auto; /* Handle scrolling internally */
}

/* Background Blurs */
.bg-blur-1 {
  position: absolute;
  top: -10%;
  left: -10%;
  width: 50vw;
  height: 50vw;
  background: radial-gradient(circle, var(--accent-glow) 0%, transparent 70%);
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
    var(--accent-secondary-glow) 0%,
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
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
  background: rgba(255, 255, 255, 0.05); /* Base glass effect */
}

.main-card:hover {
  box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.3);
}

.glass-panel {
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* Footer */
.footer {
  margin-top: 40px;
  color: var(--text-muted);
  font-size: 0.8rem;
}

/* Transitions */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-20px) translateX(-50%);
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* Toast Styles */
.toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(20, 20, 30, 0.9);
  backdrop-filter: blur(12px);
  padding: 12px 20px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  z-index: 100;
  border: 1px solid rgba(255, 255, 255, 0.1);
  min-width: 300px;
}

.toast.success {
  border-color: rgba(var(--success-rgb), 0.3);
}

.toast.error {
  border-color: rgba(var(--error-rgb), 0.3);
}

.toast-icon {
  font-size: 1.2rem;
}

.toast-message {
  font-size: 0.95rem;
  color: var(--text-primary);
}
</style>
