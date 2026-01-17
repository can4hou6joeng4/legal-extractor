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
  // Use the keys from the first record to determine visible columns
  return Object.keys(props.records[0]).map((key) => ({
    key,
    label: props.fieldLabels[key] || key,
    // Suggest a width based on key
    width:
      key === "defendant" ? "120px" : key === "idNumber" ? "180px" : "auto",
    fixed: key === "defendant" || key === "idNumber",
  }));
});
</script>

<template>
  <div class="preview-section glass-panel">
    <div class="preview-header">
      <h3>数据预览</h3>
      <div class="header-right">
        <span class="badge">{{ records.length }} 条记录</span>
      </div>
    </div>
    <div class="table-wrapper">
      <table>
        <colgroup>
          <col
            v-for="col in columns"
            :key="col.key"
            :style="{ width: col.width }"
          />
        </colgroup>
        <thead>
          <tr>
            <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(record, index) in records.slice(0, 50)" :key="index">
            <td v-for="col in columns" :key="col.key">
              <div
                class="cell-content"
                :class="{ 'fixed-text': col.fixed, truncate: !col.fixed }"
                :title="record[col.key]"
              >
                {{ record[col.key] }}
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

/* Scrollbar Styles */
.table-wrapper::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.table-wrapper::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
}

.table-wrapper::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.table-wrapper::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* Truncation Utilities */
.cell-content {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  max-height: 3em; /* Approximate for 2 lines */
  line-height: 1.5;
}

.cell-content.fixed-text {
  -webkit-line-clamp: 1;
  max-height: 1.5em;
  white-space: nowrap;
  display: block;
}

.cell-content.truncate {
  /* Inherits default multi-line truncation */
}
</style>
