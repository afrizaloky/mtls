"""HTTP request handlers."""

import json
import logging
from http.server import BaseHTTPRequestHandler

logger = logging.getLogger("mtls_server")


class Handler(BaseHTTPRequestHandler):
    """Minimal mTLS-only HTTP handler.

    Every request runs through the mTLS-protected socket, so client
    certificate verification happens before this class sees the request.
    """

    # Suppress default "logs a line per request" that duplicates our own logging.
    # We still get access-log-level info via log_message override if desired.
    def log_message(self, fmt: str, *args: object) -> None:
        logger.info(fmt, *args)

    # ---- routing -----------------------------------------------------------

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/v1/health":
            self._health()
        else:
            self._not_found()

    def do_POST(self) -> None:  # noqa: N802
        self._not_found()

    # ---- health ------------------------------------------------------------

    def _health(self) -> None:
        body = json.dumps({"status": "ok"}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    # ---- errors ------------------------------------------------------------

    def _not_found(self) -> None:
        self._write_error(404, "not found")

    def _write_error(self, status: int, msg: str) -> None:
        body = json.dumps({"error": msg}).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
