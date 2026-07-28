# Java mTLS Health Server

`GET /v1/health` mTLS server with Netty 4.1.x NIO on Java 17+.

## Build

```bash
mvn -f java/pom.xml clean package
```

Produces a shaded (fat) JAR at `java/target/mtls-server-1.0.0.jar`.

## Run

Set env vars (same as Go/Python modules):

| Env var | Required | Description |
|---|---|---|
| `MTLS_SERVER_CERT_FILE` | yes | Path to server certificate PEM |
| `MTLS_SERVER_KEY_FILE` | yes | Path to server private key PEM |
| `MTLS_CLIENT_CA_FILE` | yes | Path to CA PEM that signed client certs |
| `MTLS_LISTEN_ADDR` | no | Default `:8443` (host:port or :port) |

```bash
source dev-certs/env.sh
java -jar java/target/mtls-server-1.0.0.jar
```

## Test

```bash
mvn -f java/pom.xml test
```

## Verify with curl

```bash
curl --cacert dev-certs/ca.pem \
     --cert dev-certs/client.pem \
     --key dev-certs/client-key.pem \
     https://localhost:8443/v1/health
```

Expected: `{"status":"ok"}`

## Design

- **Netty NIO** — `NioEventLoopGroup`, `NioServerSocketChannel`
- **TLSv1.3 only** — enforced via `SslContextBuilder.protocols("TLSv1.3")`
- **Required client cert** — `ClientAuth.REQUIRE` with CA trust store
- **Bounded HTTP** — `HttpObjectAggregator` with 64 KiB max request size
- **Graceful shutdown** — `Runtime.addShutdownHook` closes event loops
- **Shaded JAR** — single executable with all dependencies

No echo endpoint, no `MTLS_STATIC_KEY`, no native/OpenSSL dependency.