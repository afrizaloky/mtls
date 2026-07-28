"""Typed configuration loaded from environment variables."""

import dataclasses
import os
import socket
import struct


@dataclasses.dataclass(frozen=True)
class Config:
    """Immutable server configuration.

    Attributes match the env vars documented in the project README.
    """

    server_cert_file: str
    server_key_file: str
    client_ca_file: str
    listen_addr: tuple[str, int]


def _required_path(env_name: str) -> str:
    path = os.environ.get(env_name)
    if not path:
        raise ValueError(f"{env_name} is required")
    if not os.path.isfile(path):
        raise ValueError(f"{env_name} must name a readable file: {path}")
    return path


def _resolve_addr(host: str, port: int) -> tuple[str, int]:
    """Validate and return (host, port)."""
    if not (0 < port < 65536):
        raise ValueError(f"port out of range: {port}")
    try:
        socket.getaddrinfo(host, port, socket.AF_UNSPEC, socket.SOCK_STREAM)
    except socket.gaierror as exc:
        raise ValueError(f"cannot resolve {host}:{port}: {exc}") from exc
    return (host, port)


def load_config() -> Config:
    """Load configuration from environment variables.

    Required:
        MTLS_SERVER_CERT_FILE, MTLS_SERVER_KEY_FILE, MTLS_CLIENT_CA_FILE

    Optional:
        MTLS_LISTEN_ADDR (default :8443)

    Raises ValueError on any validation failure.  Error messages may name the
    failing variable but never reveal secret values.
    """
    server_cert_file = _required_path("MTLS_SERVER_CERT_FILE")
    server_key_file = _required_path("MTLS_SERVER_KEY_FILE")
    client_ca_file = _required_path("MTLS_CLIENT_CA_FILE")

    listen_addr_raw = os.environ.get("MTLS_LISTEN_ADDR", ":8443")
    if ":" in listen_addr_raw:
        host, _, port_str = listen_addr_raw.rpartition(":")
        if not host:
            host = "0.0.0.0"
        try:
            port = int(port_str)
        except ValueError:
            raise ValueError(f"MTLS_LISTEN_ADDR invalid port: {listen_addr_raw}")
        listen_addr = _resolve_addr(host, port)
    else:
        raise ValueError(f"MTLS_LISTEN_ADDR must be host:port: {listen_addr_raw}")

    return Config(
        server_cert_file=server_cert_file,
        server_key_file=server_key_file,
        client_ca_file=client_ca_file,
        listen_addr=listen_addr,
    )
