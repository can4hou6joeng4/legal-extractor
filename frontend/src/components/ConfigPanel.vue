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

// Icon mapping for fields
function getFieldIcon(key: string) {
  const k = key.toLowerCase();
  if (k.includes('defendant') || k.includes('name') || k.includes('被告')) return 'user';
  if (k.includes('id') || k.includes('shenfen') || k.includes('身份证')) return 'card';
  if (k.includes('request') || k.includes('claim') || k.includes('请求')) return 'gavel';
  if (k.includes('fact') || k.includes('reason') || k.includes('事实')) return 'file-text';
  return 'tag';
}

// Watch for file changes
watch(
  () => props.selectedFile,
  async (newFile) => {
    if (!newFile) {
      availableFields.value = [];
      return;
    }
    
    isScanning.value = true;
    availableFields.value = []; 
    
    try {
      const fields = await ScanFields(newFile);
      availableFields.value = fields || [];
      
      if (fields && fields.length > 0) {
        emit("update:selectedFields", fields.map((f: any) => f.key));
      } else {
        emit("update:selectedFields", []);
      }
    } catch (e) {
      console.error("Scan failed:", e);
    } finally {
      setTimeout(() => {
        isScanning.value = false;
      }, 600);
    }
  },
  { immediate: true }
);

function toggleField(key: string) {
  const newFields = [...props.selectedFields];
  const index = newFields.indexOf(key);
  if (index > -1) {
    if (newFields.length > 1) {
      newFields.splice(index, 1);
    }
  } else {
    newFields.push(key);
  }
  emit("update:selectedFields", newFields);
}

async function handleSelectOutput() {
  if (!props.selectedFile) return;

  const ext = props.selectedFormat;
  const baseName =
    (props.fileName || "document.doc").replace(/\.[^/.]+$/, "") + "." + ext;

  try {
    const path = await SelectOutputPath(baseName);
    if (path) {
      emit("update:outputOutputPath", path);
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
    <div class="output-config glass-panel">
      <!-- Field Selection -->
      <div class="config-section">
        <div class="section-header">
           <div class="header-left">
               <span class="config-label">提取字段</span>
               <span v-if="!isScanning && availableFields.length > 0" class="badge">
                 {{ availableFields.length }}
               </span>
           </div>
           
           <span v-if="isScanning" class="status-text blink">
               <span class="loader-mini"></span> 分析结构中...
           </span>
        </div>
        
        <!-- Skeleton Loader -->
        <div v-if="isScanning" class="field-grid skeleton-grid">
           <div class="skeleton-chip" v-for="i in 4" :key="i"></div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!selectedFile" class="empty-state">
           <span class="empty-icon">
             <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 2H4a2 2 0 0 0-2 2v13.61A2.39 2.39 0 0 0 4 20Z"/></svg>
           </span>
           <span>请先选择文件以分析可提取字段</span>
        </div>

        <div v-else-if="availableFields.length === 0" class="empty-state warning">
           <span class="warning-icon">
             <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
           </span>
           <span>未检测到可提取的字段</span>
        </div>

        <!-- Field Grid -->
        <div v-else class="field-grid">
          <label 
            v-for="field in availableFields" 
            :key="field.key" 
            class="field-card"
            :class="{ active: selectedFields.includes(field.key) }"
          >
            <input 
              type="checkbox" 
              :checked="selectedFields.includes(field.key)"
              @change="toggleField(field.key)"
            >
            <div class="card-content">
                <div class="icon-box">
                    <!-- Dynamic Icons -->
                    <svg v-if="getFieldIcon(field.key) === 'user'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                    <svg v-else-if="getFieldIcon(field.key) === 'card'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="3" rx="2" ry="2"/><path d="M7 7h3"/><path d="M7 12h8"/><path d="M7 17h5"/></svg>
                    <svg v-else-if="getFieldIcon(field.key) === 'gavel'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m14 13-7.5 7.5c-.83.83-2.17.83-3 0 0 0 0 0 0 0a2.12 2.12 0 0 1 0-3L11 10"/><path d="m16 16 6-6"/><path d="m8 8 6-6"/><path d="m9 7 8 8"/><path d="m21 11-8-8"/></svg>
                    <svg v-else-if="getFieldIcon(field.key) === 'file-text'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z"/><path d="M7 7h.01"/></svg>
                </div>
                <span class="field-label">{{ field.label }}</span>
                <div class="check-mark">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                </div>
            </div>
          </label>
        </div>
      </div>

      <div class="divider"></div>

      <div class="grid-row">
        <!-- Export Format -->
        <div class="config-cell">
           <label class="cell-label">导出格式</label>
           <div class="select-wrapper">
             <select
               :value="selectedFormat"
               @input="emit('update:selectedFormat', ($event.target as HTMLSelectElement).value)"
               class="custom-select"
             >
               <option value="xlsx">Excel 表格 (.xlsx)</option>
               <option value="csv">CSV 文件 (.csv)</option>
               <option value="json">JSON 数据 (.json)</option>
             </select>
             <div class="select-arrow">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>
             </div>
           </div>
        </div>
        
        <!-- Export Path -->
        <div class="config-cell flex-grow">
           <label class="cell-label">导出位置</label>
           <div class="path-input-group">
               <div class="path-display" :class="{ placeholder: !outputOutputPath }" :title="outputOutputPath">
                 <span class="path-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
                 </span>
                 <span class="path-text">{{ outputOutputPath || "请选择保存位置..." }}</span>
               </div>
               <button class="btn-icon-only" @click="handleSelectOutput" title="更改位置">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
               </button>
           </div>
        </div>
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
        class="btn btn-primary btn-glow"
        @click="emit('extract')"
        :disabled="isLoading || !outputOutputPath"
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
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.output-config {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  padding: 24px;
  border-radius: 16px;
}

/* Header */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.config-label, .cell-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.badge {
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-primary);
}

.status-text {
  font-size: 0.85rem;
  color: var(--accent-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.loader-mini {
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-bottom-color: transparent;
  border-radius: 50%;
  animation: rotation 1s linear infinite;
}

/* Field Grid */
.field-grid {
  display: grid;
  /* Optimized for 4 items: use larger min-width to force 2 columns layout */
  /* 280px ensures 2 columns on typical widths (800px container), avoiding the 3+1 orphan layout */
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  align-items: stretch; /* Ensure all items in a row have equal height */
}

.field-card {
  position: relative;
  cursor: pointer;
  user-select: none;
  height: 100%; /* Fill grid cell height */
}

.field-card input {
  display: none;
}

.card-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  height: 100%; /* Fill card height */
}

.icon-box {
  width: 36px; /* Slightly larger icon box */
  height: 36px;
  flex-shrink: 0; /* Prevent shrinking */
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  color: var(--text-muted);
  transition: all 0.3s ease;
}

.field-label {
  font-size: 0.95rem;
  color: var(--text-secondary);
  flex: 1;
  font-weight: 500;
  line-height: 1.4;
  word-break: break-all; /* Handle long words nicely */
}

.check-mark {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  color: transparent;
  transition: all 0.3s ease;
}

/* Active State */
.field-card.active .card-content {
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.15), rgba(64, 158, 255, 0.05));
  border-color: rgba(64, 158, 255, 0.4);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
}

.field-card.active .icon-box {
  background: var(--accent-primary);
  color: white;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}

.field-card.active .field-label {
  color: var(--text-primary);
}

.field-card.active .check-mark {
  background: var(--accent-primary);
  color: white;
}

.field-card:hover .card-content {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
}

/* Grid Layout for config */
.grid-row {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 20px;
  align-items: end;
}

.config-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0; /* Critical: allows grid item to shrink below content size */
}

