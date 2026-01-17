<script setup lang="ts">
import { SelectOutputPath } from "../../wailsjs/go/app/App";

const props = defineProps<{
  selectedFile: string;
  fileName: string;
  selectedFormat: "xlsx" | "csv" | "json";
  outputOutputPath: string;
  isLoading: boolean;
}>();

const emit = defineEmits<{
  (e: "update:selectedFormat", value: string): void;
  (e: "update:outputOutputPath", value: string): void;
  (e: "preview"): void;
  (e: "extract"): void;
}>();

async function handleSelectOutput() {
  if (!props.selectedFile) return;

  // Suggest a default name based on input file and selected format
  const ext = props.selectedFormat;
  const baseName =
    (props.fileName || "document.doc").replace(/\.[^/.]+$/, "") + "." + ext;

  try {
    const path = await SelectOutputPath(baseName);
    if (path) {
      emit("update:outputOutputPath", path);
      // Auto update format selection if user picked a different extension
      if (path.toLowerCase().endsWith(".json"))
        emit("update:selectedFormat", "json");
      else if (path.toLowerCase().endsWith(".csv"))
        emit("update:selectedFormat", "csv");
      else if (path.toLowerCase().endsWith(".xlsx"))
        emit("update:selectedFormat", "xlsx");
    }
  } catch (e) {
    console.error("Output selection failed:", e);
  }
}
</script>

<template>
  <div v-if="selectedFile">
    <!-- Output Configuration -->
    <div class="output-config glass-panel">
      <div class="config-row">
        <span class="config-label">导出格式：</span>
        <select
          :value="selectedFormat"
          @input="
            emit(
              'update:selectedFormat',
              ($event.target as HTMLSelectElement).value
            )
          "
          class="format-select"
        >
          <option value="xlsx">Excel 表格 (.xlsx)</option>
          <option value="csv">CSV 文件 (.csv)</option>
          <option value="json">JSON 数据 (.json)</option>
        </select>
      </div>
      <div class="divider"></div>
      <div class="config-row">
        <span class="config-label">导出位置：</span>
        <div class="path-display" :class="{ placeholder: !outputOutputPath }">
          {{ outputOutputPath || "请选择保存位置..." }}
        </div>
        <button class="btn btn-sm btn-secondary" @click="handleSelectOutput">
          {{ outputOutputPath ? "更改" : "选择" }}
        </button>
      </div>
    </div>

    <!-- Actions -->
    <div class="actions">
      <button
        class="btn btn-secondary"
        @click="emit('preview')"
        :disabled="isLoading"
      >
        <span>👁️ 预览数据</span>
      </button>

      <button
        class="btn btn-primary"
        @click="emit('extract')"
        :disabled="isLoading || !outputOutputPath"
        :title="!outputOutputPath ? '请先选择保存位置' : ''"
      >
        <span v-if="isLoading" class="loader"></span>
        <span v-else>🚀 开始提取</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.glass-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.output-config {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
}

.config-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.config-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  width: 80px;
}

.format-select {
  flex: 1;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  outline: none;
  cursor: pointer;
}

.path-display {
  flex: 1;
  background: rgba(0, 0, 0, 0.2);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  direction: rtl;
  text-align: left;
}

.path-display.placeholder {
  color: var(--text-muted);
  direction: ltr;
}

.divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
  margin: 4px 0;
}

/* Actions */
.actions {
  display: flex;
  gap: var(--spacing-sm);
  justify-content: center;
  margin-top: var(--spacing-md);
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

.btn-sm {
  padding: 0.4rem 1rem;
  font-size: 0.85rem;
  min-width: auto;
}

.btn-primary {
  background: linear-gradient(
    135deg,
    var(--accent-primary),
    var(--accent-secondary)
  );
  color: white;
  box-shadow: 0 4px 15px
    color-mix(in srgb, var(--accent-primary) 40%, transparent);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px
    color-mix(in srgb, var(--accent-primary) 50%, transparent);
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

/* Loader */
.loader {
  width: 18px;
  height: 18px;
  border: 2px solid #fff;
  border-bottom-color: transparent;
  border-radius: 50%;
  display: inline-block;
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
</style>
