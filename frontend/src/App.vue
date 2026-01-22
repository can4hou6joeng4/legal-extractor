<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from "vue";
import { api, downloadBlob, type Record, type ExtractResult } from "./services";
import MainDropZone from "./components/MainDropZone.vue";
import ConfigPanel from "./components/ConfigPanel.vue";
import ResultCard from "./components/ResultCard.vue";
import PreviewTable from "./components/PreviewTable.vue";

// Trial State
interface TrialStatus {
  isActivated: boolean;
  isExpired: boolean;
  remaining: number;
  days: number;
  hours: number;
}

const trialStatus = ref<TrialStatus | null>(null);

// State
const selectedFile = ref<string | File | null>(null);
const machineID = ref("");
const licenseKey = ref("");
const showActivationModal = ref(false);
const selectedFields = ref<string[]>([]);

// 4. Close modal clear input
watch(showActivationModal, (newVal) => {
  if (!newVal) {
    licenseKey.value = "";
  }
});

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
const loadingText = ref("");
const result = ref<ExtractResult | null>(null);
const previewRecords = ref<Record[]>([]);
const showPreview = ref(false);

const notification = ref<{
  message: string;
  type: "success" | "error" | "info";
} | null>(null);

// Actions
async function fetchTrialStatus() {
  try {
    const status = await api.service.getTrialStatus();
    trialStatus.value = status;
    // 同时获取机器码
    const mid = await api.service.getMachineID();
    machineID.value = mid;
  } catch (e) {
    console.error("Failed to fetch trial status:", e);
  }
}

// 2. License validation
const isValidLicense = computed(() => {
  const k = licenseKey.value.trim();
  // Simple format check: XXXX-XXXX-XXXX-XXXX (approx 19 chars)
  // We relax it slightly to allow loose input but prevent empty/short nonsense
  return k.length >= 16;
});

async function handleActivate() {
  const key = licenseKey.value.trim();
  if (!key) return;

  loadingText.value = "正在验证授权...";
  isLoading.value = true;

  // 5. 强制延迟 500ms 确保 Loading 动画渲染出来，避免被系统弹窗（如权限请求）打断渲染
  await new Promise(resolve => setTimeout(resolve, 500));

  try {
    // 改用标准服务层调用
    const success = await api.service.activate(key);

    // Stop loading BEFORE showing success to prevent overlay conflict
    isLoading.value = false;

    if (success) {
      showNotification("激活成功！感谢使用专业版", "success");
      showActivationModal.value = false;
      await fetchTrialStatus(); // 刷新 UI 状态
    } else {
      // 关键修复：增加失败反馈
      showNotification("授权码无效，请检查后重试", "error");
    }
  } catch (e) {
    isLoading.value = false;
    showNotification(String(e), "error");
  } finally {
    // Redundant safety check
    if (isLoading.value) isLoading.value = false;
    loadingText.value = "";
  }
}

function copyMachineID() {
  navigator.clipboard.writeText(machineID.value);
  showNotification("特征码已复制到剪贴板", "success");
}

onMounted(() => {
  fetchTrialStatus();
});

let toastTimer: ReturnType<typeof setTimeout> | null = null;

function showNotification(
  message: string,
  type: "success" | "error" | "info" = "info",
) {
  // 1. Clear previous timer to prevent race conditions
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }

  notification.value = { message, type };
  toastTimer = setTimeout(() => {
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
          <p class="loading-desc" v-if="loadingText">{{ loadingText }}</p>
          <template v-else>
            <p class="loading-desc" v-if="fileName.toLowerCase().endsWith('.pdf')">
              正在进行文档智能解析，请稍候...
            </p>
            <p class="loading-desc" v-else>正在解析本地文档结构...</p>
          </template>
        </div>
      </div>
    </Transition>

    <!-- Trial Banner -->
    <div v-if="trialStatus && api.isDesktop && !trialStatus.isActivated" class="trial-banner" :class="{ 'expired': trialStatus.isExpired }" role="status" aria-live="polite">
      <div class="trial-container">
        <template v-if="!trialStatus.isExpired">
          <span class="trial-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </span>
          <span class="trial-text">试用期剩余：<strong>{{ trialStatus.days }}</strong> 天 <strong>{{ trialStatus.hours }}</strong> 小时</span>
          <button class="trial-cta-btn" @click="showActivationModal = true">获取正式版</button>
        </template>
        <template v-else>
          <span class="trial-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
          </span>
          <span class="trial-text">试用期已结束，核心功能已锁定。</span>
          <button class="trial-cta-btn urgent" @click="showActivationModal = true">联系授权</button>
        </template>
      </div>
    </div>

    <!-- Professional Active Badge -->
    <div v-if="trialStatus?.isActivated" class="active-badge-fixed">
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
      <span>专业授权版</span>
    </div>

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
          :isLoading="isLoading || (trialStatus?.isExpired ?? false)"
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

    <!-- Activation Modal -->
    <Transition name="fade">
      <div v-if="showActivationModal" class="modal-overlay">
        <div class="activation-card glass-panel">
          <div class="modal-header">
            <h2 class="font-heading">软件激活中心</h2>
            <button class="close-btn" @click="showActivationModal = false">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>

          <div class="modal-body">
            <div class="info-section">
              <label>您的设备特征码</label>
              <div class="machine-id-box" @click="copyMachineID">
                <code>{{ machineID }}</code>
                <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
              </div>
              <p class="helper-text">请将上方代码发送给开发者以获取授权码</p>
            </div>

            <div class="input-section">
              <label>输入授权码</label>
              <input
                v-model="licenseKey"
                type="text"
                placeholder="XXXX-XXXX-XXXX-XXXX"
                class="license-input"
                @keyup.enter="handleActivate"
              />
            </div>

            <button class="btn btn-primary btn-glow full-width" @click="handleActivate" :disabled="!isValidLicense">
              立即激活专业版
            </button>
          </div>
        </div>
      </div>
    </Transition>

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
  padding-top: 60px; /* 为试用期横幅留出空间 */
  position: relative;
  overflow-x: hidden;
  height: 100vh; /* Fixed height to viewport */
  overflow-y: auto; /* Handle scrolling internally */
}

/* Trial Banner 试用期横幅 */
.trial-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 44px;
  background: linear-gradient(90deg, rgba(15, 23, 42, 0.95) 0%, rgba(30, 58, 138, 0.95) 100%);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  transition: all 0.3s ease;
}

