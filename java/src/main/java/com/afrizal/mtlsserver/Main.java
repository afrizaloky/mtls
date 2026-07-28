package com.afrizal.mtlsserver;

import java.util.logging.Level;
import java.util.logging.Logger;

public final class Main {
  private static final Logger log = Logger.getLogger(Main.class.getName());
  static final int MAX_REQUEST_SIZE = 65536;

  public static void main(String[] args) throws Exception {
    Config config = Config.fromEnvironment();
    log.info("starting server on " + config.listenAddress());
    try (MtlsServer server = MtlsServer.start(config, MAX_REQUEST_SIZE)) {
      Runtime.getRuntime().addShutdownHook(new Thread(() -> {
        log.info("shutting down");
        try { server.close(); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
      }, "shutdown"));
      server.waitForClose();
    }
  }
}