package com.afrizal.mtlsserver;

import java.net.InetSocketAddress;
import java.nio.file.Files;
import java.nio.file.Path;

public record Config(Path serverCert, Path serverKey, Path clientCa, InetSocketAddress listenAddress) {
  public static Config fromEnvironment() {
    return new Config(requiredFile("MTLS_SERVER_CERT_FILE"), requiredFile("MTLS_SERVER_KEY_FILE"), requiredFile("MTLS_CLIENT_CA_FILE"), address(System.getenv().getOrDefault("MTLS_LISTEN_ADDR", ":8443")));
  }
  private static Path requiredFile(String name) { String value = System.getenv(name); if (value == null || value.isBlank()) throw new IllegalArgumentException(name + " is required"); Path p = Path.of(value); if (!Files.isRegularFile(p) || !Files.isReadable(p)) throw new IllegalArgumentException(name + " must name a readable file"); return p; }
  static InetSocketAddress address(String value) { if (value.startsWith(":")) return new InetSocketAddress("0.0.0.0", port(value.substring(1))); int colon=value.lastIndexOf(':'); if (colon < 1) throw new IllegalArgumentException("MTLS_LISTEN_ADDR is invalid"); return new InetSocketAddress(value.substring(0, colon), port(value.substring(colon+1))); }
  private static int port(String value) { try { int p=Integer.parseInt(value); if(p<1||p>65535) throw new NumberFormatException(); return p; } catch(NumberFormatException e) { throw new IllegalArgumentException("MTLS_LISTEN_ADDR is invalid", e); } }
}
