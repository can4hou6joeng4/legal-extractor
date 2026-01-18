<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { SelectOutputPath, GetSupportedFields } from "../../wailsjs/go/app/App";

const props = defineProps<{
  selectedFile: string;
  fileName: string;
  selectedFormat: "xlsx" | "csv" | "json";
  outputOutputPath: string;
  isLoading: boolean;
  selectedFields: string[];
}>();

const emit = defineEmits<{
  (e: "update:selectedFormat", value: string): void;
  (e: "update:outputOutputPath", value: string): void;
  (e: "update:selectedFields", value: string[]): void;
  (e: "preview"): void;
  (e: "extract"): void;
}>();

const availableFields = ref<any[]>([]);

onMounted(async () => {
  try {
    const fields = await GetSupportedFields();
    availableFields.value = fields;
    // Default to all selected if none provided
    if (props.selectedFields.length === 0) {
      emit(
        "update:selectedFields",
        fields.map((f: any) => f.key),
      );
    }
  } catch (e) {
    console.error("Failed to fetch fields:", e);
  }
});

function toggleField(key: string) {
  const newFields = [...props.selectedFields];
  const index = newFields.indexOf(key);
  if (index > -1) {
    if (newFields.length > 1) {
      // Keep at least one
      newFields.splice(index, 1);
    }
  } else {
    newFields.push(key);
  }
  emit("update:selectedFields", newFields);
}

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
    <!-- Output Configuration -->
    <div class="output-config glass-panel">
      <!-- Field Selection -->
      <div class="config-section">
        <span class="config-label">提取字段：</span>
        <div class="field-list">
          <label 
            v-for="field in availableFields" 
            :key="field.key" 
            class="field-item"
            :class="{ active: selectedFields.includes(field.key) }"
          >
            <input 
              type="checkbox" 
              :checked="selectedFields.includes(field.key)"
              @change="toggleField(field.key)"
            >
            <span class="checkbox-custom"></span>
            <span class="field-label">{{ field.label }}</span>
          </label>
        </div>
      </div>

      <div class="divider"></div>

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
        <span class="btn-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
        </span>
        <span>预览数据</span>
      </button>

      <button
        class="btn btn-primary"
        @click="emit('extract')"
        :disabled="isLoading || !outputOutputPath"
        :title="!outputOutputPath ? '请先选择保存位置' : ''"
      >
        <span v-if="isLoading" class="loader"></span>
        <span v-else class="btn-content">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m4.5 16.5 4-4c2-2 2.83-2.83 4-1.17 1.17 1.66 2 1.66 4 0L21 7.5"/><path d="M21 20.66 4.5 4.16"/><path d="M16 4h5v5"/></svg>
            <span>开始提取</span>
        </span>
      </button>
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

.config-row, .config-section {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
}

.config-section {
  flex-direction: column;
}

.config-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  width: 80px;
  margin-top: 4px;
}

.field-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 4px 0;
}

.field-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  transition: all 0.2s ease;
  user-select: none;
}

.field-item:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.field-item.active {
  background: rgba(64, 158, 255, 0.15);
  border-color: var(--accent-primary);
}

.field-item input {
  display: none;
}

.checkbox-custom {
  width: 16px;
  height: 16px;
  border: 1.5px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  position: relative;
  transition: all 0.2s ease;
}

.field-item.active .checkbox-custom {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
}

.checkbox-custom::after {
  content: "✓";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 10px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.field-item.active .checkbox-custom::after {
  opacity: 1;
}

.field-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.field-item.active .field-label {
  color: var(--text-primary);
  font-weight: 500;
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
  gap: 8px; /* Added gap */
  padding: 0.8rem 1.8rem;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 140px;
}

.btn-content, .btn-icon {
    display: flex;
    align-items: center;
    gap: 6px;
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