/* Custom Select */
.select-wrapper {
  position: relative;
  height: 44px; /* Match path input height */
}

.custom-select {
  width: 100%;
  height: 100%;
  appearance: none;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
  padding: 0 40px 0 16px; /* Adjust padding for height centering */
  border-radius: 10px;
  font-size: 0.95rem;
  outline: none;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
}

.custom-select:hover, .custom-select:focus {
  border-color: var(--accent-primary);
  background: rgba(0, 0, 0, 0.3);
}

.select-arrow {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  pointer-events: none;
}

/* Path Input */
.path-input-group {
  display: flex;
  gap: 12px;
  align-items: center;
}

.path-display {
  flex: 1;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 0 16px;
  height: 44px; /* Fixed height */
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.9rem;
  color: var(--text-primary);
  overflow: hidden;
  transition: all 0.2s ease;
}

.path-display:hover {
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(0, 0, 0, 0.3);
}

.path-display.placeholder .path-text {
  color: var(--text-muted);
  font-style: italic;
}

.path-text {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1; /* Fix vertical alignment */
  transform: translateY(1px); /* Optical correction for monospace font */
  flex: 1; /* Take remaining space */
  min-width: 0; /* Allow shrinking below content size */
}

.path-icon {
  opacity: 0.7;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.btn-icon-only {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.btn-icon-only:hover {
  background: rgba(255, 255, 255, 0.15);
  border-color: var(--text-muted);
  transform: translateY(-1px);
}

/* Skeleton */
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.skeleton-chip {
  height: 56px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  animation: pulse 1.5s infinite;
}

.divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  margin: 16px 0;
}

/* Actions */
.actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin-top: 32px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0 32px; /* Horizontal padding only */
  height: 48px; /* Fixed height for perfect centering */
  border-radius: 12px;
  font-weight: 600;
  font-size: 1rem;
  letter-spacing: 0.5px;
  cursor: pointer;
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 160px; /* Slightly wider */
}

.btn-content {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 100%;
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.btn-primary.btn-glow {
  background: linear-gradient(135deg, var(--color-cta), var(--color-cta-hover));
  box-shadow: 0 4px 20px rgba(245, 158, 11, 0.3);
  position: relative;
  overflow: hidden;
  color: #fff;
}

.btn-primary.btn-glow:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(245, 158, 11, 0.4);
}

.btn:disabled,
.btn-primary.btn-glow:disabled {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.4);
  box-shadow: none;
  cursor: not-allowed;
  transform: none;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-primary.btn-glow:disabled::after {
  display: none;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  gap: 12px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  border: 1px dashed rgba(255, 255, 255, 0.1);
}

.empty-icon {
  font-size: 2rem;
  opacity: 0.5;
}

@keyframes pulse {
  0% { opacity: 0.3; }
  50% { opacity: 0.6; }
  100% { opacity: 0.3; }
}

@keyframes rotation {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
