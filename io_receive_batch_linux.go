//go:build linux && !386 && !s390x

package udp

import (
	"runtime"
	"syscall"
	"unsafe"

	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

type datagramBatchReader struct {
	messages  [maxDatagramReadBatch]receivedDatagram
	buffers   [maxDatagramReadBatch]buffer.ByteBuf
	headers   [maxDatagramReadBatch]multiMessageHeader
	iovecs    [maxDatagramReadBatch]unix.Iovec
	addresses [maxDatagramReadBatch]unix.RawSockaddrAny
}

// receive 由单一 EventLoop 调用并复用系统调用描述符，成功返回的 payload 所有权交给调用方。
func (r *datagramBatchReader) receive(fd transport.FDRef, alloc buffer.Allocator, size int, limit int) ([]receivedDatagram, bool, error) {
	limit = min(limit, maxDatagramReadBatch)
	if limit <= 0 {
		return nil, true, nil
	}
	prepared, err := r.prepare(alloc, size, limit)
	if err != nil {
		return nil, false, err
	}
	defer r.resetDescriptors(prepared)

	for {
		n, _, errno := syscall.Syscall6(
			unix.SYS_RECVMMSG,
			uintptr(fd.FD),
			uintptr(unsafe.Pointer(&r.headers[0])),
			uintptr(prepared),
			unix.MSG_DONTWAIT,
			0,
			0,
		)
		runtime.KeepAlive(r)
		if errno == unix.EINTR {
			continue
		}
		if rawWouldBlock(errno) {
			r.releaseBuffers(0, prepared)
			return nil, true, nil
		}
		if errno != 0 {
			r.releaseBuffers(0, prepared)
			return nil, false, errno
		}
		count := int(n)
		messages, err := r.unpack(count, prepared)
		return messages, count < prepared, err
	}
}

func (r *datagramBatchReader) prepare(alloc buffer.Allocator, size int, limit int) (int, error) {
	for index := 0; index < limit; index++ {
		payload, err := alloc.Acquire(size)
		if err != nil {
			r.releaseBuffers(0, index)
			return 0, err
		}
		r.buffers[index] = payload
		view := payload.WritableBytesView()
		r.iovecs[index].Base = unsafe.SliceData(view)
		r.iovecs[index].SetLen(len(view))
		r.headers[index].header.Name = (*byte)(unsafe.Pointer(&r.addresses[index]))
		r.headers[index].header.Namelen = uint32(unsafe.Sizeof(r.addresses[index]))
		r.headers[index].header.Iov = &r.iovecs[index]
		r.headers[index].header.Iovlen = 1
		r.headers[index].header.Flags = 0
		r.headers[index].length = 0
	}
	return limit, nil
}

func (r *datagramBatchReader) unpack(count int, prepared int) ([]receivedDatagram, error) {
	for index := 0; index < count; index++ {
		payload := r.buffers[index]
		n := min(int(r.headers[index].length), len(payload.WritableBytesView()))
		addr, err := socketAddressFromRawSockaddr(&r.addresses[index], r.headers[index].header.Namelen)
		if err != nil {
			r.releaseBuffers(index, prepared)
			return r.messages[:index], err
		}
		if err := payload.AdvanceWriter(n); err != nil {
			r.releaseBuffers(index, prepared)
			return r.messages[:index], err
		}
		r.messages[index] = receivedDatagram{payload: payload, addr: addr}
		r.buffers[index] = nil
	}
	r.releaseBuffers(count, prepared)
	return r.messages[:count], nil
}

func (r *datagramBatchReader) releaseBuffers(start int, end int) {
	for index := start; index < end; index++ {
		if r.buffers[index] != nil {
			r.buffers[index].Release()
			r.buffers[index] = nil
		}
	}
}

func (r *datagramBatchReader) resetDescriptors(count int) {
	for index := 0; index < count; index++ {
		r.headers[index].header.Name = nil
		r.headers[index].header.Iov = nil
		r.iovecs[index].Base = nil
	}
}
