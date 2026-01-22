<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from "vue";
import { api } from "../services";
// 动态导入 Wails 运行时，避免 Web 模式下报错（实际上 api.isDesktop 保护下不会执行，但为了安全）
// 注意：在 Vite 构建中，import 可能会被静态分析，所以这里保留原有导入，但在运行时做判断

const props = defineProps<{
  selectedFile: string | File | null;
  fileName: string;
}>();

const displayPath = computed(() => {
  if (!props.selectedFile) return "";
  if (typeof props.selectedFile === "string") {
    return props.selectedFile;
  } else {
    // Web File object: display size
    const size = props.selectedFile.size;
    if (size < 1024) return size + " B";
    if (size < 1024 * 1024) return (size / 1024).toFixed(1) + " KB";
    return (size / (1024 * 1024)).toFixed(1) + " MB";
  }
});

const emit = defineEmits<{
  (e: "update:selectedFile", value: string | File): void;
  (
    e: "notification",
    message: string,
    type: "success" | "error" | "info"
  ): void;
}>();

const isDragging = ref(false);

function setFile(file: string | File) {
  emit("update:selectedFile", file);
}

// 桌面版：Wails 原生拖拽处理
let cleanupWailsDrop: (() => void) | null = null;

onMounted(async () => {
  if (api.isDesktop) {
    try {
      const { OnFileDrop, OnFileDropOff } = await import("../../wailsjs/runtime/runtime");
      OnFileDrop((x: number, y: number, paths: string[]) => {
        isDragging.value = false;
        if (paths && paths.length > 0) {
          const filePath = paths[0];
          const lowerPath = filePath.toLowerCase();
          if (lowerPath.endsWith(".docx") || lowerPath.endsWith(".pdf") || lowerPath.endsWith(".jpg") || lowerPath.endsWith(".png")) {
            setFile(filePath);
            emit("notification", "文件已加载", "success");
          } else {
            emit("notification", "不支持的文件格式", "error");
          }
        }
      }, true);

      cleanupWailsDrop = OnFileDropOff;
    } catch (e) {
      console.warn("Wails runtime not available:", e);
    }
  }
});

onUnmounted(() => {
  if (cleanupWailsDrop) {
    cleanupWailsDrop();
  }
});

// Web 版：HTML5 拖拽处理
function handleWebDrop(e: DragEvent) {
  if (api.isDesktop) return; // 桌面版由 Wails 处理

  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    const file = files[0];
    const name = file.name.toLowerCase();
    if (name.endsWith(".docx") || name.endsWith(".pdf") || name.endsWith(".jpg") || name.endsWith(".png")) {
      setFile(file);
      emit("notification", "文件已加载", "success");
    } else {
      emit("notification", "不支持的文件格式", "error");
    }
  }
}

async function handleSelectFile() {
  try {
    const file = await api.service.selectFile();
    if (file) {
      setFile(file);
    }
  } catch (e) {
    console.error("File selection failed:", e);
    // 用户取消选择不报错
    if ((e as Error).message !== "未选择文件") {
       emit("notification", "选择文件失败", "error");
    }
  }
}
</script>

<template>
  <div
    class="drop-zone"
    :class="{ 'is-dragging': isDragging, 'has-file': !!selectedFile }"
    style="--wails-drop-target: drop"
    @click="handleSelectFile"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="handleWebDrop"
  >
    <div class="drop-content">
      <div class="icon-wrapper">
        <div v-if="!selectedFile" class="icon-svg">
            <!-- Folder Icon -->
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
        </div>
        <div v-else class="icon-svg">
            <!-- File Icon -->
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/></svg>
        </div>
      </div>
      <div class="text-content">
        <h3 v-if="!selectedFile">点击或拖拽上传文件</h3>
        <div v-else class="selected-file-info">
          <h3 class="file-name-display">{{ fileName }}</h3>
          <p class="file-path-text" :title="String(selectedFile)">{{ selectedFile }}</p>
        </div>
        <p v-if="!selectedFile" class="hint">支持 .docx / .pdf 格式法律文书</p>
      </div>
      <button v-if="selectedFile" class="change-file-btn">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2v6h-6"/><path d="M3 12a9 9 0 0 1 15-6.7L21 8"/><path d="M3 22v-6h6"/><path d="M21 12a9 9 0 0 1-15 6.7L3 16"/></svg>
        <span>更换文件</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
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
  background: color-mix(in srgb, var(--accent-primary) 5%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--accent-primary) 10%, transparent);
}

.drop-zone.has-file {
  border-style: solid;
  background: color-mix(in srgb, var(--accent-primary) 8%, transparent);
  border-color: color-mix(in srgb, var(--accent-primary) 30%, transparent);
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
  background: color-mix(in srgb, var(--accent-primary) 20%, transparent);
}

.text-content h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-heading);
  letter-spacing: 0.5px;
}

.selected-file-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.file-name-display {
  color: var(--text-primary) !important;
  max-width: 500px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-path-text {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-family: var(--font-body);
  max-width: 450px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  opacity: 0.7;
}

.hint {
  color: var(--text-muted);
  font-size: 0.9rem;
  margin-top: 4px;
}

.change-file-btn {
  margin-top: var(--spacing-lg);
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary);
  padding: 8px 18px;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.change-file-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--accent-primary);
  border-color: var(--accent-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
}

.change-file-btn:hover svg {
  transform: rotate(180deg);
}

.change-file-btn svg {
  transition: transform 0.5s ease;
}
</style>
