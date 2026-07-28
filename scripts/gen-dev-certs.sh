#!/usr/bin/env bash
set -euo pipefail

# gen-dev-certs.sh — Generate ephemeral CA, server, and client certificates
#                        for local development of mtls-server.
#
# USAGE:  ./scripts/gen-dev-certs.sh [output-dir]
#         Default output-dir: ./dev-certs
#
# WARNING: Development use only. Contains hard-coded subject fields,
#          no CRL/CDP, and CA key without password. DO NOT use in
#          production or expose the generated files publicly.
#
# OUTPUT:
#   <output-dir>/
#     ca.pem          CA certificate (PEM)
#     ca-key.pem      CA private key  (PEM, no password)
#     server.pem      Server certificate (PEM)
#     server-key.pem  Server private key (PEM, no password)
#     client.pem      Client certificate (PEM)
#     client-key.pem  Client private key (PEM, no password)
#     client.p12      Client PKCS#12 keystore (password: changeit)
#     env.sh          Source this to set env vars for mtls-server

OUT="${1:-dev-certs}"
mkdir -p "$OUT"

DAYS=3650                          # 10 years, dev only
CA_SUBJ="/CN=mtls-dev-root-ca"
SERVER_SUBJ="/CN=localhost"
CLIENT_SUBJ="/CN=mtls-dev-client"

# ---- CA ----
openssl ecparam -genkey -name prime256v1 -out "$OUT/ca-key.pem"
openssl req -x509 -new -key "$OUT/ca-key.pem" -days "$DAYS" \
  -subj "$CA_SUBJ" -out "$OUT/ca.pem" \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign"

# ---- Server ----
openssl ecparam -genkey -name prime256v1 -out "$OUT/server-key.pem"
openssl req -new -key "$OUT/server-key.pem" -subj "$SERVER_SUBJ" \
  -out "$OUT/server.csr"
openssl x509 -req -in "$OUT/server.csr" -CA "$OUT/ca.pem" -CAkey "$OUT/ca-key.pem" \
  -CAcreateserial -days "$DAYS" -out "$OUT/server.pem" \
  -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1\nextendedKeyUsage=serverAuth")
rm -f "$OUT/server.csr"

# ---- Client ----
openssl ecparam -genkey -name prime256v1 -out "$OUT/client-key.pem"
openssl req -new -key "$OUT/client-key.pem" -subj "$CLIENT_SUBJ" \
  -out "$OUT/client.csr"
openssl x509 -req -in "$OUT/client.csr" -CA "$OUT/ca.pem" -CAkey "$OUT/ca-key.pem" \
  -CAcreateserial -days "$DAYS" -out "$OUT/client.pem" \
  -extfile <(printf "extendedKeyUsage=clientAuth")
rm -f "$OUT/client.csr"

# ---- PKCS#12 for JMeter / browsers ----
openssl pkcs12 -export \
  -in "$OUT/client.pem" -inkey "$OUT/client-key.pem" \
  -certfile "$OUT/ca.pem" \
  -out "$OUT/client.p12" -passout pass:changeit

# ---- Env file ----
cat > "$OUT/env.sh" <<ENVEOF
export MTLS_SERVER_CERT_FILE="$PWD/$OUT/server.pem"
export MTLS_SERVER_KEY_FILE="$PWD/$OUT/server-key.pem"
export MTLS_CLIENT_CA_FILE="$PWD/$OUT/ca.pem"
export MTLS_STATIC_KEY="\$(head -c 32 /dev/urandom | base64)"
export MTLS_LISTEN_ADDR=":8443"
ENVEOF

chmod 600 "$OUT/ca-key.pem" "$OUT/server-key.pem" "$OUT/client-key.pem"
rm -f "$OUT/ca.srl"

echo "=== Dev certs generated in $OUT ==="
echo "Source $OUT/env.sh then run: go run ./cmd/mtls-server"
echo "Test with:"
echo "  curl --cert $OUT/client.pem --key $OUT/client-key.pem"
echo "       --cacert $OUT/ca.pem"
echo "       https://localhost:8443/v1/echo"