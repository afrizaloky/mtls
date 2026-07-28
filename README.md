# mtls-server

Minimal Go HTTPS server accepting only verified mTLS clients. Clients send AES-256-GCM encrypted payloads over TLS 1.3; server decrypts, re-encrypts with a fresh nonce, and returns the ciphertext.

No plaintext endpoint. No embedded keys. No third-party dependencies. Standard library only.

## Prerequisites

- Go 1.27+ (toolchain from `go.mod`)

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `MTLS_SERVER_CERT_FILE` | yes | - | Path to PEM-encoded server certificate |
| `MTLS_SERVER_KEY_FILE` | yes | - | Path to PEM-encoded server private key |
| `MTLS_CLIENT_CA_FILE` | yes | - | Path to PEM-encoded client CA bundle |
| `MTLS_STATIC_KEY` | yes | - | Padded base64-encoded 32-byte AES-256-GCM key |
| `MTLS_LISTEN_ADDR` | no | `:8443` | Listen address (`host:port`) |

All config errors fail before the server starts listening. Error messages may name the failing variable but never reveal its value.

## Key Generation

Generate a 32-byte key and print its base64 encoding. This command writes nothing to disk.

```shell
head -c 32 /dev/urandom | base64
```

Or with OpenSSL:

```shell
openssl rand -base64 32
```

Set the output as `MTLS_STATIC_KEY`. Never commit the key or commit files containing it.

## Certificate Requirements

Three PEM files are required at startup:

- **Server certificate and key** (`MTLS_SERVER_CERT_FILE`, `MTLS_SERVER_KEY_FILE`): A valid TLS certificate pair. Loaded via `tls.LoadX509KeyPair`.
- **Client CA bundle** (`MTLS_CLIENT_CA_FILE`): One or more CA certificates used to verify client certificates. Must contain at least one valid PEM block.

The server enforces:

- TLS 1.3 minimum (`MinVersion: tls.VersionTLS13`)
- Client certificate required and verified (`ClientAuth: tls.RequireAndVerifyClientCert`)
- Rejects TLS 1.2, missing/untrusted/expired client certs at the handshake level

## Build

```shell
go build ./...
```

## Run

```shell
export MTLS_SERVER_CERT_FILE=/path/to/server.pem
export MTLS_SERVER_KEY_FILE=/path/to/server-key.pem
export MTLS_CLIENT_CA_FILE=/path/to/client-ca.pem
export MTLS_STATIC_KEY="$(head -c 32 /dev/urandom | base64)"

go run ./cmd/mtls-server
```

The server logs startup address to stderr as structured JSON via `log/slog`. Send SIGINT or SIGTERM for graceful shutdown with a 10-second deadline.

## API

### `POST /v1/echo`

**Request:**

```json
{
  "nonce": "<padded-base64>",
  "ciphertext": "<padded-base64>"
}
```

- `nonce`: Must decode to exactly 12 bytes (AES-GCM nonce size).
- `ciphertext`: Must decode to at least 16 bytes (GCM overhead). Authenticated payload.
- `Content-Type`: `application/json`.
- Maximum encoded body: 65,536 bytes (64 KiB).
- Duplicate and unknown JSON fields are rejected.

**Response (200 OK):**

```json
{
  "nonce": "<padded-base64>",
  "ciphertext": "<padded-base64>"
}
```

- Contains the same plaintext re-encrypted with a fresh `crypto/rand` nonce.
- The server never echoes the request nonce or request ciphertext.

**Error Statuses:**

| Status | Body `error` | When |
|---|---|---|
| 400 | `invalid request` | Malformed JSON, wrong base64, wrong nonce/ciphertext size, auth failure, trailing data |
| 404 | `not found` | Path other than `/v1/echo` |
| 405 | `method not allowed` | Method other than POST |
| 413 | `request too large` | Body exceeds 64 KiB |
| 415 | `unsupported media type` | Content-Type not `application/json` |
| 500 | `internal server error` | Unexpected server failure |

All error bodies are generic JSON: `{"error":"<message>"}`. The server never leaks plaintext, keys, nonces, ciphertext, or internal error details.

## Security Notes

- **TLS 1.3 only.** No TLS 1.2 or plaintext fallback.
- **mTLS required.** Every request must come from a client with a certificate chaining to the configured CA.
- **No plaintext on the wire.** The server accepts and returns only encrypted envelopes.
- **Nonce uniqueness is the sender's responsibility.** The server generates a fresh nonce per response but does not track request nonces.
- **No replay protection, key rotation, persistence, or additional authentication.** The server is a single-purpose encrypted echo.

## Tests

Tests are documented in the plan but not included in this implementation.

```shell
go test ./...
```

## Exclusions

This server explicitly does not include:

- Certificate generation or management
- Embedded secrets, sample keys, or default credentials
- Plaintext or unauthenticated endpoints
- Key rotation, derivation, or versioning
- Replay detection or nonce tracking
- Client identity checks beyond the TLS handshake
- Persistence, databases, or caches
- CORS, metrics, tracing, or health checks
- Configuration files, CLI flags, or hot reload
- Docker, Kubernetes, CI/CD, or deployment assets
- Third-party dependencies outside the Go standard library