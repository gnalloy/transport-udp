//go:build linux && !386 && !s390x

package udp

import (
	"testing"

	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func TestDatagramBatchReaderReceivesQueuedMessages(t *testing.T) {
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		t.Run(familyName(family), func(t *testing.T) {
			testDatagramBatchReader(t, family)
		})
	}
}

func TestDatagramBatchReaderReturnsDrainedForEmptySocket(t *testing.T) {
	receiver, _ := newLinuxDatagramSocket(t, unix.AF_INET)
	reader := datagramBatchReader{}
	messages, drained, err := reader.receive(transport.FDRef{FD: receiver}, buffer.NewHeapAllocator(), 64, 4)
	if err != nil || !drained || len(messages) != 0 {
		t.Fatalf("messages=%d drained=%t err=%v", len(messages), drained, err)
	}
}

func testDatagramBatchReader(t *testing.T, family int) {
	receiver, target := newLinuxDatagramSocket(t, family)
	sender, _ := newLinuxDatagramSocket(t, family)
	reader := datagramBatchReader{}
	alloc := buffer.NewHeapAllocator()

	for _, payloads := range [][]string{{"one", "two", "three"}, {"four", "five"}} {
		for _, payload := range payloads {
			again, err := sendDatagram(transport.FDRef{FD: sender}, testDatagram(t, []byte(payload), target))
			if err != nil || again {
				t.Fatalf("send payload=%q again=%t err=%v", payload, again, err)
			}
		}
		messages, drained, err := reader.receive(transport.FDRef{FD: receiver}, alloc, 64, 4)
		if err != nil || !drained || len(messages) != len(payloads) {
			t.Fatalf("messages=%d drained=%t err=%v", len(messages), drained, err)
		}
		for index, message := range messages {
			if got := string(message.payload.Bytes()); got != payloads[index] {
				t.Fatalf("payload[%d]=%q, want %q", index, got, payloads[index])
			}
			message.payload.Release()
			messages[index] = receivedDatagram{}
		}
	}
}

func familyName(family int) string {
	if family == unix.AF_INET6 {
		return "ipv6"
	}
	return "ipv4"
}
