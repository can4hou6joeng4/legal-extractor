<script setup lang="ts">
import { ref, watch } from "vue";
import { SelectOutputPath, ScanFields } from "../../wailsjs/go/app/App";

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
const isScanning = ref(false);

// Watch for file changes to trigger scan
watch(
  () => props.selectedFile,
  async (newFile) => {
    if (!newFile) {
      availableFields.value = [];
      return;
    }
    
    isScanning.value = true;
    availableFields.value = []; // Clear previous
    
    try {
      const fields = await ScanFields(newFile);
      availableFields.value = fields || [];
      
      // Auto-select all found fields initially
      // In a real app we might want to preserve user choice if re-scanning same file type,
      // but for now, fresh scan = fresh selection is cleaner.
      if (fields && fields.length > 0) {
        emit("update:selectedFields", fields.map((f: any) => f.key));
      } else {
        emit("update:selectedFields", []);
      }
    } catch (e) {
      console.error("Scan failed:", e);
      // Fallback or error state could be handled here
    } finally {
      // Small delay for visual consistency so skeleton doesn't flash too fast
      setTimeout(() => {
        isScanning.value = false;
      }, 600);
    }
  },
  { immediate: true } // Trigger on mount if file already selected
);

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
        <div class="section-header">
           <span class="config-label">提取字段</span>
           <span v-if="isScanning" class="status-text blink">正在分析文档结构...</span>
           <span v-else-if="availableFields.length > 0" class="status-text">
             已识别 {{ availableFields.length }} 个字段
           </span>
        </div>
        
        <!-- Skeleton Loader -->
        <div v-if="isScanning" class="field-list skeleton-list">
           <div class="skeleton-chip" style="width: 80px"></div>
           <div class="skeleton-chip" style="width: 120px"></div>
           <div class="skeleton-chip" style="width: 100px"></div>
           <div class="skeleton-chip" style="width: 90px"></div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!selectedFile" class="empty-state">
           <span class="empty-icon">📂</span>
           <span>请先选择文件以分析可提取字段</span>
        </div>
        
        <div v-else-if="availableFields.length === 0" class="empty-state warning">
           <span>⚠️ 未检测到可提取的字段</span>
        </div>

        <!-- Field List -->
        <div v-else class="field-list">
          <label 
            v-for="field in availableFields" 
            :key="field.key" 
            class="field-chip"
            :class="{ active: selectedFields.includes(field.key) }"
          >
            <input 
              type="checkbox" 
              :checked="selectedFields.includes(field.key)"
              @change="toggleField(field.key)"
            >
            <div class="chip-content">
                <div class="check-icon-wrapper">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" class="check-icon"><polyline points="20 6 9 17 4 12"></polyline></svg>
                </div>
                <span class="field-label">{{ field.label }}</span>
            </div>
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


/* Section Header */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.status-text {
  font-size: 0.8rem;
  color: var(--accent-primary);
}

.status-text.blink {
  animation: pulse-text 1.5s infinite;
}

/* Skeleton Loader */
.skeleton-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 4px 0;
}

.skeleton-chip {
  height: 32px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { opacity: 0.3; }
  50% { opacity: 0.6; }
  100% { opacity: 0.3; }
}

@keyframes pulse-text {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}

/* Empty State */
.empty-state {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: rgba(0,0,0,0.1);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-size: 0.9rem;
  font-style: italic;
}

.empty-state.warning {
  color: #ff9f43;
  background: rgba(255, 159, 67, 0.1);
}

/* Field Chips (New Design) */
.field-chip {
  position: relative;
  cursor: pointer;
  user-select: none;
}

.field-chip input {
  display: none;
}

.chip-content {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: rgba(255, 255, 255, 0.05); /* Glassmorphism base */
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.field-chip.active .chip-content {
  background: rgba(64, 158, 255, 0.2); /* Accent color */
  border-color: var(--accent-primary);
  box-shadow: 0 0 10px rgba(64, 158, 255, 0.15);
}

.field-chip:hover .chip-content {
  background: rgba(255, 255, 255, 0.1);
}

.field-chip.active:hover .chip-content {
  background: rgba(64, 158, 255, 0.25);
}

.check-icon-wrapper {
  width: 0;
  overflow: hidden;
  transition: width 0.3s ease, margin-right 0.3s ease;
  display: flex;
  align-items: center;
  margin-right: -4px; /* Offset for hidden state */
}

.field-chip.active .check-icon-wrapper {
  width: 14px;
  margin-right: 4px;
}

.check-icon {
  stroke: var(--accent-primary);
  transform: scale(0);
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1); 
}

.field-chip.active .check-icon {
  transform: scale(1);
}

.field-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
  transition: color 0.3s ease;
}

.field-chip.active .field-label {
  color: white;
  font-weight: 500;
}

.divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
  margin: 12px 0;
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
