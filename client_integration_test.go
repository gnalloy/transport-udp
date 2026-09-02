package udp_test

import (
	"context"
	"testing"
	"time"

	"gnalloy.org/gnalloy/bootstrap"
	"gnalloy.org/gnalloy/channel"
	"gnalloy.org/gnalloy/transport"
	"gnalloy.org/transport-udp"
)

func TestUDPDialerEchoUsesDefaultRemote(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pooled bool
	}{
		{name: "value"},
		{name: "pooled-pointer", pooled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testUDPDialerEcho(t, tc.pooled)
		})
	}
}

func testUDPDialerEcho(t *testing.T, pooled bool) {
	serverGroup, err := transport.NewEventLoopGroup(transport.EventLoopGroupConfig{
		Size:         1,
		PollerConfig: transport.Config{Backend: transport.DefaultBackend()},
	})
	if err != nil {
		t.Fatal(err)
	}
	clientGroup, err := transport.NewEventLoopGroup(transport.EventLoopGroupConfig{
		Size:         1,
		PollerConfig: transport.Config{Backend: transport.DefaultBackend()},
	})
	if err != nil {
		_ = serverGroup.Close()
		t.Fatal(err)
	}
	defer shutdownGroup(t, serverGroup)
	defer shutdownGroup(t, clientGroup)

	server, err := bootstrap.NewServerBootstrap().
		Group(serverGroup, serverGroup).
		Transport(udp.NewTransport(udp.DefaultConfig())).
		ChildInitializer(func(ch channel.Channel) error {
			return ch.Pipeline().AddLast("echo", udpDatagramEchoHandler{})
		}).
		Bind("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	recorder := newUDPClientRecorder()
	clientConfig := udp.DefaultConfig()
	clientConfig.PooledInboundDatagrams = pooled
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch, err := bootstrap.NewDialer().
		Group(clientGroup).
		Transport(udp.NewTransport(clientConfig)).
		Initializer(func(ch channel.Channel) error {
			return ch.Pipeline().AddLast("recorder", recorder)
		}).
		DialContext(ctx, server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	out, err := ch.Allocator().Acquire(4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := out.WriteBytes([]byte("ping")); err != nil {
		out.Release()
		t.Fatal(err)
	}
	if err := ch.WriteAndFlush(out); err != nil {
		t.Fatal(err)
	}
	recorder.waitPayload(t, "ping")
}

type udpDatagramEchoHandler struct{}

func (udpDatagramEchoHandler) ChannelRead(ctx *channel.HandlerContext, msg any) {
	datagram, ok := msg.(udp.Datagram)
	if !ok {
		ctx.FireChannelRead(msg)
		return
	}
	if err := ctx.Channel().WriteAndFlush(datagram); err != nil {
		ctx.FireExceptionCaught(err)
	}
}

type udpClientRecorder struct {
	payloads chan string
}

func newUDPClientRecorder() *udpClientRecorder {
	return &udpClientRecorder{payloads: make(chan string, 1)}
}

func (r *udpClientRecorder) ChannelRead(_ *channel.HandlerContext, msg any) {
	var datagram udp.Datagram
	switch value := msg.(type) {
	case udp.Datagram:
		datagram = value
	case *udp.Datagram:
		if value == nil {
			return
		}
		datagram = *value
	default:
		return
	}
	defer datagram.Release()
	select {
	case r.payloads <- string(datagram.Payload.Bytes()):
	default:
	}
}

func (r *udpClientRecorder) waitPayload(t *testing.T, want string) {
	t.Helper()
	select {
	case got := <-r.payloads:
		if got != want {
			t.Fatalf("payload=%q, want %q", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for payload %q", want)
	}
}
