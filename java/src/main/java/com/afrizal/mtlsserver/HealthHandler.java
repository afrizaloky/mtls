package com.afrizal.mtlsserver;

import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import java.nio.charset.StandardCharsets;

final class HealthHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
  private static final byte[] OK = "{\"status\":\"ok\"}".getBytes(StandardCharsets.US_ASCII);
  private static final byte[] NOT_FOUND = "{\"error\":\"not found\"}".getBytes(StandardCharsets.US_ASCII);
  private static final byte[] METHOD = "{\"error\":\"method not allowed\"}".getBytes(StandardCharsets.US_ASCII);
  @Override protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
    if (!request.uri().equals("/v1/health")) { write(ctx, HttpResponseStatus.NOT_FOUND, NOT_FOUND, null); return; }
    if (!request.method().equals(HttpMethod.GET)) { write(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, METHOD, "GET"); return; }
    write(ctx, HttpResponseStatus.OK, OK, null);
  }
  private static void write(ChannelHandlerContext ctx, HttpResponseStatus status, byte[] body, String allow) { FullHttpResponse response=new DefaultFullHttpResponse(HttpVersion.HTTP_1_1,status,Unpooled.wrappedBuffer(body)); response.headers().set(HttpHeaderNames.CONTENT_TYPE,"application/json").setInt(HttpHeaderNames.CONTENT_LENGTH,body.length); if(allow!=null) response.headers().set(HttpHeaderNames.ALLOW,allow); ctx.writeAndFlush(response); }
}
