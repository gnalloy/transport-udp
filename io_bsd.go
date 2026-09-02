//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package udp

import (
	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func recvDatagram(fd transport.FDRef, dst []byte) (int, Address, bool, error) {
	n, from, err := unix.Recvfrom(fd.FD, dst, 0)
	if isAgain(err) {
		return n, Address{}, true, nil
	}
	if err != nil {
		return n, Address{}, false, err
	}
	return n, unixSockaddrToAddress(from), false, nil
}

func sendDatagram(fd transport.FDRef, datagram Datagram) (bool, error) {
	sa, err := addressToUnixSockaddr(datagram.Addr)
	if err != nil {
		return false, err
	}
	err = unix.Sendto(fd.FD, datagram.Payload.Bytes(), 0, sa)
	if isAgain(err) {
		return true, nil
	}
	return false, err
}
