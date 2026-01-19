#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
PDF Bridge - 法律文书智能提取器
功能：
1. 使用 pdfplumber 进行高精度文本提取
2. 智能过滤电子签章等干扰信息
3. 直接提取法律文书的关键字段（被告、身份证号码、诉讼请求、事实与理由）
4. 为 Go 后端提供结构化 JSON 输出
"""

import sys
import json
import pdfplumber
import os
import re

try:
    from rapidocr_onnxruntime import RapidOCR
    import fitz  # PyMuPDF
    OCR_AVAILABLE = True
except ImportError:
    OCR_AVAILABLE = False

class OCRExtractor:
    def __init__(self):
        if not OCR_AVAILABLE:
            raise ImportError("OCR libraries not installed")
        self.ocr = RapidOCR()
        
    def extract(self, pdf_path):
        if not os.path.exists(pdf_path):
            raise FileNotFoundError(f"{pdf_path} not found")
            
        doc = fitz.open(pdf_path)
        full_text = []
        
        for i, page in enumerate(doc):
            # Render page to image
            # zoom=2 for better resolution (approx 144 dpi -> 288 dpi)
            pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
            img_bytes = pix.tobytes("png")
            
            # OCR recognition
            # result format: [[[[x1,y1], [x2,y2], [x3,y3], [x4,y4]], text, confidence]]
            result, _ = self.ocr(img_bytes)
            
            page_content = []
            if result:
                # Custom sort: bucket Y by 20 pixels (line height approx), then X
                # This helps reconstruct lines from scattered text boxes
                boxes = []
                for item in result:
                    box, text, conf = item
                    # Calculate center Y
                    y_center = sum(p[1] for p in box) / 4
                    x_min = min(p[0] for p in box)
                    boxes.append({"text": text, "y": y_center, "x": x_min})
                
                # Sort criteria: 
                # 1. Y coordinate (bucketed to handle slight misalignment)
                # 2. X coordinate (left to right)
                boxes.sort(key=lambda b: (int(b['y'] / 20), b['x']))
                
                for box in boxes:
                    page_content.append(box['text'])
            
            full_text.append('\n'.join(page_content))
            
        doc.close()
        return '\n\n'.join(full_text)


def is_seal_like_text(text_line):
    """
    检测是否是电子章/水印相关的干扰文本
    增强版：识别多种水印特征模式
    """
    if not text_line or len(text_line.strip()) == 0:
        return True
    
    # 基础关键词检测
    seal_keywords = ['章 Z', '签 Y', 'F F F', '印章', '子 C', '电 B']
    if any(keyword in text_line for keyword in seal_keywords):
        return True
    
    # 检测重复字符（3个或更多相同字符）
    if re.search(r'(.)\1{2,}', text_line):
        return True
    
    # 检测连续的大写字母（3个或更多）- 水印常见模式
    if re.search(r'[A-Z]{3,}', text_line):
        return True
    
    # 检测单行只有数字和字母的情况（可能是水印编号）
    if re.match(r'^[0-9A-Za-z\s]+$', text_line) and len(text_line.strip()) < 20:
        return True
    
    # 检测重复词组（如"银 O 银 O"）
    if re.search(r'(.{2,})\s*\1\s*\1', text_line):
        return True
    
    # 检测行开头有特殊重复模式
    if re.match(r'^([A-Z\d]{1,3}\s*){3,}', text_line):
        return True
    
    return False


def clean_watermark_chars(text):
    """
    清理文本中的水印字符
    处理如 "码码码" 这样重复出现的单字
    """
    # 移除重复的单字（连续2次以上）
    text = re.sub(r'([一-龥])\1+', r'\1', text)
    # 移除一些常见的水印词组
    watermark_patterns = [
        r'招招招', r'证证证', r'验验验', r'银银银',
        r'码码码', r'商商商', r'行行行'
    ]
    for pattern in watermark_patterns:
        text = text.replace(pattern, '')
    return text

def extract_text_only(pdf_path):
    """
    使用布局分析提取文本，智能过滤干扰 (纯文本模式)
    """
    extracted_pages = []
    total_chars = 0  # 统计总字符数
    
    with pdfplumber.open(pdf_path) as pdf:
        for page_num, page in enumerate(pdf.pages):
            # ... (这部分逻辑保持不变，只需改函数名)
            # 方法1: 基于坐标的字符级提取（更精确）
            chars = page.chars
            total_chars += len(chars)
            
            if len(chars) > 0:
                lines = {}
                for char in chars:
                    y = round(char['top'], 1)
                    if y not in lines:
                        lines[y] = []
                    lines[y].append(char)
                
                sorted_lines = []
                for y in sorted(lines.keys()):
                    line_chars = sorted(lines[y], key=lambda c: c['x0'])
                    line_text = ''.join(c['text'] for c in line_chars)
                    
                    if not is_seal_like_text(line_text):
                        cleaned_line = clean_watermark_chars(line_text)
                        if cleaned_line.strip():
                            sorted_lines.append(cleaned_line)
                
                page_text = '\n'.join(sorted_lines)
            else:
                page_text = page.extract_text(x_tolerance=3, y_tolerance=3, layout=True)
            
            if page_text:
                extracted_pages.append(page_text)
    
    full_text = '\n\n'.join(extracted_pages)
    
    # 标记是否可能是扫描件，供调用者判断
    if total_chars < 50: # 稍微提高阈值，只要字符极少就认为是扫描件
        return full_text, True
    
    return full_text, False

def extract_smart(pdf_path):
    """
    智能提取：优先文本提取，若可能是扫描件则切换到 OCR
    注意：此函数现在主要用于判断是否需要 OCR 以及获取 OCR 内容，
    具体的字段合并逻辑上移到 main 函数处理
    """
    text, is_scanned_likely = extract_text_only(pdf_path)
    ocr_text = None
    
    # 策略增强：检查关键字段完整性
    if not is_scanned_likely and OCR_AVAILABLE:
        # 如果有"身份证"标签但无法提取出号码，建议开启 OCR
        if "身份证" in text and extract_id_number(text) == "":
            is_scanned_likely = True
    
    if is_scanned_likely and OCR_AVAILABLE:
        try:
            ocr = OCRExtractor()
            ocr_text = ocr.extract(pdf_path)
        except Exception:
            pass
            
    return text, ocr_text

def merge_records(text_records, ocr_records):
    """
    智能合并 Text 模式和 OCR 模式提取的记录
    原则：
    1. Defendant, Request, FactsReason: 优先信赖 Text (排版更好)
    2. ID Number: 优先信赖 OCR (专门解决图片 ID 问题)
    """
    if not ocr_records:
        return text_records
    
    if not text_records:
        return ocr_records
        
    # 简单场景：如果不涉及到多案件分割，或者分割数量一致
    if len(text_records) == len(ocr_records):
        merged = []
        for i in range(len(text_records)):
            t_rec = text_records[i]
            o_rec = ocr_records[i]
            
            new_rec = t_rec.copy()
            
            # 1. 身份证号：OCR 优先
            if o_rec.get("idNumber"):
                new_rec["idNumber"] = o_rec["idNumber"]
            
            # 2. 其他字段：如果 Text 缺失，尝试用 OCR 补全
            if not new_rec.get("defendant") and o_rec.get("defendant"):
                new_rec["defendant"] = o_rec["defendant"]
                
             # 诉讼请求和事实理由通常 Text 提取的更好（即使有部分乱码），
             # 除非 Text 完全为空
            if not new_rec.get("request") and o_rec.get("request"):
                 new_rec["request"] = o_rec["request"]
                 
            if not new_rec.get("factsReason") and o_rec.get("factsReason"):
                 new_rec["factsReason"] = o_rec["factsReason"]
                 
            merged.append(new_rec)
        return merged
        
    # 复杂场景：分割数量不一致
    # 此时通常 Text 分割更准（OCR 可能把分界词识别错了）
    # 我们尝试根据"被告"姓名进行匹配合并
    # (简化处理：以 Text 为主，尝试将 OCR 里的 ID 填入匹配的 Text 记录)
    for t_rec in text_records:
        t_name = t_rec.get("defendant", "")
        for o_rec in ocr_records:
            o_name = o_rec.get("defendant", "")
            # 如果姓名相似或为空
            if t_name == o_name or t_name in o_name or o_name in t_name:
                if not t_rec.get("idNumber") and o_rec.get("idNumber"):
                    t_rec["idNumber"] = o_rec["idNumber"]
                break
    
    return text_records

def smart_merge(text):
    """
    智能合并换行符
    保留句号、分号、冒号后的换行，或者新条目序号之前的换行
    """
    if not text:
        return ""
    
    text = text.strip()
    
    # 1. 标准化换行符
    text = text.replace('\r\n', '\n')
    text = re.sub(r'\n+', '\n', text)
    
    # 2. 标记需要保留的"逻辑断点"
    # A. 句末标点后：。；？！
    text = re.sub(r'([。；？！])\n', r'\1[LOGICAL_NL]', text)
    
    # B. 条目序号前：\n一、 \n(1) 等
    text = re.sub(r'\n(\s*(?:[一二三四五六七八九十\d]+[、．]|[(（][一二三四五六七八九十\d]+[)）]))', r'[LOGICAL_NL]\1', text)
    
    # 3. 将剩余的所有普通换行符替换为空（彻底合并）
    text = text.replace('\n', '')
    
    # 4. 将占位符还原为真正的换行
    text = text.replace('[LOGICAL_NL]', '\n')
    
    # 5. 深度清理：合并每行内部的多余空格
    lines = text.split('\n')
    result_lines = []
    for line in lines:
        trimmed = line.strip()
        if trimmed:
            # 移除行内多余空格
            cleaned = ''.join(trimmed.split())
            result_lines.append(cleaned)
    
    return '\n'.join(result_lines)

def extract_defendant(text):
    """
    提取被告姓名
    """
    # 匹配 "被告:" 或 "被 告："（允许空格）
    pattern_start = re.compile(r'被\s*告\s*[:：]')
    match_start = pattern_start.search(text)
    
    if match_start:
        start_idx = match_start.end()
        remaining = text[start_idx:]
        
        # 先移除所有换行和多余空格，获取连续文本
        clean_remaining = remaining.replace('\n', '').replace('\r', '')
        
        # 查找结束标记（性别、生日、身份证、住址等）
        pattern_end = re.compile(r'[,，、；;、\s]*(?:性\s*别|生\s*日|身\s*份\s*证|住\s*址|联\s*系\s*电\s*话|现\s*住|案\s*由)|[。]|$')
        match_end = pattern_end.search(clean_remaining)
        
        if match_end:
            name = clean_remaining[:match_end.start()]
        else:
            # 如果没找到结束标记，取前50个字符
            name = clean_remaining[:50] if len(clean_remaining) > 50 else clean_remaining
            # 尝试找到第一个非姓名字符
            for i, char in enumerate(name):
                if char in '性男女生住联':
                    name = name[:i]
                    break
        
        # 清洗姓名
        name = name.strip(' ,，、:：；;\t')
        name = name.lstrip('被告')
        return name.strip()
    
    # 回退方案：简单匹配
    pattern_fallback = re.compile(r'被\s*告\s*[:：]\s*(.*?)\n')
    match_fallback = pattern_fallback.search(text)
    if match_fallback:
        return match_fallback.group(1).strip()
    
    return ""

def extract_id_number(text):
    """
    提取身份证号码
    """
    pattern = re.compile(r'身\s*份\s*证\s*号\s*码\s*[:：]\s*([\dXx]{15,18})')
    match = pattern.search(text)
    if match:
        return match.group(1).strip()
    
    # 额外尝试：直接匹配18位身份证号
    pattern_direct = re.compile(r'\b(\d{17}[\dXx])\b')
    match_direct = pattern_direct.search(text)
    if match_direct:
        return match_direct.group(1)
    
    return ""

def extract_request(text):
    """
    提取诉讼请求
    """
    pattern = re.compile(r'(?s)诉\s*讼\s*请\s*求\s*[:：]\s*(.*?)\s*事\s*实\s*与\s*理\s*由', re.DOTALL)
    match = pattern.search(text)
    if match:
        content = match.group(1)
        return smart_merge(content)
    return ""

def extract_facts_reason(text):
    """
    提取事实与理由
    """
    pattern = re.compile(r'(?s)事\s*实\s*与\s*理\s*由\s*[:：]\s*(.*?)\s*此\s*致', re.DOTALL)
    match = pattern.search(text)
    if match:
        content = match.group(1)
        return smart_merge(content)
    
    # 回退方案：只找开始，不找结束
    pattern_fallback = re.compile(r'(?s)事\s*实\s*与\s*理\s*由\s*[:：]\s*(.*)', re.DOTALL)
    match_fallback = pattern_fallback.search(text)
    if match_fallback:
        content = match_fallback.group(1)
        # 限制长度避免包含过多内容
        if len(content) > 2000:
            content = content[:2000]
        return smart_merge(content)
    
    return ""

def extract_all_fields(text):
    """
    从文本中提取所有字段
    返回结构化数据
    """
    # 分割多个案例（基于"民事起诉状"）
    cases = re.split(r'民\s*事\s*起\s*诉\s*状', text)
    
    records = []
    
    for case_text in cases:
        if not case_text.strip():
            continue
        
        record = {
            "defendant": extract_defendant(case_text),
            "idNumber": extract_id_number(case_text),
            "request": extract_request(case_text),
            "factsReason": extract_facts_reason(case_text)
        }
        
        # 只有至少一个字段有值时才添加记录
        if any(v for v in record.values()):
            records.append(record)
    
    return records

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No input file provided", "status": "failed"}))
        sys.exit(1)

    input_file = sys.argv[1]
    
    if not os.path.exists(input_file):
        print(json.dumps({"error": f"File not found: {input_file}", "status": "failed"}))
        sys.exit(1)

    try:
        # 1. 获取 Text 和 OCR 文本
        text_content, ocr_content = extract_smart(input_file)
        
        # 2. 分别提取字段
        text_records = extract_all_fields(text_content)
        
        final_records = text_records
        
        # 3. 如果有 OCR 内容，进行合并
        if ocr_content:
            ocr_records = extract_all_fields(ocr_content)
            final_records = merge_records(text_records, ocr_records)
        
        # 4. 构建结果
        result = {
            "path": input_file,
            "records": final_records,
            "count": len(final_records),
            "status": "success"
        }
        
        # 输出 JSON 给 Go 后端解析
        print(json.dumps(result, ensure_ascii=False, indent=2))
        
    except Exception as e:
        error_result = {
            "error": str(e),
            "path": input_file,
            "status": "failed"
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)

if __name__ == "__main__":
    main()
