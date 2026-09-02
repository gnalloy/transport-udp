//go:build !linux || 386 || s390x

package udp

import (
	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
)

type datagramBatchReader struct {
	messages [maxDatagramReadBatch]receivedDatagram
}

func (r *datagramBatchReader) receive(fd transport.FDRef, alloc buffer.Allocator, size int, limit int) ([]receivedDatagram, bool, error) {
	limit = min(limit, maxDatagramReadBatch)
	if limit <= 0 {
		return nil, true, nil
	}
	for index := 0; index < limit; index++ {
		payload, err := alloc.Acquire(size)
		if err != nil {
			return r.messages[:index], false, err
		}
		n, addr, again, err := recvDatagram(fd, payload.WritableBytesView())
		if again || err != nil {
			payload.Release()
			return r.messages[:index], again, err
		}
		if err := payload.AdvanceWriter(n); err != nil {
			payload.Release()
			return r.messages[:index], false, err
		}
		if !addr.Valid() {
			payload.Release()
			return r.messages[:index], true, nil
		}
		r.messages[index] = receivedDatagram{payload: payload, addr: addr}
	}
	return r.messages[:limit], false, nil
}
