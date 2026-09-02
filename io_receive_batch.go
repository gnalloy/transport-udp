package udp

import (
	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
)

const maxDatagramReadBatch = 64

type receivedDatagram struct {
	payload buffer.ByteBuf
	addr    transport.SocketAddress
}

func receiveSingleDatagram(fd transport.FDRef, alloc buffer.Allocator, size int) (receivedDatagram, bool, error) {
	payload, err := alloc.Acquire(size)
	if err != nil {
		return receivedDatagram{}, false, err
	}
	n, addr, again, err := recvDatagram(fd, payload.WritableBytesView())
	if again || err != nil {
		payload.Release()
		return receivedDatagram{}, again, err
	}
	if err := payload.AdvanceWriter(n); err != nil {
		payload.Release()
		return receivedDatagram{}, false, err
	}
	if !addr.Valid() {
		payload.Release()
		return receivedDatagram{}, true, nil
	}
	return receivedDatagram{payload: payload, addr: addr}, false, nil
}
