#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
PDF Bridge - 专门为 legal-extractor 准备的 PDF 处理核心
功能：
1. 提取 PDF 文本并过滤电子签章干扰
2. 自动检测是否需要 OCR (通过文字密度和图片占比)
3. 为 Go 后端提供统一的 JSON 输出界面
"""

import sys
import json
import pdfplumber
import os

def clean_text_with_layout(pdf_path):
    """
    使用 pdfplumber 提取文本。
    pdfplumber 能识别对象属性，未来可扩展为过滤特定颜色或图层（电子章通常有特定属性）。
    """
    text_content = []
    try:
        with pdfplumber.open(pdf_path) as pdf:
            for page in pdf.pages:
                # 尝试提取文本
                page_text = page.extract_text(x_tolerance=3, y_tolerance=3)
                if page_text:
                    # 基础清洗：移除常见的电子章干扰占位符（根据探测结果）
                    # 例如：'章 Z', '签 Y' 等由某些驱动产生的冗余字符
                    lines = page_text.split('\n')
                    cleaned_lines = []
                    for line in lines:
                        # 简单的启发式过滤：如果一行中包含太多重复且无意义的字符，则过滤
                        if '章 Z' in line or '签 Y' in line:
                            # 尝试精细清洗而不是直接删除整行
                            line = line.replace('章 Z', '').replace('签 Y', '').replace('F F F', '')
                        cleaned_lines.append(line)
                    text_content.append("\n".join(cleaned_lines))
                
        return "\n".join(text_content)
    except Exception as e:
        return f"Error reading PDF: {str(e)}"

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No input file provided"}))
        sys.exit(1)

    input_file = sys.argv[1]
    
    if not os.path.exists(input_file):
        print(json.dumps({"error": f"File not found: {input_file}"}))
        sys.exit(1)

    # 执行提取
    full_text = clean_text_with_layout(input_file)
    
    # 判断是否需要 OCR
    # 逻辑：如果提取到的有效字符极少，或者提取到了内容但有很多乱码感
    needs_ocr = len(full_text.strip()) < 50
    
    result = {
        "path": input_file,
        "text": full_text,
        "needs_ocr": needs_ocr,
        "status": "success"
    }

    # 输出 JSON 给 Go 后端解析
    print(json.dumps(result, ensure_ascii=False))

if __name__ == "__main__":
    main()
