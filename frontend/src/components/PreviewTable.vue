<script setup lang="ts">
import { computed } from "vue";

interface Record {
  [key: string]: string;
}

const props = defineProps<{
  records: Record[];
  fieldLabels: Record;
}>();

const columns = computed(() => {
  if (props.records.length === 0) return [];

  // 使用固定顺序，与后端导出保持一致
  const orderedKeys = ["defendant", "idNumber", "request", "factsReason"];

  // 找出所有在记录中出现的键
  const allKeys = new Set<string>();
  props.records.forEach(rec => {
    Object.keys(rec).forEach(k => allKeys.add(k));
  });

  return orderedKeys
    .filter(key => allKeys.has(key))
    .map((key) => ({
      key,
      label: props.fieldLabels[key] || key,
      isLongText: key === "request" || key === "factsReason",
      width: key === "defendant" ? "120px" : key === "idNumber" ? "200px" : "auto",
    }));
});
</script>

<template>
  <div class="preview-section glass-panel">
    <div class="preview-header">
      <div class="header-left">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="header-icon"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
        <h3>数据预览与编辑</h3>
      </div>
      <div class="header-right">
        <span class="hint">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="hint-icon"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
          提取结果支持直接编辑修正
        </span>
        <span class="badge">{{ records.length }} 条记录</span>
      </div>
    </div>
    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th v-for="col in columns" :key="col.key" :style="{ width: col.width }">
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(record, index) in records" :key="index">
            <td v-for="col in columns" :key="col.key">
              <div class="edit-cell">
                <textarea
                  v-if="col.isLongText"
                  v-model="records[index][col.key]"
                  rows="3"
                  class="edit-input scroll-mini"
                  spellcheck="false"
                  :aria-label="col.label + ' 输入框'"
                ></textarea>
                <input
                  v-else
                  v-model="records[index][col.key]"
                  type="text"
                  class="edit-input"
                  spellcheck="false"
                  :aria-label="col.label + ' 输入框'"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
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

.preview-section {
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 600px;
  margin-top: 20px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.2);
}

.preview-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--surface-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 0, 0, 0.2);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: var(--accent-primary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint {
  font-size: 0.75rem;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 4px;
}

.hint-icon {
  opacity: 0.7;
}

.badge {
  background: var(--accent-primary);
  color: white;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
}

.table-wrapper {
  overflow-y: auto;
  flex: 1;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  background: rgba(255, 255, 255, 0.03);
  padding: 12px;
  text-align: left;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 1px solid var(--surface-border);
}

td {
  padding: 8px;
  border-bottom: 1px solid var(--surface-border);
  vertical-align: top;
}

.edit-cell {
  width: 100%;
}

.edit-input {
  width: 100%;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 0.9rem;
  padding: 6px 8px;
  transition: all 0.2s ease;
  outline: none;
  font-family: inherit;
}

.edit-input:focus {
  background: rgba(255, 255, 255, 0.05);
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 2px rgba(var(--accent-primary-rgb), 0.2);
}

.edit-input:hover:not(:focus) {
  background: rgba(255, 255, 255, 0.03);
}

textarea.edit-input {
  resize: vertical;
  line-height: 1.5;
}

/* Mini scrollbar for textareas */
.scroll-mini::-webkit-scrollbar {
  width: 4px;
}
.scroll-mini::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}

.table-wrapper::-webkit-scrollbar {
  width: 8px;
}
.table-wrapper::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
}
.table-wrapper::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}
</style>
