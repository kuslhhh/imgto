"""
OCR Service — FastAPI Application
===================================

Provides a REST API for OCR processing using PaddleOCR.
The model is loaded at startup and kept warm in memory.

Endpoints:
    POST /ocr      — Accept image upload, return OCR results as JSON
    GET  /health   — Health check endpoint
"""

import io
import logging
import os
import time

from fastapi import FastAPI, File, Form, UploadFile, HTTPException
from fastapi.responses import JSONResponse

from ocr_engine import init_engine, get_engine

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("ocr-service")

# ---------------------------------------------------------------------------
# App
# ---------------------------------------------------------------------------
app = FastAPI(
    title="OCR Service",
    version="1.0.0",
    description="PaddleOCR-based text extraction service for the OCR MCP Server",
)


@app.on_event("startup")
async def startup():
    """Initialize the OCR engine on startup (load model into memory)."""
    use_gpu = os.getenv("OCR_USE_GPU", "false").lower() == "true"
    lang = os.getenv("OCR_LANG", "en")

    logger.info("Starting OCR service", extra={"use_gpu": use_gpu, "lang": lang})
    start = time.time()

    # Run model loading in a thread to avoid blocking the event loop
    # (PaddleOCR can take a few seconds to load)
    import asyncio
    await asyncio.to_thread(init_engine, use_gpu=use_gpu, lang=lang)

    elapsed = time.time() - start
    logger.info("OCR service ready", extra={"startup_seconds": round(elapsed, 2)})


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------


@app.get("/health")
async def health():
    """Health check endpoint."""
    engine = get_engine()
    return {
        "status": "ok" if engine and engine.is_loaded else "loading",
        "model_loaded": engine.is_loaded if engine else False,
    }


@app.post("/ocr")
async def ocr(
    image: UploadFile = File(..., description="Image file (JPEG, PNG, etc.)"),
    preprocess: str = Form("auto", description="Preprocessing mode: auto, none, grayscale, threshold"),
):
    """
    Extract text from an uploaded image using OCR.

    Accepts multipart form data with an 'image' field containing the image file.
    Returns structured JSON with extracted text, confidence scores, and bounding boxes.
    """
    engine = get_engine()
    if not engine or not engine.is_loaded:
        raise HTTPException(status_code=503, detail="OCR engine not loaded")

    # Validate file type
    if image.content_type:
        allowed_types = [
            "image/jpeg", "image/png", "image/webp",
            "image/bmp", "image/tiff", "image/gif",
        ]
        if image.content_type not in allowed_types:
            return JSONResponse(
                status_code=400,
                content={
                    "text": "",
                    "confidence": 0.0,
                    "blocks": [],
                    "tables": [],
                    "error": f"Unsupported image type: {image.content_type}. "
                             f"Supported types: {', '.join(allowed_types)}",
                },
            )

    # Read image bytes
    try:
        image_bytes = await image.read()
    except Exception as e:
        return JSONResponse(
            status_code=400,
            content={
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": f"Failed to read image: {e}",
            },
        )

    if len(image_bytes) == 0:
        return JSONResponse(
            status_code=400,
            content={
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": "Empty image file",
            },
        )

    # Validate size (max 20MB by default)
    max_size = int(os.getenv("OCR_MAX_IMAGE_SIZE_MB", "20")) * 1024 * 1024
    if len(image_bytes) > max_size:
        return JSONResponse(
            status_code=413,
            content={
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": f"Image too large: {len(image_bytes)} bytes (max {max_size} bytes)",
            },
        )

    # Run OCR
    try:
        import asyncio
        result = await asyncio.to_thread(engine.extract_text, image_bytes)
    except Exception as e:
        logger.error("OCR processing failed", extra={"error": str(e)})
        return JSONResponse(
            status_code=500,
            content={
                "text": "",
                "confidence": 0.0,
                "blocks": [],
                "tables": [],
                "error": f"OCR processing failed: {e}",
            },
        )

    # Return results
    return JSONResponse(content=result)


# ---------------------------------------------------------------------------
# Entrypoint
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    import uvicorn

    host = os.getenv("OCR_HOST", "0.0.0.0")
    port = int(os.getenv("OCR_PORT", "9090"))
    log_level = os.getenv("OCR_LOG_LEVEL", "info").lower()

    uvicorn.run(
        "app:app",
        host=host,
        port=port,
        log_level=log_level,
        reload=False,
    )
