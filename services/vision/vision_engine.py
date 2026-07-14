"""
Vision Engine — Image Understanding Wrapper
=============================================

Loads a vision model (Florence-2 or Qwen2.5-VL) into memory at startup
and provides image description, UI component detection, and layout analysis.

This is designed to run as a long-lived service so model loading overhead
is paid only once.
"""

import io
import logging
import time
from typing import Optional

from PIL import Image

logger = logging.getLogger(__name__)


class VisionEngine:
    """Wrapper around a vision-language model for image understanding."""

    def __init__(self, model_name: str = "microsoft/Florence-2-large",
                 device: str = "cpu"):
        """
        Initialize the vision engine.

        Args:
            model_name: HuggingFace model name or path.
            device: "cpu" or "cuda" for GPU.
        """
        self.model_name = model_name
        self.device = device
        self._model = None
        self._processor = None
        self._load_model()

    def _load_model(self) -> None:
        """Load the vision model (called once at startup)."""
        start = time.time()
        logger.info(
            "Loading vision model (this may take 10-30 seconds)...",
            extra={"model": self.model_name, "device": self.device},
        )

        try:
            from transformers import AutoModelForCausalLM, AutoProcessor

            self._processor = AutoProcessor.from_pretrained(
                self.model_name,
                trust_remote_code=True,
            )
            self._model = AutoModelForCausalLM.from_pretrained(
                self.model_name,
                trust_remote_code=True,
                device_map=self.device if self.device == "cpu" else "auto",
            ).eval()

            elapsed = time.time() - start
            logger.info(
                "Vision model loaded successfully",
                extra={"elapsed_seconds": round(elapsed, 2)},
            )
        except ImportError as e:
            logger.error(
                "Failed to import transformers. Is it installed?",
                extra={"error": str(e)},
            )
            raise
        except Exception as e:
            logger.error(
                "Failed to load vision model. The model might be too large "
                "for this device. Try a smaller model like "
                "'microsoft/Florence-2-base'.",
                extra={"error": str(e)},
            )
            raise

    @property
    def is_loaded(self) -> bool:
        return self._model is not None

    def describe(self, image_bytes: bytes,
                 detail_level: str = "detailed") -> dict:
        """
        Analyze an image and return a structured description.

        Args:
            image_bytes: Raw image bytes.
            detail_level: 'basic', 'detailed', or 'ui'.

        Returns:
            Dictionary with description, UI components, layout, and tags.
        """
        if not self._model:
            return {
                "description": "",
                "ui_components": [],
                "layout": {"type": "", "regions": [], "description": ""},
                "tags": [],
                "error": "Vision model not loaded",
            }

        start = time.time()

        try:
            image = Image.open(io.BytesIO(image_bytes)).convert("RGB")

            prompt = self._build_prompt(detail_level)
            inputs = self._processor(text=prompt, images=image, return_tensors="pt")

            if self.device == "cpu":
                inputs = {k: v.to("cpu") for k, v in inputs.items()}
            else:
                inputs = {k: v.to(self._model.device) for k, v in inputs.items()}

            generated_ids = self._model.generate(
                **inputs,
                max_new_tokens=512,
                num_beams=3,
                temperature=0.7,
            )

            generated_text = self._processor.batch_decode(
                generated_ids, skip_special_tokens=True
            )[0]

            elapsed = time.time() - start

            # Parse the generated text into structured output
            result = self._parse_output(generated_text, detail_level, elapsed)

            logger.debug(
                "Vision completed",
                extra={
                    "elapsed_ms": round(elapsed * 1000),
                    "detail_level": detail_level,
                },
            )

            return result

        except Exception as e:
            logger.error("Vision processing error", extra={"error": str(e)})
            return {
                "description": "",
                "ui_components": [],
                "layout": {"type": "", "regions": [], "description": ""},
                "tags": [],
                "error": str(e),
            }

    def _build_prompt(self, detail_level: str) -> str:
        """Build the appropriate prompt based on detail level."""
        prompts = {
            "basic": "Describe this image briefly.",
            "detailed": (
                "Describe this image in detail. What are the main elements, "
                "text content, colors, and layout?"
            ),
            "ui": (
                "Analyze this screenshot/UI. List all UI components "
                "(buttons, inputs, modals, navigation), describe the layout, "
                "and tag the type of interface."
            ),
        }
        return prompts.get(detail_level, prompts["detailed"])

    def _parse_output(self, text: str, detail_level: str,
                      elapsed: float) -> dict:
        """Parse model output into structured result."""
        result = {
            "description": text,
            "ui_components": [],
            "layout": {"type": "unknown", "regions": [], "description": ""},
            "tags": [],
            "processing_time_ms": round(elapsed * 1000),
        }

        # For UI level, try to extract structured information
        if detail_level == "ui":
            result["tags"] = self._extract_tags(text)
            result["layout"]["description"] = self._extract_layout(text)

        return result

    def _extract_tags(self, text: str) -> list:
        """Extract relevant tags from the description."""
        tags = []
        keywords = {
            "form": ["form", "input", "field", "submit"],
            "dashboard": ["dashboard", "chart", "graph", "metric"],
            "article": ["article", "text", "paragraph", "content"],
            "login": ["login", "sign in", "password", "username"],
            "settings": ["setting", "preference", "configuration"],
            "modal": ["modal", "dialog", "popup", "overlay"],
            "navigation": ["nav", "menu", "sidebar", "header"],
            "table": ["table", "grid", "row", "column"],
        }
        text_lower = text.lower()
        for tag, kws in keywords.items():
            if any(kw in text_lower for kw in kws):
                tags.append(tag)
        return tags

    def _extract_layout(self, text: str) -> str:
        """Extract layout description."""
        lines = text.split(".")
        relevant = [l.strip() for l in lines
                    if any(w in l.lower() for w in
                           ["layout", "arranged", "section", "column",
                            "row", "top", "bottom", "left", "right",
                            "center", "header", "footer", "sidebar"])]
        return ". ".join(relevant[:3]) if relevant else ""


# Global engine instance
_engine: Optional[VisionEngine] = None


def init_engine(model_name: str = None, device: str = "cpu") -> VisionEngine:
    """Initialize the global vision engine."""
    global _engine
    if _engine is None:
        if model_name is None:
            model_name = "microsoft/Florence-2-base"
        _engine = VisionEngine(model_name=model_name, device=device)
    return _engine


def get_engine() -> Optional[VisionEngine]:
    """Get the global vision engine instance."""
    return _engine
