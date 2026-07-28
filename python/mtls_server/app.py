"""FastAPI application — mTLS health-check ASGI app."""

from __future__ import annotations

import logging
import os

from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from starlette.requests import Request

logger = logging.getLogger("mtls_server")

app = FastAPI(title="mtls-server", version="0.1.0")


@app.get("/v1/health")
async def health() -> JSONResponse:
    """Return status ok after exercising AES-GCM."""
    try:
        key = os.urandom(32)
        nonce = os.urandom(12)
        plaintext = os.urandom(256)
        aesgcm = AESGCM(key)
        aesgcm.encrypt(nonce, plaintext, None)
    except Exception:
        logger.exception("health probe encryption failed")
        return JSONResponse(
            status_code=500,
            content={"error": "internal server error"},
        )

    return JSONResponse(content={"status": "ok"})


@app.exception_handler(404)
async def _not_found(_request: Request, _exc: Exception) -> JSONResponse:
    return JSONResponse(status_code=404, content={"error": "not found"})


@app.exception_handler(500)
async def _internal_error(_request: Request, _exc: Exception) -> JSONResponse:
    return JSONResponse(status_code=500, content={"error": "internal server error"})
