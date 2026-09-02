//go:build linux

package udp

import (
	"errors"
	"net"
	"testing"
	"unsafe"

	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func TestRawDatagramIPv4RoundTrip(t *testing.T) {
	receiver, target := newLinuxDatagramSocket(t, unix.AF_INET)
	sender, _ := newLinuxDatagramSocket(t, unix.AF_INET)
	payload := []byte("gnalloy")

	again, err := sendDatagram(transport.FDRef{FD: sender}, testDatagram(t, payload, target))
	if err != nil || again {
		t.Fatalf("send again=%t err=%v", again, err)
	}
	dst := make([]byte, len(payload))
	n, from, again, err := recvDatagram(transport.FDRef{FD: receiver}, dst)
	if err != nil || again {
		t.Fatalf("receive again=%t err=%v", again, err)
	}
	if n != len(payload) || string(dst) != string(payload) {
		t.Fatalf("receive bytes=%d payload=%q", n, dst)
	}
	if !from.IP.IsLoopback() || from.Port <= 0 {
		t.Fatalf("source address=%v", from)
	}
}

func TestRawDatagramIPv6RoundTrip(t *testing.T) {
	receiver, target := newLinuxDatagramSocket(t, unix.AF_INET6)
	sender, _ := newLinuxDatagramSocket(t, unix.AF_INET6)
	payload := []byte("gnalloy-ipv6")

	again, err := sendDatagram(transport.FDRef{FD: sender}, testDatagram(t, payload, target))
	if err != nil || again {
		t.Fatalf("send again=%t err=%v", again, err)
	}
	dst := make([]byte, len(payload))
	n, from, again, err := recvDatagram(transport.FDRef{FD: receiver}, dst)
	if err != nil || again {
		t.Fatalf("receive again=%t err=%v", again, err)
	}
	if n != len(payload) || string(dst) != string(payload) {
		t.Fatalf("receive bytes=%d payload=%q", n, dst)
	}
	if !from.IP.Equal(net.IPv6loopback) || from.Port <= 0 {
		t.Fatalf("source address=%v", from)
	}
}

func TestRawDatagramReceiveReturnsAgain(t *testing.T) {
	fd, _ := newLinuxDatagramSocket(t, unix.AF_INET)
	n, addr, again, err := recvDatagram(transport.FDRef{FD: fd}, make([]byte, 1))
	if n != 0 || addr.Valid() || !again || err != nil {
		t.Fatalf("receive bytes=%d addr=%v again=%t err=%v", n, addr, again, err)
	}
}

func TestRawDatagramEmptyPayload(t *testing.T) {
	receiver, target := newLinuxDatagramSocket(t, unix.AF_INET)
	sender, _ := newLinuxDatagramSocket(t, unix.AF_INET)

	again, err := sendDatagram(transport.FDRef{FD: sender}, testDatagram(t, nil, target))
	if err != nil || again {
		t.Fatalf("send again=%t err=%v", again, err)
	}
	n, from, again, err := recvDatagram(transport.FDRef{FD: receiver}, make([]byte, 1))
	if err != nil || again || n != 0 || !from.Valid() {
		t.Fatalf("receive bytes=%d addr=%v again=%t err=%v", n, from, again, err)
	}
}

func TestRawSockaddrRejectsInvalidValues(t *testing.T) {
	if _, _, err := rawSockaddrFromAddress(Address{IP: net.IPv4zero, Port: -1}); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("invalid port error=%v", err)
	}
	var raw unix.RawSockaddrAny
	raw.Addr.Family = unix.AF_UNIX
	if _, err := addressFromRawSockaddr(&raw, uint32(unsafe.Sizeof(raw))); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("invalid family error=%v", err)
	}
	raw.Addr.Family = unix.AF_INET
	if _, err := addressFromRawSockaddr(&raw, unix.SizeofSockaddrInet4-1); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("short address error=%v", err)
	}
}

func newLinuxDatagramSocket(t *testing.T, family int) (int, Address) {
	t.Helper()
	fd, err := unix.Socket(family, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP)
	if err != nil {
		if family == unix.AF_INET6 && errors.Is(err, unix.EAFNOSUPPORT) {
			t.Skip("当前系统未启用 IPv6")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unix.Close(fd) })
	var bind unix.Sockaddr
	if family == unix.AF_INET6 {
		bind = &unix.SockaddrInet6{Addr: [16]byte{15: 1}}
	} else {
		bind = &unix.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}
	}
	if err := unix.Bind(fd, bind); err != nil {
		if family == unix.AF_INET6 && errors.Is(err, unix.EADDRNOTAVAIL) {
			t.Skip("当前系统未配置 IPv6 loopback")
		}
		t.Fatal(err)
	}
	name, err := unix.Getsockname(fd)
	if err != nil {
		t.Fatal(err)
	}
	return fd, unixSockaddrToAddress(name)
}

func testDatagram(t *testing.T, payload []byte, addr Address) Datagram {
	t.Helper()
	buf := buffer.NewHeapBuffer(len(payload))
	if _, err := buf.WriteBytes(payload); err != nil {
		buf.Release()
		t.Fatal(err)
	}
	t.Cleanup(func() { buf.Release() })
	return Datagram{Payload: buf, Addr: addr}
}
