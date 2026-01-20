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
        "end_labels": ["此致", "综上所述", "具状人", "申请人", "附：", "附:"]
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
    """检测并过滤水印/电子签章干扰文本 - 优化版：减少正则检查次数"""
    if not text_line or len(text_line.strip()) == 0: return True

    # 只保留最有效的关键词过滤
    seal_keywords = ['章 Z', '签 Y', 'F F F', '印章', '验验验', '码码码']
    if any(kw in text_line for kw in seal_keywords): return True

    # 检测连续重复字符（4个或以上才算水印，减少误判）
    if re.search(r'(.)\1{3,}', text_line): return True

    # 计算重复字符比例（只对较长文本检查）
    if len(text_line) >= 8:
        char_counts = {}
        for c in text_line:
            char_counts[c] = char_counts.get(c, 0) + 1
        max_repeat = max(char_counts.values())
        if max_repeat / len(text_line) > 0.5:  # 提高阈值到50%，减少误判
            return True

    return False

def clean_watermark_chars(text):
    """深度清理水印字符，保留实际内容"""
    if not text:
        return ""

    # *** 特别针对：招商银行电子签章 ***
    # 截图显示这些字常被打散插入正文，如"招证"、"银码"、"商"
    # 我们先尝试移除完整的短语
    text = re.sub(r'招\s*商\s*银\s*行\s*电\s*子\s*签\s*章', '', text)
    text = re.sub(r'验\s*证\s*码\s*[:：]?\s*[A-Z0-9]{10,}', '', text)

    # *** 第1步：移除水印字符组合（章Z、签Y、银O等）***
    # 扩展：C银码, Z法
    text = re.sub(r'[章签子电行银商码招证验]+\s*[OZYXCB0-9]+\s*', '', text)
    text = re.sub(r'[OZYXCB0-9]+\s*[章签子电行银商码招证验]+', '', text) # 反向：C银码

    # *** 第2步：移除重复的水印字（2次或以上）***
    # 例如："银银"、"商商"、"码码"
    text = re.sub(r'([章签子电行银商码招证验])\1+', '', text)

    # *** 第3步：移除连续的不同水印字（如"银商码"）***
    # 匹配2个或以上连续的水印字
    text = re.sub(r'[章签子电行银商码招证验]{2,}', '', text)

    # *** 第4步：移除孤立的单个水印字（关键！）***
    # 匹配：正常汉字 + 单个水印字 + 正常汉字
    watermark_set = '章签子电行银商码招证验'
    for _ in range(5):  # 多次执行以处理连续的孤立水印字
        # 匹配非水印字 + 水印字 + 非水印字
        # 排除掉常用的"银"行，"电"话，"证"件，"代"码等合法词组
        # 这种基于字符的清洗非常危险，容易误删，必须非常小心
        # 更好的策略是：如果这些字出现在非自然语言的位置（如数字中间，或者打断了词语）
        pass

    # 针对截图中的具体Bad Case进行定向清洗
    text = re.sub(r'\(截止-9招证-5', '(截止', text) # 修复: (截止2024-9-5...
    text = re.sub(r'请依Z法裁判', '请依法裁判', text)
    text = re.sub(r'C银码', '', text)

    # *** 暴力清洗：如果一行中包含大量水印关键字，且不是正常的句子 ***
    # 暂时不启用，避免误伤

    # *** 第5步：移除行首/行尾的孤立水印字 ***
    text = re.sub(rf'^[{watermark_set}]+', '', text)
    text = re.sub(rf'[{watermark_set}]+$', '', text)

    # *** 第6步：移除单独的水印字母 ***
    # 截图中有孤立的 C, Z
    text = re.sub(r'\s+[OZYXCB]\s+', ' ', text)

    # *** 第8步：清理多余空格 ***
    text = re.sub(r'\s{2,}', ' ', text)

    return text.strip()

