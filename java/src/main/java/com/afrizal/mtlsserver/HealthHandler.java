package com.afrizal.mtlsserver;

import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import javax.crypto.Cipher;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.SecretKeySpec;

final class HealthHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
  private static final byte[] OK = "{\"status\":\"ok\"}".getBytes(StandardCharsets.US_ASCII);
  private static final byte[] NOT_FOUND = "{\"error\":\"not found\"}".getBytes(StandardCharsets.US_ASCII);
  private static final byte[] METHOD = "{\"error\":\"method not allowed\"}".getBytes(StandardCharsets.US_ASCII);
  private static final byte[] ERROR = "{\"error\":\"internal server error\"}".getBytes(StandardCharsets.US_ASCII);
  private static final SecureRandom RNG = new SecureRandom();

  @Override protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
    if (!"/v1/health".equals(request.uri())) { write(ctx, HttpResponseStatus.NOT_FOUND, NOT_FOUND, null); return; }
    if (!HttpMethod.GET.equals(request.method())) { write(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, METHOD, "GET"); return; }
    try {
      var key = new byte[32];
      var nonce = new byte[12];
      var plaintext = new byte[256];
      RNG.nextBytes(key);
      RNG.nextBytes(nonce);
      RNG.nextBytes(plaintext);
      var cipher = Cipher.getInstance("AES/GCM/NoPadding");
      cipher.init(Cipher.ENCRYPT_MODE, new SecretKeySpec(key, "AES"), new GCMParameterSpec(128, nonce));
      cipher.doFinal(plaintext);
    } catch (Exception e) {
      write(ctx, HttpResponseStatus.INTERNAL_SERVER_ERROR, ERROR, null);
      return;
    }
    write(ctx, HttpResponseStatus.OK, OK, null);
  }

  private static void write(ChannelHandlerContext ctx, HttpResponseStatus status, byte[] body, String allow) {
    var response = new DefaultFullHttpResponse(HttpVersion.HTTP_1_1, status, Unpooled.wrappedBuffer(body));
    response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json").setInt(HttpHeaderNames.CONTENT_LENGTH, body.length);
    if (allow != null) response.headers().set(HttpHeaderNames.ALLOW, allow);
    ctx.writeAndFlush(response);
  }
}
