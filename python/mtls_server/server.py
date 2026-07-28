"""mTLS server launcher via Uvicorn with multi-worker support."""

from __future__ import annotations

import logging
import os
import ssl
from typing import NoReturn

import uvicorn
from uvicorn import Config

from .config import load_config

logger = logging.getLogger("mtls_server")


class _MtlsConfig(uvicorn.Config):
    """Uvicorn Config that enforces TLSv1.3 minimum on the SSL context.

    The standard ``create_ssl_context`` creates an ``ssl.SSLContext`` with the
    default minimum version (TLS 1.2).  This subclass overrides ``load()`` to
    pin the minimum to TLS 1.3 after the parent builds the context.
    """

    def load(self) -> None:
        super().load()
        if self.ssl is not None:
            self.ssl.minimum_version = ssl.TLSVersion.TLSv1_3
            if self.ssl_cert_reqs is not None:
                self.ssl.verify_mode = ssl.VerifyMode(self.ssl_cert_reqs)


def run_server(
    *,
    addr: tuple[str, int] | None = None,
    ssl_ctx: ssl.SSLContext | None = None,
) -> NoReturn:
    """Start the mTLS server via Uvicorn and block until shutdown.

    Handles SIGINT/SIGTERM via Uvicorn's own signal handling (no conflicting
    signal handlers installed here).

    Config is loaded from env vars in every worker process (pickling-safe).
    The *ssl_ctx* parameter is accepted for backward compatibility but not
    used internally — the picklable file-based path always wins.

    Never returns — exits the process.
    """
    cfg = load_config()
    host, port = addr if addr else cfg.listen_addr
    workers = max(1, os.cpu_count() or 1)

    config = _MtlsConfig(
        app="mtls_server.app:app",
        host=host,
        port=port,
        workers=workers,
        ssl_keyfile=cfg.server_key_file,
        ssl_certfile=cfg.server_cert_file,
        ssl_ca_certs=cfg.client_ca_file,
        ssl_cert_reqs=ssl.CERT_REQUIRED,
        log_level="info",
        access_log=True,
    )

    uvicorn.run(
        "mtls_server.app:app",
        host=host,
        port=port,
        workers=workers,
        ssl_keyfile=cfg.server_key_file,
        ssl_certfile=cfg.server_cert_file,
        ssl_ca_certs=cfg.client_ca_file,
        ssl_cert_reqs=ssl.CERT_REQUIRED,
        log_level="info",
        access_log=True,
    )

    raise SystemExit(0)
