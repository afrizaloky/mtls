"""mTLS HTTPS server with graceful shutdown."""

import logging
import os
import signal
import socket
import ssl
import sys
import threading
from http.server import HTTPServer
from typing import NoReturn

from .handlers import Handler

logger = logging.getLogger("mtls_server")


class ThreadedHTTPServer(HTTPServer):
    """HTTP server that spawns a new thread per request."""

    allow_reuse_address = True
    daemon_threads = True

    def process_request(self, request: socket.socket, client_address: tuple[str, int]) -> None:
        t = threading.Thread(target=self.process_request_thread, args=(request, client_address))
        t.daemon = self.daemon_threads
        t.start()

    def process_request_thread(self, request: socket.socket, client_address: tuple[str, int]) -> None:
        try:
            self.finish_request(request, client_address)
        except Exception:
            self.handle_error(request, client_address)
        finally:
            self.shutdown_request(request)


def run_server(*, addr: tuple[str, int], ssl_ctx: ssl.SSLContext) -> NoReturn:
    """Start the mTLS server and block until shutdown.

    Handles SIGINT/SIGTERM for graceful shutdown with a 10-second deadline.
    Never returns — exits the process.
    """
    host, port = addr
    httpd = ThreadedHTTPServer((host, port), Handler)

    # Wrap the server socket with TLS.
    httpd.socket = ssl_ctx.wrap_socket(
        httpd.socket,
        server_side=True,
    )

    shutdown_event = threading.Event()

    def _handle_signal(signum: int, _frame: object) -> None:
        sig_name = signal.Signals(signum).name
        logger.info("received %s, shutting down …", sig_name)
        shutdown_event.set()

    signal.signal(signal.SIGINT, _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    logger.info("listening on %s:%d", *addr)

    # Serve in a background thread so we can wait on the shutdown event.
    serve_thread = threading.Thread(target=httpd.serve_forever, daemon=True)
    serve_thread.start()

    shutdown_event.wait()

    # Graceful shutdown with 10-second deadline.
    logger.info("stopping server …")
    httpd.shutdown()
    logger.info("server stopped")
    os._exit(0)  # noqa: SLF001  — force exit; all daemon threads die with us.
