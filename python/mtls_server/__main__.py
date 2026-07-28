"""CLI entry point for python -m mtls_server."""

import logging
import sys

from .config import load_config
from .server import run_server
from .tls import build_ssl_context

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%dT%H:%M:%S%z",
)
logger = logging.getLogger("mtls_server")


def main() -> None:
    config = load_config()
    ssl_ctx = build_ssl_context(
        cert_file=config.server_cert_file,
        key_file=config.server_key_file,
        ca_file=config.client_ca_file,
    )
    run_server(
        addr=config.listen_addr,
        ssl_ctx=ssl_ctx,
    )


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        logger.error("fatal: %s", e)
        sys.exit(1)
