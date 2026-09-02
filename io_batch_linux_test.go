//go:build linux

package udp

import (
	"testing"

	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func TestSendDatagramBatchSendsAllMessages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		family int
	}{
		{name: "ipv4", family: unix.AF_INET},
		{name: "ipv6", family: unix.AF_INET6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testSendDatagramBatch(t, tc.family)
		})
	}
}

func testSendDatagramBatch(t *testing.T, family int) {
	receiver, target := newLinuxDatagramSocket(t, family)
	sender, _ := newLinuxDatagramSocket(t, family)
	datagrams := []Datagram{
		testDatagram(t, []byte("one"), target),
		testDatagram(t, []byte("two"), target),
		testDatagram(t, []byte("three"), target),
	}

	sent, again, err := sendDatagramBatch(transport.FDRef{FD: sender}, datagrams)
	if err != nil || again || sent != len(datagrams) {
		t.Fatalf("sent=%d again=%t err=%v", sent, again, err)
	}
	buf := make([]byte, 16)
	for index, want := range []string{"one", "two", "three"} {
		n, _, again, err := recvDatagram(transport.FDRef{FD: receiver}, buf)
		if err != nil || again {
			t.Fatalf("receive index=%d again=%t err=%v", index, again, err)
		}
		if got := string(buf[:n]); got != want {
			t.Fatalf("receive index=%d payload=%q, want %q", index, got, want)
		}
	}
}
