//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !windows

package udp

import "gnalloy.org/gnalloy/transport"

func listenUDP(string, socketOptions) (udpSocket, error) {
	return udpSocket{}, transport.ErrUnsupportedPoller
}

func recvDatagram(transport.FDRef, []byte) (int, transport.SocketAddress, bool, error) {
	return 0, transport.SocketAddress{}, false, transport.ErrUnsupportedPoller
}

func sendDatagram(transport.FDRef, Datagram) (bool, error) {
	return false, transport.ErrUnsupportedPoller
}

func closeFD(transport.FDRef) error {
	return nil
}
