/**
 * API 适配层 - 实现 Web/Desktop 双模式支持
 *
 * 根据运行环境自动选择调用方式：
 * - Desktop (Wails): 使用 Go 绑定直接调用
 * - Web (Browser): 使用 fetch API 调用 HTTP 接口
 */

// 类型定义
export interface Record {
  [key: string]: string | undefined;
  defendant?: string;
  idNumber?: string;
  request?: string;
  factsReason?: string;
}

export interface ExtractResult {
  success: boolean;
  recordCount: number;
  outputPath?: string;
  errorMessage?: string;
  records?: Record[];
  fieldLabels?: { [key: string]: string };
}

export interface FieldOption {
  key: string;
  label: string;
}

// 环境检测
export function isDesktopMode(): boolean {
  // Wails 会在 window 对象上注入 go 命名空间
  return typeof (window as any).go !== 'undefined';
}

export function isWebMode(): boolean {
  return !isDesktopMode();
}

// API 服务接口
export interface IApiService {
  // 选择文件（Desktop: 打开对话框, Web: 触发 input）
  selectFile(): Promise<string | File>;

  // 预览数据
  previewData(file: string | File, fields: string[]): Promise<ExtractResult>;

  // 提取并保存
  extractToPath(file: string | File, outputPath: string, fields: string[]): Promise<ExtractResult>;

  // 导出数据
  exportData(records: Record[], format: string): Promise<ExtractResult | Blob>;

  // 选择输出路径（仅 Desktop 模式）
  selectOutputPath(defaultName: string): Promise<string>;

  // 扫描字段
  scanFields(file: string | File): Promise<FieldOption[]>;

  // 打开文件（仅 Desktop 模式）
  openFile(path: string): Promise<void>;
}

// ============================================
// Desktop 适配器 (Wails)
// ============================================
class DesktopAdapter implements IApiService {
  private wailsApp: any;

  constructor() {
    // 动态导入 Wails 绑定
    this.wailsApp = (window as any).go?.app?.App;
  }

  async selectFile(): Promise<string> {
    const { SelectFile } = await import('../../wailsjs/go/app/App');
    return SelectFile();
  }

  async previewData(filePath: string, fields: string[]): Promise<ExtractResult> {
    const { PreviewData } = await import('../../wailsjs/go/app/App');
    return PreviewData(filePath, fields);
  }

  async extractToPath(filePath: string, outputPath: string, fields: string[]): Promise<ExtractResult> {
    const { ExtractToPath } = await import('../../wailsjs/go/app/App');
    return ExtractToPath(filePath, outputPath, fields);
  }

  async exportData(records: Record[], outputPath: string): Promise<ExtractResult> {
    const { ExportData } = await import('../../wailsjs/go/app/App');
    return ExportData(records, outputPath);
  }

  async selectOutputPath(defaultName: string): Promise<string> {
    const { SelectOutputPath } = await import('../../wailsjs/go/app/App');
    return SelectOutputPath(defaultName);
  }

  async scanFields(filePath: string): Promise<FieldOption[]> {
    const { ScanFields } = await import('../../wailsjs/go/app/App');
    return ScanFields(filePath);
  }

  async openFile(path: string): Promise<void> {
    const { OpenFile } = await import('../../wailsjs/go/app/App');
    return OpenFile(path);
  }
}

// ============================================
// Web 适配器 (HTTP API)
// ============================================
class WebAdapter implements IApiService {
  private baseUrl: string;

  constructor(baseUrl: string = '') {
    // 默认使用当前域名，或可通过环境变量配置
    this.baseUrl = baseUrl || (import.meta as any).env?.VITE_API_URL || '';
  }

  async selectFile(): Promise<File> {
    // Web 模式：创建隐藏的 input 元素触发文件选择
    return new Promise((resolve, reject) => {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = '.pdf,.docx,.jpg,.jpeg,.png';

      input.onchange = (e) => {
        const file = (e.target as HTMLInputElement).files?.[0];
        if (file) {
          resolve(file);
        } else {
          reject(new Error('未选择文件'));
        }
      };

      input.click();
    });
  }

  async previewData(file: File, fields: string[]): Promise<ExtractResult> {
    const formData = new FormData();
    formData.append('file', file);

    // 构建查询参数
    const params = new URLSearchParams();
    fields.forEach(f => params.append('fields', f));

    const response = await fetch(`${this.baseUrl}/api/extract?${params.toString()}`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json();
      return { success: false, recordCount: 0, errorMessage: error.error };
    }

    return response.json();
  }

  async extractToPath(file: File, _outputPath: string, fields: string[]): Promise<ExtractResult> {
    // Web 模式下，提取后返回数据，由前端处理导出
    return this.previewData(file, fields);
  }

  async exportData(records: Record[], format: string): Promise<Blob> {
    const response = await fetch(`${this.baseUrl}/api/export`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ records, format }),
    });

    if (!response.ok) {
      throw new Error('导出失败');
    }

    return response.blob();
  }

  async selectOutputPath(_defaultName: string): Promise<string> {
    // Web 模式不支持选择输出路径，返回空字符串
    // 导出时直接触发浏览器下载
    return '';
  }

  async scanFields(file: File): Promise<FieldOption[]> {
    // 复用 previewData 接口获取字段信息
    const result = await this.previewData(file, ['defendant', 'idNumber', 'request', 'factsReason']);

    if (!result.success || !result.fieldLabels) {
      return [];
    }

    // 转换为 FieldOption 数组
    return Object.entries(result.fieldLabels).map(([key, label]) => ({
      key,
      label: label as string,
    }));
  }

  async openFile(_path: string): Promise<void> {
    // Web 模式不支持打开本地文件
    console.warn('Web 模式不支持打开本地文件');
  }
}

// ============================================
// 工厂函数：根据环境返回对应的适配器
// ============================================
let apiServiceInstance: IApiService | null = null;

export function getApiService(): IApiService {
  if (!apiServiceInstance) {
    apiServiceInstance = isDesktopMode() ? new DesktopAdapter() : new WebAdapter();
  }
  return apiServiceInstance;
}

// 导出便捷方法
export const api = {
  get isDesktop() {
    return isDesktopMode();
  },
  get isWeb() {
    return isWebMode();
  },
  get service() {
    return getApiService();
  },
};

// 工具函数：触发浏览器下载
export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
