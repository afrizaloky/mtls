package com.afrizal.mtlsserver;

import static org.junit.jupiter.api.Assertions.*;
import org.junit.jupiter.api.Test;
import java.net.InetSocketAddress;
import java.nio.file.Path;

class ConfigTest {
  @Test void parseDefaultAddress() { InetSocketAddress a = Config.address(":8443"); assertEquals("0.0.0.0", a.getHostString()); assertEquals(8443, a.getPort()); }
  @Test void parseFullAddress() { InetSocketAddress a = Config.address("127.0.0.1:9443"); assertEquals("127.0.0.1", a.getHostString()); assertEquals(9443, a.getPort()); }
  @Test void parseIpv6Address() { InetSocketAddress a = Config.address("[::1]:8443"); assertEquals("0:0:0:0:0:0:0:1", a.getHostString()); assertEquals(8443, a.getPort()); }
  @Test void parseMinPort() { InetSocketAddress a = Config.address(":1"); assertEquals(1, a.getPort()); }
  @Test void parseMaxPort() { InetSocketAddress a = Config.address(":65535"); assertEquals(65535, a.getPort()); }
  @Test void rejectNegativePort() { assertThrows(IllegalArgumentException.class, () -> Config.address(":-1")); }
  @Test void rejectZeroPort() { assertThrows(IllegalArgumentException.class, () -> Config.address(":0")); }
  @Test void rejectPortAboveMax() { assertThrows(IllegalArgumentException.class, () -> Config.address(":65536")); }
  @Test void rejectEmpty() { assertThrows(IllegalArgumentException.class, () -> Config.address("")); }
  @Test void rejectNoColon() { assertThrows(IllegalArgumentException.class, () -> Config.address("host")); }

  @Test void fromEnvironmentFailsWhenMissing() { assertThrows(IllegalArgumentException.class, () -> Config.fromEnvironment()); }
}