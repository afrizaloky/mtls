package com.afrizal.mtlsserver;

import io.netty.buffer.Unpooled;
import io.netty.channel.embedded.EmbeddedChannel;
import io.netty.handler.codec.http.*;
import static org.junit.jupiter.api.Assertions.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import java.nio.charset.StandardCharsets;

class HealthHandlerTest {
  private EmbeddedChannel channel;

  @BeforeEach void setUp() { channel = new EmbeddedChannel(new HttpRequestDecoder(), new HttpObjectAggregator(65536), new HealthHandler()); }

  @Test void getHealthReturnsOk() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/v1/health"));
    var response = channel.readOutbound();
    assertInstanceOf(FullHttpResponse.class, response);
    FullHttpResponse resp = (FullHttpResponse) response;
    assertEquals(HttpResponseStatus.OK, resp.status());
    assertEquals("application/json", resp.headers().get(HttpHeaderNames.CONTENT_TYPE));
    assertEquals("{\"status\":\"ok\"}", resp.content().toString(StandardCharsets.US_ASCII));
  }

  @Test void getHealthContentLengthIsCorrect() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/v1/health"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(response.content().readableBytes(), response.headers().getInt(HttpHeaderNames.CONTENT_LENGTH));
  }

  @Test void unknownPathReturns404() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/v1/unknown"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.NOT_FOUND, response.status());
    assertEquals("application/json", response.headers().get(HttpHeaderNames.CONTENT_TYPE));
    assertEquals("{\"error\":\"not found\"}", response.content().toString(StandardCharsets.US_ASCII));
  }

  @Test void postToHealthReturns405() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.POST, "/v1/health"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.METHOD_NOT_ALLOWED, response.status());
    assertEquals("GET", response.headers().get(HttpHeaderNames.ALLOW));
    assertEquals("application/json", response.headers().get(HttpHeaderNames.CONTENT_TYPE));
    assertEquals("{\"error\":\"method not allowed\"}", response.content().toString(StandardCharsets.US_ASCII));
  }

  @Test void putToHealthReturns405() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.PUT, "/v1/health"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.METHOD_NOT_ALLOWED, response.status());
    assertEquals("GET", response.headers().get(HttpHeaderNames.ALLOW));
  }

  @Test void deleteToHealthReturns405() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.DELETE, "/v1/health"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.METHOD_NOT_ALLOWED, response.status());
    assertEquals("GET", response.headers().get(HttpHeaderNames.ALLOW));
  }

  @Test void rootPathReturns404() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.NOT_FOUND, response.status());
  }

  @Test void healthWithQueryReturns404() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/v1/health?foo=bar"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.NOT_FOUND, response.status());
  }

  @Test void healthWithFragmentReturns404() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/v1/health#section"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.NOT_FOUND, response.status());
  }

  @Test void unknownMethodReturns405() {
    channel.writeInbound(new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.OPTIONS, "/v1/health"));
    var response = (FullHttpResponse) channel.readOutbound();
    assertEquals(HttpResponseStatus.METHOD_NOT_ALLOWED, response.status());
  }
}