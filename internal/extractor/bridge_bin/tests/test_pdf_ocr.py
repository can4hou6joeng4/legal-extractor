import sys
import os
import unittest

# Add parent dir to path to import pdf_bridge
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pdf_bridge

class TestPDFOCR(unittest.TestCase):
    def test_imports(self):
        """Verify libraries are installed and importable"""
        try:
            import rapidocr_onnxruntime
            import fitz  # PyMuPDF
        except ImportError as e:
            self.fail(f"Import failed: {e}")
        self.assertTrue(True)

    def test_ocr_extractor_init(self):
        """Test OCRExtractor initialization"""
        # Ensure imports are available for the module under test
        try:
            extractor = pdf_bridge.OCRExtractor()
            self.assertIsNotNone(extractor.ocr)
        except AttributeError:
             self.fail("OCRExtractor not defined in pdf_bridge")
        except ImportError:
             self.skipTest("OCR libraries not installed, skipping OCR tests")

if __name__ == '__main__':
    unittest.main()