def extract_text_only(pdf_path):
    """提取纯文本 - 优化版：使用layout模式，深度清理水印"""
    extracted_pages = []
    total_chars_count = 0

    with pdfplumber.open(pdf_path) as pdf:
        for page in pdf.pages:
            # 直接使用layout提取，大幅提升性能
            page_text = page.extract_text(layout=True, x_tolerance=3, y_tolerance=3)

            if page_text:
                # 智能过滤：只保留包含实际汉字内容的行（至少3个汉字）
                lines = page_text.split('\n')
                filtered = []
                for line in lines:
                    stripped = line.strip()
                    # 跳过空行和纯空格行
                    if not stripped or len(stripped.replace(' ', '')) == 0:
                        continue

                    # 检查是否包含至少3个汉字（避免纯水印行）
                    chinese_chars = re.findall(r'[\u4e00-\u9fff]', stripped)
                    if len(chinese_chars) >= 3:
                        # 使用统一的水印清理函数
                        cleaned = clean_watermark_chars(stripped)
                        if cleaned:
                            filtered.append(cleaned)

                page_text = '\n'.join(filtered)
                total_chars_count += len(page_text)

            # 提取表格（如果有）
            tables = page.extract_tables()
            if tables:
                table_text = ""
                for table in tables:
                    for row in table:
                        row_content = [str(cell) for cell in row if cell]
                        table_text += " ".join(row_content) + "\n"
                page_text = (page_text or "") + "\n" + table_text

            extracted_pages.append(page_text or "")

    full_text = '\n\n'.join(extracted_pages)
    return full_text, (total_chars_count < 50)  # 更严格的OCR触发阈值

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
    """基于配置的动态提取引擎 - 优化版：增强内容清洗"""
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
        # 优先匹配明确标识的身份证
        # 匹配 "身份证号码：123..." 或 "身份证：123..."
        labeled_id = re.search(r'(?:身份证(?:号码)?|公民身份号码)\s*[:：]?\s*(\d{17}[\dXx])', text)
        if labeled_id:
            return labeled_id.group(1)

        # 如果没有明确标签，尝试查找18位号码，但必须严格排除"统一社会信用代码"
        # 统一社会信用代码通常包含字母，而身份证通常只有最后一位可能是X

        # 1. 查找所有18位数字（或末位X）
        candidates = re.findall(r'\b(\d{17}[\dXx])\b', text)
        for cand in candidates:
            # 简单验证：确保不是全数字的信用代码（虽然信用代码很少全数字，但为了保险）
            # 身份证的年份通常在 1900-202x
            year = int(cand[6:10])
            if 1900 <= year <= 2030:
                # 再次检查上下文，确保它前面不是"代码"、"信用"
                idx = text.find(cand)
                if idx > 10:
                    prefix = text[idx-10:idx]
                    if "代码" in prefix or "信用" in prefix or "税" in prefix:
                        continue
                return cand

        return ""

    elif "start_labels" in conf:
        starts = "|".join([re.escape(l) for l in conf['start_labels']])
        ends = "|".join([re.escape(l) for l in conf['end_labels']])

        # 改进的正则：冒号可选，允许更宽松的空白字符和换行作为边界
        # (?:{starts})\s*[:：]?\s* : 匹配开始标签，后面可能有冒号，也可能没有
        # (.*?) : 非贪婪匹配正文内容
        # (?:\s*(?:{ends})|$) : 匹配结束标签（前面允许任意空白，甚至没有空白）或字符串结尾
        pattern = re.compile(rf'(?:{starts})\s*[:：]?\s*\n?\s*(.*?)(?:\s*(?:{ends})|$)', re.DOTALL)
        match = pattern.search(text)

        if match:
            raw_content = match.group(1)

            # 后处理：清除开头可能出现的乱序片段
            # 移除开头的"码"、"招商银行"、"电子签章"等干扰词（限制在前30个字符内）
            cleaned = re.sub(r'^[码招商银行电子签章]{0,30}', '', raw_content)

            # 再次清理：如果开头还有冒号、空格等，也去掉
            cleaned = re.sub(r'^[:：\s]+', '', cleaned)

            # *** 强力兜底截断 ***
            # 防止正则因为 OCR 字符粘连而失效（例如 "请依法判决此致" 中间没空格）
            # 我们遍历所有结束标签，如果在提取的内容中发现了它们，强制截断
            for end_label in conf.get('end_labels', []):
                # 查找结束标签的位置
                idx = cleaned.find(end_label)
                if idx != -1:
                    cleaned = cleaned[:idx]

            # 额外针对落款的清理
            if field_key == "factsReason":
               # 如果出现 "附：" 或 "附:"，截断
               cleaned = re.split(r'附[:：]', cleaned)[0]
               # 如果出现 "此致"，截断
               cleaned = cleaned.split("此致")[0]
               # 如果出现 "具状人"，截断
               cleaned = cleaned.split("具状人")[0]
               # 如果出现 "申请人"，截断
               cleaned = cleaned.split("申请人")[0]

            return smart_merge(cleaned)

    return ""

