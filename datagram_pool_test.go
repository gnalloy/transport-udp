package udp

import (
	"net"
	"testing"

	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/channel"
	"gnalloy.org/gnalloy/transport"
)

func TestDatagramPoolClearsReleasedEnvelope(t *testing.T) {
	var pool datagramPool
	buf := buffer.NewHeapBuffer(32)
	datagram := pool.acquire(buf, Address{IP: net.IPv4(127, 0, 0, 1), Port: 9000})
	if !datagram.Valid() {
		t.Fatal("pooled datagram is invalid")
	}
	datagram.Release()
	if datagram.Payload != nil || datagram.Addr.IP != nil || datagram.pool != nil || datagram.self != nil {
		t.Fatalf("released datagram=%+v", datagram)
	}
}

func TestDatagramPoolReusesEnvelopeWithoutAllocating(t *testing.T) {
	var pool datagramPool
	datagram := pool.acquire(nil, Address{})
	pool.release(datagram)
	allocs := testing.AllocsPerRun(1000, func() {
		value := pool.acquire(nil, Address{})
		pool.release(value)
	})
	if allocs != 0 {
		t.Fatalf("allocations=%f, want 0", allocs)
	}
}

func TestPooledDatagramPointerReleaseIsIdempotent(t *testing.T) {
	var pool datagramPool
	buf := buffer.NewHeapBuffer(32)
	datagram := pool.acquire(buf, Address{IP: net.IPv4(127, 0, 0, 1), Port: 9000})

	datagram.Release()
	datagram.Release()

	if buf.RefCnt() != 0 {
		t.Fatalf("ref=%d, want 0", buf.RefCnt())
	}
}

func TestEndpointDefaultInboundDatagramUsesValueContract(t *testing.T) {
	collector := &udpCaptureInbound{}
	ep := newInboundTestEndpoint(t, collector, false)
	buf := newUDPTestBuf("value")

	ep.fireChannelRead(buf, testAddr)

	if len(collector.msgs) != 1 {
		t.Fatalf("messages=%d, want 1", len(collector.msgs))
	}
	if _, ok := collector.msgs[0].(Datagram); !ok {
		t.Fatalf("message=%T, want Datagram", collector.msgs[0])
	}
	releaseMessage(collector.msgs[0])
}

func TestEndpointPooledInboundDatagramUsesPointerContract(t *testing.T) {
	collector := &udpCaptureInbound{}
	ep := newInboundTestEndpoint(t, collector, true)
	buf := newUDPTestBuf("pointer")

	ep.fireChannelRead(buf, testAddr)

	if len(collector.msgs) != 1 {
		t.Fatalf("messages=%d, want 1", len(collector.msgs))
	}
	datagram, ok := collector.msgs[0].(*Datagram)
	if !ok {
		t.Fatalf("message=%T, want *Datagram", collector.msgs[0])
	}
	datagram.Release()
	if buf.RefCnt() != 0 {
		t.Fatalf("ref=%d, want 0", buf.RefCnt())
	}
}

func TestEndpointPooledCompletionReleasesAtPipelineTail(t *testing.T) {
	ep := newInboundTestEndpoint(t, nil, true)
	ep.closed.Store(true)
	buf := newUDPTestBuf("completion")

	ep.handleReadCompletion(transport.PollEvent{
		Op:  transport.OpRead,
		N:   buf.ReadableBytes(),
		Buf: buf,
		Addr: transport.SocketAddress{
			Family: transport.SocketFamilyIPv4,
			IP:     [16]byte{127, 0, 0, 1},
			Port:   9000,
		},
	})

	if buf.RefCnt() != 0 {
		t.Fatalf("ref=%d, want 0", buf.RefCnt())
	}
}

func newInboundTestEndpoint(t *testing.T, collector *udpCaptureInbound, pooled bool) *endpoint {
	t.Helper()
	ch := channel.NewLocalChannel(1, buffer.NewHeapAllocator(), nil)
	if collector != nil {
		if err := ch.Pipeline().AddLast("collector", collector); err != nil {
			t.Fatal(err)
		}
	}
	return &endpoint{ch: ch, pooledInboundDatagrams: pooled}
}
