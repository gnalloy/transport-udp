//go:build linux

package udp

import (
	"testing"

	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func TestSendDatagramBatchSendsAllMessages(t *testing.T) {
	for _, tc := range []struct {
		name     string
		family   int
		payloads []string
	}{
		{name: "ipv4_single", family: unix.AF_INET, payloads: []string{"one"}},
		{name: "ipv4_batch", family: unix.AF_INET, payloads: []string{"one", "two", "three"}},
		{name: "ipv6_single", family: unix.AF_INET6, payloads: []string{"one"}},
		{name: "ipv6_batch", family: unix.AF_INET6, payloads: []string{"one", "two", "three"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testSendDatagramBatch(t, tc.family, tc.payloads)
		})
	}
}

func testSendDatagramBatch(t *testing.T, family int, payloads []string) {
	receiver, target := newLinuxDatagramSocket(t, family)
	sender, _ := newLinuxDatagramSocket(t, family)
	datagrams := make([]Datagram, len(payloads))
	for index, payload := range payloads {
		datagrams[index] = testDatagram(t, []byte(payload), target)
	}

	sent, again, err := sendDatagramBatch(transport.FDRef{FD: sender}, datagrams)
	if err != nil || again || sent != len(datagrams) {
		t.Fatalf("sent=%d again=%t err=%v", sent, again, err)
	}
	buf := make([]byte, 16)
	for index, want := range payloads {
		n, _, again, err := recvDatagram(transport.FDRef{FD: receiver}, buf)
		if err != nil || again {
			t.Fatalf("receive index=%d again=%t err=%v", index, again, err)
		}
		if got := string(buf[:n]); got != want {
			t.Fatalf("receive index=%d payload=%q, want %q", index, got, want)
		}
	}
}