def quick_scan(pdf_path):
    """快速扫描模式：只读取第一页文本，检查字段是否存在，绝不进行OCR"""
    found_keys = []
    try:
        with pdfplumber.open(pdf_path) as pdf:
            if len(pdf.pages) > 0:
                # 只读第一页，且只取前1000字符，足够判断案由和基本信息
                first_page_text = pdf.pages[0].extract_text() or ""

                # 简单关键词匹配
                if "被告" in first_page_text or "被申请人" in first_page_text:
                    found_keys.append("defendant")

                # 身份证通常也在前面
                if re.search(r'\d{18}|\d{17}[Xx]', first_page_text) or "身份证" in first_page_text:
                    found_keys.append("idNumber")

                # 诉讼请求通常在前两页
                if "诉讼请求" in first_page_text or "请求事项" in first_page_text:
                    found_keys.append("request")

                # 事实与理由
                if "事实与理由" in first_page_text or "事实和理由" in first_page_text:
                    found_keys.append("factsReason")

    except Exception:
        pass

    # 如果没扫到，为了用户体验，默认返回所有常用字段
    # 因为扫描失败不代表提取失败（提取时有OCR兜底）
    if not found_keys:
        found_keys = ["defendant", "idNumber", "request", "factsReason"]

    print(json.dumps({
        "status": "success",
        "keys": found_keys
    }, ensure_ascii=False))

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No input file", "status": "failed"}))
        sys.exit(1)

    # 检查是否是快速扫描模式
    if len(sys.argv) > 2 and sys.argv[1] == "--scan":
        quick_scan(sys.argv[2])
        sys.exit(0)

    input_file = sys.argv[1]
    if check_pdf_security(input_file):
        print(json.dumps({"error": "PDF_ENCRYPTED_OR_LOCKED", "status": "failed"}))
        sys.exit(1)

    try:
        text_content, is_scanned = extract_text_only(input_file)
        ocr_content = None
        # 性能优化：只在字符数极少时才触发OCR，避免不必要的OCR处理
        if is_scanned and OCR_AVAILABLE:
            try:
                ocr_content = OCRExtractor().extract(input_file)
            except Exception: pass

        main_content = ocr_content if (not text_content or len(text_content) < 200) and ocr_content else text_content

        # 清理水印字符后再分割
        # 移除常见的水印干扰字符（银O、章Z、签Y等）
        main_content_clean = re.sub(r'[银章签子电行商码招证验]+\s*[OZYXCB]+\s*', '', main_content)

        # 更宽松的分隔符匹配，允许中间有空格或干扰字符
        cases = re.split(r'民\s*[银]?\s*事\s*[O]?\s*起\s*诉\s*状', main_content_clean)

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