.trial-container {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  max-width: 800px;
  width: 100%;
}

.trial-banner.expired {
  background: linear-gradient(90deg, rgba(69, 10, 10, 0.95) 0%, rgba(153, 27, 27, 0.95) 100%);
  border-bottom-color: rgba(239, 68, 68, 0.3);
}

.trial-icon {
  display: flex;
  align-items: center;
  color: var(--warning);
}

.trial-banner.expired .trial-icon {
  color: var(--error);
}

.trial-text {
  font-size: 0.85rem;
  color: var(--text-primary);
  font-weight: 500;
  letter-spacing: 0.3px;
}

.trial-text strong {
  color: var(--warning);
  margin: 0 2px;
}

.trial-cta-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-left: 8px;
}

.trial-cta-btn:hover {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  transform: scale(1.05);
}

.trial-cta-btn.urgent {
  background: var(--error);
  border-color: var(--error);
}

.trial-cta-btn.urgent:hover {
  background: #dc2626;
  box-shadow: 0 0 12px rgba(239, 68, 68, 0.4);
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
  gap: var(--spacing-sm); /* 从 lg 改为 sm，显著缩小间距 */
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
  top: 80px; /* 移至横幅下方，预留足够空间 */
  left: 50%;
  transform: translateX(-50%);
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  padding: 10px 24px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
  z-index: 6000; /* Ensure it's above the modal (5000) */
  border: 1px solid rgba(255, 255, 255, 0.1);
  min-width: 320px;
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

/* ============================
   Activation Modal (New UI)
   ============================ */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(2, 6, 23, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 5000;
}

.activation-card {
  width: 90%;
  max-width: 440px;
  padding: 40px;
  border-radius: 24px;
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  position: relative;
  animation: modalIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.9) translateY(20px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.modal-header h2 {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--text-primary);
}

.close-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-muted);
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.info-section label, .input-section label {
  display: block;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-weight: 500;
}

.machine-id-box {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(14, 165, 233, 0.2);
  padding: 14px 18px;
  border-radius: 12px;
  display: flex;
  justify-content: center; /* 3. Centered */
  position: relative;      /* For absolute positioning of icon */
  align-items: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.machine-id-box:hover {
  border-color: var(--accent-primary);
  background: rgba(14, 165, 233, 0.08);
}

.machine-id-box code {
  font-family: 'JetBrains Mono', monospace;
  color: var(--accent-primary);
  font-size: 1.2rem;
  letter-spacing: 2px;
  font-weight: 600;
}

.machine-id-box .copy-icon {
  position: absolute;
  right: 18px;
  color: var(--text-secondary);
}

.license-input {
  width: 100%;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 14px 18px;
  border-radius: 12px;
  color: var(--text-primary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.1rem;
  outline: none;
  transition: all 0.3s ease;
  text-align: center;
}

.license-input:focus {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 4px rgba(14, 165, 233, 0.15);
}

.full-width {
  width: 100%;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem !important;
}

/* Professional Active Badge */
.active-badge-fixed {
  position: fixed;
  top: 12px;
  right: 20px;
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #34d399;
  padding: 4px 12px;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
  backdrop-filter: blur(10px);
  z-index: 1000;
}

/* Loading Overlay */
.loading-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 5500;
  gap: 20px;
}

.loading-spinner {
  color: var(--accent-primary);
}

.loading-content {
  text-align: center;
}

.loading-title {
  font-size: 1.2rem;
  font-weight: 600;
  margin-bottom: 8px;
}

.loading-desc {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
</style>
