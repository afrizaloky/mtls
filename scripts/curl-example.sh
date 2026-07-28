#!/usr/bin/env bash
set -euo pipefail

# Send plaintext bytes to local dev server, receive encrypted envelope.
# Optionally decrypt response when MTLS_STATIC_KEY is set.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CERT_DIR="${CERT_DIR:-$ROOT/dev-certs}"
URL="${URL:-https://localhost:8443/v1/echo}"
PLAINTEXT="${1:-hello from curl}"

RESPONSE_FILE="$(mktemp)"
trap 'rm -f "$RESPONSE_FILE"' EXIT

curl --fail-with-body --silent --show-error \
  --cert "$CERT_DIR/client.pem" \
  --key "$CERT_DIR/client-key.pem" \
  --cacert "$CERT_DIR/ca.pem" \
  --header 'Content-Type: application/octet-stream' \
  --data-binary "$PLAINTEXT" \
  "$URL" > "$RESPONSE_FILE"

echo "encrypted response:"
cat "$RESPONSE_FILE"
echo

# Decrypt response if MTLS_STATIC_KEY is set
if [ -n "${MTLS_STATIC_KEY:-}" ]; then
	echo "decrypted:"
	MTLS_STATIC_KEY="$MTLS_STATIC_KEY" \
	  go run "$ROOT/scripts/decrypt-response.go" "$RESPONSE_FILE"
fi