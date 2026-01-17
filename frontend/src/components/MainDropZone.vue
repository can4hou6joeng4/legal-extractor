<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { SelectFile } from "../../wailsjs/go/app/App";
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";

const props = defineProps<{
  selectedFile: string;
  fileName: string;
}>();

const emit = defineEmits<{
  (e: "update:selectedFile", value: string): void;
  (
    e: "notification",
    message: string,
    type: "success" | "error" | "info"
  ): void;
}>();

const isDragging = ref(false);

function setFile(file: string) {
  emit("update:selectedFile", file);
}

// Wails 原生拖拽处理
onMounted(() => {
  OnFileDrop((x: number, y: number, paths: string[]) => {
    isDragging.value = false;
    if (paths && paths.length > 0) {
      const filePath = paths[0];
      const lowerPath = filePath.toLowerCase();
      if (lowerPath.endsWith(".docx") || lowerPath.endsWith(".pdf")) {
        setFile(filePath);
        emit("notification", "文件已加载", "success");
      } else {
        emit(
          "notification",
          "不支持的文件格式，请使用 .docx 或 .pdf 文件",
          "error"
        );
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
</script>

<template>
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
        <p v-if="!selectedFile" class="hint">支持 .docx / .pdf 格式法律文书</p>
        <p v-else class="hint file-path">{{ selectedFile }}</p>
      </div>
      <div v-if="selectedFile" class="change-file-btn">更换</div>
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
  transform: scale(1.01);
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
</style>
