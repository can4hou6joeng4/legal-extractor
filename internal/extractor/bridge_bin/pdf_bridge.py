#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import json
import pdfplumber
import os
import re

# 模式配置：通过别名和锚点提高匹配灵活性，无需用户自定义正则
FIELD_CONFIG = {
    "defendant": {
        "labels": ["被告人", "被告", "被上诉人", "原审被告", "被申请人"],
        "anchors": ["性别", "男", "女", "出生", "住址", "身份证", "现住", "联系电话", "案由", "住所地", "法定代表人"]
    },
    "idNumber": {
        "labels": ["身份证号码", "身份证号", "公民身份号码", "统一社会信用代码", "证件号码", "组织机构代码"],
    },
    "request": {
        "start_labels": ["诉讼请求", "请求事项", "申请事项"],
        "end_labels": ["事实与理由", "事实和理由", "事实及理由"]
    },
    "factsReason": {
        "start_labels": ["事实与理由", "事实和理由", "事实及理由"],
        "end_labels": ["此致", "综上所述", "具状人", "申请人"]
    }
}

OCR_ERROR_MSG = ""
try:
    from rapidocr_onnxruntime import RapidOCR
    import fitz  # PyMuPDF
    OCR_AVAILABLE = True
except ImportError as e:
    OCR_AVAILABLE = False
    OCR_ERROR_MSG = str(e)
except Exception as e:
    OCR_AVAILABLE = False
    OCR_ERROR_MSG = f"OCR Init Failed: {str(e)}"

class OCRExtractor:
    def __init__(self):
        if not OCR_AVAILABLE:
            raise ImportError(f"OCR libraries not installed or failed to load: {OCR_ERROR_MSG}")
        try:
            self.ocr = RapidOCR()
        except Exception as e:
            raise RuntimeError(f"RapidOCR initialization failed: {str(e)}")

    def extract(self, pdf_path):
        doc = fitz.open(pdf_path)
        full_text = []
        for page in doc:
            pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
            img_bytes = pix.tobytes("png")
            result, _ = self.ocr(img_bytes)
            page_content = []
            if result:
                boxes = []
                for item in result:
                    box, text, _ = item
                    y_center = sum(p[1] for p in box) / 4
                    x_min = min(p[0] for p in box)
                    boxes.append({"text": text, "y": y_center, "x": x_min})
                # 按行聚合排序
                boxes.sort(key=lambda b: (int(b['y'] / 20), b['x']))
                page_content = [box['text'] for box in boxes]
            full_text.append('\n'.join(page_content))
        doc.close()
        return '\n\n'.join(full_text)

def check_pdf_security(pdf_path):
    """检测 PDF 是否加密或受限"""
    try:
        with pdfplumber.open(pdf_path) as pdf:
            if pdf.metadata.get('Encryption') or getattr(pdf, "is_encrypted", False):
                return True
    except Exception:
        return True
    return False

def is_seal_like_text(text_line):
    """检测并过滤水印/电子签章干扰文本"""
    if not text_line or len(text_line.strip()) == 0: return True

    # 基础关键词过滤
    seal_keywords = ['章 Z', '签 Y', 'F F F', '印章', '子 C', '电 B', '验验验', '码码码', '电电电', '签签签', '章章章', '银银银', '商商商']
    if any(kw in text_line for kw in seal_keywords): return True

    # 检测连续重复字符（3个或以上相同字符连续出现）
    if re.search(r'(.)\1{2,}', text_line): return True

    # 检测混合重复模式：如 "OOO银银银", "777RRR444ZZZ"
    # 模式：(字母或数字重复2次以上) + (汉字重复2次以上) 或反过来
    if re.search(r'([A-Za-z0-9])\1{1,}(.)\2{1,}', text_line): return True
    if re.search(r'(.)\1{1,}([A-Za-z0-9])\2{1,}', text_line): return True

    # 检测数字+大写字母交替重复模式：如 "777RRR444ZZZ333YYY"
    if re.search(r'(\d{2,}[A-Z]{2,}){2,}', text_line): return True

    # 检测大写字母+汉字交替重复：如 "FFF子子子CCCCCC"
    if re.search(r'([A-Z]{2,}[\u4e00-\u9fff]{2,}){2,}', text_line): return True

    # 连续大写字母常见于水印（3个或以上）
    if re.search(r'[A-Z]{3,}', text_line) and len(text_line) < 20: return True

    # 计算重复字符比例，如果过高则认为是干扰
    if len(text_line) >= 6:
        char_counts = {}
        for c in text_line:
            char_counts[c] = char_counts.get(c, 0) + 1
        max_repeat = max(char_counts.values())
        if max_repeat / len(text_line) > 0.4:  # 某字符占比超过40%
            return True

    return False

