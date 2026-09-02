//go:build linux

package udp

import (
	"runtime"
	"syscall"
	"unsafe"

	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

type multiMessageHeader struct {
	header unix.Msghdr
	length uint32
}

func sendDatagramBatch(fd transport.FDRef, datagrams []Datagram) (int, bool, error) {
	if len(datagrams) == 0 {
		return 0, false, nil
	}
	if len(datagrams) == 1 {
		again, err := sendDatagram(fd, datagrams[0])
		if err != nil || again {
			return 0, again, err
		}
		return 1, false, nil
	}
	if len(datagrams) > maxDatagramWriteBatch {
		datagrams = datagrams[:maxDatagramWriteBatch]
	}
	var headers [maxDatagramWriteBatch]multiMessageHeader
	var iovecs [maxDatagramWriteBatch]unix.Iovec
	var addresses [maxDatagramWriteBatch]unix.RawSockaddrAny
	for index := range datagrams {
		length, err := prepareMultiMessage(&headers[index], &iovecs[index], &addresses[index], datagrams[index])
		if err != nil {
			return 0, false, err
		}
		headers[index].header.Namelen = length
	}
	for {
		sent, _, errno := syscall.Syscall6(
			unix.SYS_SENDMMSG,
			uintptr(fd.FD),
			uintptr(unsafe.Pointer(&headers[0])),
			uintptr(len(datagrams)),
			0,
			0,
			0,
		)
		runtime.KeepAlive(datagrams)
		runtime.KeepAlive(headers)
		runtime.KeepAlive(iovecs)
		runtime.KeepAlive(addresses)
		if errno == unix.EINTR {
			continue
		}
		if rawWouldBlock(errno) {
			return 0, true, nil
		}
		if errno != 0 {
			return 0, false, errno
		}
		count := int(sent)
		return count, count < len(datagrams), nil
	}
}

func prepareMultiMessage(header *multiMessageHeader, iovec *unix.Iovec, raw *unix.RawSockaddrAny, datagram Datagram) (uint32, error) {
	address, length, err := rawSockaddrFromAddress(datagram.Addr)
	if err != nil {
		return 0, err
	}
	*raw = address
	payload := datagram.Payload.Bytes()
	if len(payload) != 0 {
		iovec.Base = unsafe.SliceData(payload)
	}
	iovec.SetLen(len(payload))
	header.header.Name = (*byte)(unsafe.Pointer(raw))
	header.header.Iov = iovec
	header.header.Iovlen = 1
	return length, nil
}
