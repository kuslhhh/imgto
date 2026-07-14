"""
Vision Service — FastAPI Application
======================================

Provides image understanding capabilities: semantic descriptions,
UI component detection, and layout analysis.

Endpoints:
    POST /describe  — Analyze an image and return structured description
    GET  /health    — Health check endpoint
"""

import logging
import os
import time

from fastapi import FastAPI, File, Form, UploadFile
from fastapi.responses import JSONResponse

from vision_engine import init_engine, get_engine

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("vision-service")

# ---------------------------------------------------------------------------
# App
# ---------------------------------------------------------------------------
app = FastAPI(
    title="Vision Service",
    version="1.0.0",
    description="Vision-Language model service for image understanding",
)


@app.on_event("startup")
async def startup():
    """Initialize the vision engine on startup."""
    model_name = os.getenv("VISION_MODEL", "microsoft/Florence-2-base")
    device = os.getenv("VISION_DEVICE", "cpu")

    logger.info("Starting vision service",
                extra={"model": model_name, "device": device})
    start = time.time()

    import asyncio
    await asyncio.to_thread(init_engine, model_name=model_name, device=device)

    elapsed = time.time() - start
    logger.info("Vision service ready",
                extra={"startup_seconds": round(elapsed, 2)})


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


@app.post("/describe")
async def describe(
    image: UploadFile = File(..., description="Image file (JPEG, PNG, etc.)"),
    detail_level: str = Form("detailed",
                             description="Detail level: basic, detailed, ui"),
):
    """
    Analyze an image and return structured description.

    Accepts multipart form data with an 'image' field.
    Returns JSON with description, UI components, layout, and tags.
    """
    engine = get_engine()
    if not engine or not engine.is_loaded:
        return JSONResponse(
            status_code=503,
            content={
                "description": "",
                "ui_components": [],
                "layout": {},
                "tags": [],
                "error": "Vision engine not loaded",
            },
        )

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
                    "description": "",
                    "ui_components": [],
                    "layout": {},
                    "tags": [],
                    "error": f"Unsupported image type: {image.content_type}",
                },
            )

    # Read image bytes
    try:
        image_bytes = await image.read()
    except Exception as e:
        return JSONResponse(
            status_code=400,
            content={
                "description": "",
                "ui_components": [],
                "layout": {},
                "tags": [],
                "error": f"Failed to read image: {e}",
            },
        )

    if len(image_bytes) == 0:
        return JSONResponse(
            status_code=400,
            content={
                "description": "",
                "ui_components": [],
                "layout": {},
                "tags": [],
                "error": "Empty image file",
            },
        )

    # Validate size (max 20MB by default)
    max_size = int(os.getenv("VISION_MAX_IMAGE_SIZE_MB", "20")) * 1024 * 1024
    if len(image_bytes) > max_size:
        return JSONResponse(
            status_code=413,
            content={
                "description": "",
                "ui_components": [],
                "layout": {},
                "tags": [],
                "error": f"Image too large: {len(image_bytes)} bytes",
            },
        )

    # Run vision analysis
    try:
        import asyncio
        result = await asyncio.to_thread(
            engine.describe, image_bytes, detail_level
        )
    except Exception as e:
        logger.error("Vision processing failed", extra={"error": str(e)})
        return JSONResponse(
            status_code=500,
            content={
                "description": "",
                "ui_components": [],
                "layout": {},
                "tags": [],
                "error": f"Vision processing failed: {e}",
            },
        )

    return JSONResponse(content=result)


# ---------------------------------------------------------------------------
# Entrypoint
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    import uvicorn

    host = os.getenv("VISION_HOST", "0.0.0.0")
    port = int(os.getenv("VISION_PORT", "9080"))
    log_level = os.getenv("VISION_LOG_LEVEL", "info").lower()

    uvicorn.run(
        "app:app",
        host=host,
        port=port,
        log_level=log_level,
        reload=False,
    )