def extract_text_only(pdf_path):
    """提取纯文本，支持简单的表格识别还原"""
    extracted_pages = []
    total_chars = 0
    with pdfplumber.open(pdf_path) as pdf:
        for page in pdf.pages:
            chars = page.chars
            total_chars += len(chars)

            # 优先检查是否存在表格
            tables = page.extract_tables()
            table_text = ""
            if tables:
                for table in tables:
                    for row in table:
                        row_content = [str(cell) for cell in row if cell]
                        table_text += " ".join(row_content) + "\n"

            if len(chars) > 0:
                lines = {}
                for char in chars:
                    y = round(char['top'], 1)
                    if y not in lines: lines[y] = []
                    lines[y].append(char)

                sorted_lines = []
                for y in sorted(lines.keys()):
                    line_chars = sorted(lines[y], key=lambda c: c['x0'])
                    line_text = ''.join(c['text'] for c in line_chars)
                    if not is_seal_like_text(line_text):
                        sorted_lines.append(line_text.strip())
                page_text = '\n'.join(sorted_lines) + "\n" + table_text
            else:
                page_text = page.extract_text(layout=True) or ""

            extracted_pages.append(page_text)

    full_text = '\n\n'.join(extracted_pages)
    return full_text, (total_chars < 100)

def smart_merge(text):
    """智能合并换行符，优化阅读排版"""
    if not text: return ""
    text = text.replace('\r\n', '\n')
    text = re.sub(r'([。；？！])\n', r'\1[LOGICAL_NL]', text)
    text = re.sub(r'\n(\s*(?:[一二三四五六七八九十\d]+[、．]|[(（][一二三四五六七八九十\d]+[)）]))', r'[LOGICAL_NL]\1', text)
    text = text.replace('\n', '').replace('[LOGICAL_NL]', '\n')
    lines = [line.strip() for line in text.split('\n') if line.strip()]
    return '\n'.join([''.join(line.split()) for line in lines])

def extract_field_by_config(text, field_key):
    """基于配置的动态提取引擎"""
    conf = FIELD_CONFIG.get(field_key)
    if not conf: return ""

    if field_key == "defendant":
        labels_pattern = "|".join([re.escape(l) for l in conf['labels']])
        match = re.search(rf'({labels_pattern})\s*[:：]?\s*(.*)', text)
        if match:
            content = match.group(2).replace('\n', '')
            end_pos = len(content)
            for anchor in conf['anchors']:
                pos = content.find(anchor)
                if pos != -1 and pos < end_pos: end_pos = pos
            name = content[:end_pos].strip(' ,，、:：；;\t')
            return name[:20]

    elif field_key == "idNumber":
        id_match = re.search(r'\b(\d{17}[\dXx])\b', text)
        if id_match: return id_match.group(1)
        code_match = re.search(r'([A-Z0-9]{18}|[A-Z0-9]{9}-[A-Z0-9])', text)
        if code_match: return code_match.group(1)

    elif "start_labels" in conf:
        starts = "|".join([re.escape(l) for l in conf['start_labels']])
        ends = "|".join([re.escape(l) for l in conf['end_labels']])
        pattern = re.compile(rf'(?:{starts})\s*[:：]?\s*(.*?)(?:{ends}|$)', re.DOTALL)
        match = pattern.search(text)
        if match: return smart_merge(match.group(1))

    return ""

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No input file", "status": "failed"}))
        sys.exit(1)

    input_file = sys.argv[1]
    if check_pdf_security(input_file):
        print(json.dumps({"error": "PDF_ENCRYPTED_OR_LOCKED", "status": "failed"}))
        sys.exit(1)

    try:
        text_content, is_scanned = extract_text_only(input_file)
        ocr_content = None
        if (is_scanned or "被告" not in text_content) and OCR_AVAILABLE:
            try:
                ocr_content = OCRExtractor().extract(input_file)
            except Exception: pass

        main_content = ocr_content if (not text_content or len(text_content) < 200) and ocr_content else text_content
        cases = re.split(r'民\s*事\s*起\s*诉\s*状', main_content)

        records = []
        for case_text in cases:
            if not case_text.strip(): continue
            rec = {k: extract_field_by_config(case_text, k) for k in FIELD_CONFIG.keys()}
            if any(rec.values()): records.append(rec)

        print(json.dumps({
            "path": input_file,
            "records": records,
            "count": len(records),
            "status": "success",
            "is_ocr_used": ocr_content is not None
        }, ensure_ascii=False, indent=2))

    except Exception as e:
        print(json.dumps({"error": str(e), "status": "failed", "path": input_file}, ensure_ascii=False))
        sys.exit(1)

if __name__ == "__main__":
    main()
