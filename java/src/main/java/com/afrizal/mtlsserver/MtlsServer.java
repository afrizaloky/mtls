package com.afrizal.mtlsserver;

import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.logging.LoggingHandler;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.ClientAuth;
import java.util.concurrent.CountDownLatch;
import javax.net.ssl.SSLException;
import java.io.IOException;

public final class MtlsServer implements AutoCloseable {
  private final EventLoopGroup bossGroup;
  private final EventLoopGroup workerGroup;
  private final Channel serverChannel;
  private final CountDownLatch closed = new CountDownLatch(1);

  private MtlsServer(EventLoopGroup bossGroup, EventLoopGroup workerGroup, Channel serverChannel) {
    this.bossGroup = bossGroup;
    this.workerGroup = workerGroup;
    this.serverChannel = serverChannel;
  }

  public static MtlsServer start(Config config, int maxRequestSize) throws SSLException, InterruptedException {
    SslContext sslCtx = SslContextBuilder.forServer(config.serverCert().toFile(), config.serverKey().toFile())
      .trustManager(config.clientCa().toFile())
      .clientAuth(ClientAuth.REQUIRE)
      .protocols("TLSv1.3")
      .build();
    EventLoopGroup boss = new NioEventLoopGroup(1);
    EventLoopGroup workers = new NioEventLoopGroup();
    ServerBootstrap b = new ServerBootstrap()
      .group(boss, workers)
      .channel(NioServerSocketChannel.class)
      .handler(new LoggingHandler(LogLevel.INFO))
      .childHandler(new ChannelInitializer<SocketChannel>() {
        @Override protected void initChannel(SocketChannel ch) {
          ch.pipeline().addLast(sslCtx.newHandler(ch.alloc()));
          ch.pipeline().addLast(new HttpServerCodec());
          ch.pipeline().addLast(new HttpObjectAggregator(maxRequestSize));
          ch.pipeline().addLast(new HealthHandler());
        }
      });
    Channel ch = b.bind(config.listenAddress()).sync().channel();
    return new MtlsServer(boss, workers, ch);
  }

  /** Blocks until shutdown completes. */
  public void waitForClose() throws InterruptedException { closed.await(); }

  /** Initiates graceful shutdown and waits for completion. */
  @Override public void close() throws InterruptedException {
    serverChannel.close().syncUninterruptibly();
    bossGroup.shutdownGracefully().sync();
    workerGroup.shutdownGracefully().sync();
    closed.countDown();
  }
}