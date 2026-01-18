<script setup lang="ts">
import { OpenFile } from "../../wailsjs/go/app/App";

interface Record {
  defendant: string;
  idNumber: string;
  request: string;
  factsReason: string;
}

interface ExtractResult {
  success: boolean;
  recordCount: number;
  outputPath: string;
  errorMessage?: string;
  records?: Record[];
}

const props = defineProps<{
  result: ExtractResult | null;
}>();

const emit = defineEmits<{
  (
    e: "notification",
    message: string,
    type: "success" | "error" | "info"
  ): void;
}>();

async function handleOpenFile(path: string) {
  if (!path) return;
  try {
    await OpenFile(path);
  } catch (e) {
    console.error("Failed to open file:", e);
    emit("notification", "无法打开文件", "error");
  }
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="result"
      class="result-card glass-panel"
      :class="{ error: !result.success }"
    >
      <div class="result-header">
        <span class="status-icon">
            <svg v-if="result.success" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="M22 4 12 14.01l-3-3"/></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
        </span>
        <h3 class="font-heading">{{ result.success ? "提取成功" : "提取失败" }}</h3>
      </div>

      <div v-if="result.success" class="result-body">
        <div class="stat-item">
          <span class="label">提取记录</span>
          <span class="value">{{ result.recordCount }}</span>
        </div>
        <div class="path-box">
          <span class="label">保存至：</span>
          <code
            @click="handleOpenFile(result.outputPath)"
            class="clickable-path"
            title="点击打开文件"
            >{{ result.outputPath }}</code
          >
        </div>
      </div>
      <div v-else class="result-body">
        <p class="error-msg">{{ result.errorMessage }}</p>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.glass-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* Result Card */
.result-card {
  border-radius: var(--radius-lg);
  padding: var(--spacing-md);
  border-left: 4px solid var(--success);
}

.result-card.error {
  border-left-color: var(--error);
}

.result-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.status-icon {
  font-size: 1.5rem;
}

.result-body {
  margin-left: calc(1.5rem + var(--spacing-sm));
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xs);
}

.stat-item .value {
  font-weight: 700;
  font-size: 1.2rem;
  color: var(--success);
}

.path-box {
  background: rgba(0, 0, 0, 0.2);
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  margin-top: var(--spacing-sm);
  font-family: monospace;
  font-size: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.path-box code {
  color: var(--accent-primary);
  word-break: break-all;
}

.clickable-path {
  cursor: pointer;
  text-decoration: underline;
  transition: opacity 0.2s;
}

.clickable-path:hover {
  opacity: 0.8;
}

.error-msg {
  color: var(--error);
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
