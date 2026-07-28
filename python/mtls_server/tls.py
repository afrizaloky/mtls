"""Build an SSLContext for mTLS (TLS 1.3, client cert required)."""

import ssl


def build_ssl_context(*, cert_file: str, key_file: str, ca_file: str) -> ssl.SSLContext:
    """Create a server-side SSLContext enforcing mTLS with TLS 1.3.

    Args:
        cert_file: Path to PEM server certificate.
        key_file:  Path to PEM server private key.
        ca_file:   Path to PEM CA bundle for client certificate verification.

    Returns:
        Configured SSLContext ready for wrapping sockets.

    Raises:
        ssl.SSLError: If certificate files are invalid.
        OSError:      If files cannot be read.
    """
    ctx = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    ctx.minimum_version = ssl.TLSVersion.TLSv1_3
    ctx.verify_mode = ssl.CERT_REQUIRED
    ctx.load_cert_chain(cert_file, key_file)
    ctx.load_verify_locations(cafile=ca_file)
    return ctx
