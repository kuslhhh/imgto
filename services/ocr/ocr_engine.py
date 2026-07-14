"""
OCR Engine — PaddleOCR Wrapper
================================

Loads PaddleOCR into memory once at startup and provides a stateless
extraction function that can be called per request.

This is designed to run as a long-lived service so model loading overhead
(1-3s) is paid only once.
"""

import io
import logging
import time
from typing import Optional

import numpy as np
from PIL import Image

logger = logging.getLogger(__name__)


class OCREngine:
    """Wrapper around PaddleOCR that keeps the model warm in memory."""

    def __init__(self, use_gpu: bool = False, lang: str = "en"):
        """
        Initialize the OCR engine.

        Args:
            use_gpu: Whether to use GPU acceleration.
            lang: Language code for OCR (e.g., "en", "ch", "en,ch").
        """
        self.lang = lang
        self.use_gpu = use_gpu
        self._model = None
        self._load_model()

    def _load_model(self) -> None:
        """Load PaddleOCR model (called once at startup)."""
        start = time.time()
        logger.info(
            "Loading PaddleOCR model (this may take a few seconds)...",
            extra={"lang": self.lang, "use_gpu": self.use_gpu},
        )

        try:
            from paddleocr import PaddleOCR

            self._model = PaddleOCR(
                use_angle_cls=True,
                lang=self.lang,
                use_gpu=self.use_gpu,
                show_log=False,
                # Use a small detection model for faster inference
                det_db_thresh=0.3,
                det_db_box_thresh=0.5,
                # Use a modest recognition model
                rec_batch_num=6,
            )
            elapsed = time.time() - start
            logger.info(
                "PaddleOCR model loaded successfully",
                extra={"elapsed_seconds": round(elapsed, 2)},
            )
        except ImportError as e:
            logger.error(
                "Failed to import PaddleOCR. Is paddleocr installed?",
                extra={"error": str(e)},
            )
            raise
        except Exception as e:
            logger.error(
                "Failed to load PaddleOCR model",
                extra={"error": str(e)},
            )
            raise

    @property
    def is_loaded(self) -> bool:
        """Whether the model is loaded and ready."""
        return self._model is not None

    def extract_text(self, image_bytes: bytes) -> dict:
        """
        Run OCR on image bytes and return structured results.

        Args:
            image_bytes: Raw image bytes (JPEG, PNG, etc.).

        Returns:
            Dictionary with keys:
                - text: Concatenated extracted text
                - confidence: Overall confidence (0.0 to 1.0)
                - blocks: List of text blocks with bounding boxes
                - tables: List of detected tables (empty for now)
                - error: Error message if something went wrong
        """
        if not self._model:
            return {
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": "OCR model not loaded",
            }

        start = time.time()

        try:
            # Load image from bytes
            image = Image.open(io.BytesIO(image_bytes)).convert("RGB")
            image_np = np.array(image)

            # Run OCR
            result = self._model.ocr(image_np, cls=True)

            elapsed = time.time() - start

            if not result or not result[0]:
                return {
                    "text": "",
                    "confidence": 0.0,
                    "blocks": [],
                    "tables": [],
                    "error": "",
                    "processing_time_ms": round(elapsed * 1000),
                }

            # Parse results
            blocks = []
            all_text_parts = []
            total_confidence = 0.0
            block_count = 0

            for line in result[0]:
                if line is None:
                    continue

                # line format: [bbox, (text, confidence)]
                bbox, (text, confidence) = line

                # Convert bbox from list of lists to [[x1,y1],[x2,y2],[x3,y3],[x4,y4]]
                bounding_box = [[int(pt[0]), int(pt[1])] for pt in bbox]

                blocks.append({
                    "text": text,
                    "confidence": round(float(confidence), 4),
                    "bounding_box": bounding_box,
                })

                all_text_parts.append(text)
                total_confidence += float(confidence)
                block_count += 1

            # Calculate overall confidence (average)
            overall_confidence = round(total_confidence / max(block_count, 1), 4)

            logger.debug(
                "OCR completed",
                extra={
                    "blocks": block_count,
                    "text_length": len(" ".join(all_text_parts)),
                    "confidence": overall_confidence,
                    "elapsed_ms": round(elapsed * 1000),
                },
            )

            return {
                "text": "\n".join(all_text_parts),
                "confidence": overall_confidence,
                "blocks": blocks,
                "tables": [],  # Table detection not yet implemented
                "error": "",
                "processing_time_ms": round(elapsed * 1000),
            }

        except Exception as e:
            logger.error("OCR processing error", extra={"error": str(e)})
            return {
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": str(e),
            }


# Global engine instance — initialized once at module import
_engine: Optional[OCREngine] = None


def init_engine(use_gpu: bool = False, lang: str = "en") -> OCREngine:
    """Initialize the global OCR engine (called at startup)."""
    global _engine
    if _engine is None:
        _engine = OCREngine(use_gpu=use_gpu, lang=lang)
    return _engine


def get_engine() -> Optional[OCREngine]:
    """Get the global OCR engine instance."""
    return _engine
