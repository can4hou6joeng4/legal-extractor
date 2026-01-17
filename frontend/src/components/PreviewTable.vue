<script setup lang="ts">
interface Record {
  defendant: string;
  idNumber: string;
  request: string;
  factsReason: string;
}

const props = defineProps<{
  records: Record[];
}>();
</script>

<template>
  <div class="preview-section glass-panel">
    <div class="preview-header">
      <h3>数据预览</h3>
      <span class="badge">{{ records.length }} 条记录</span>
    </div>
    <div class="table-wrapper">
      <table>
        <colgroup>
          <col style="width: 100px" />
          <col style="width: 180px" />
          <col style="width: auto" />
          <col style="width: auto" />
        </colgroup>
        <thead>
          <tr>
            <th>被告</th>
            <th>身份证号码</th>
            <th>诉讼请求</th>
            <th>事实与理由</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(record, index) in records.slice(0, 50)" :key="index">
            <td>
              <div class="cell-content fixed-text" :title="record.defendant">
                {{ record.defendant }}
              </div>
            </td>
            <td>
              <div class="cell-content fixed-text" :title="record.idNumber">
                {{ record.idNumber }}
              </div>
            </td>
            <td>
              <div class="cell-content truncate" :title="record.request">
                {{ record.request }}
              </div>
            </td>
            <td>
              <div class="cell-content truncate" :title="record.factsReason">
                {{ record.factsReason }}
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
