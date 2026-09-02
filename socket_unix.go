//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package udp

import (
	"errors"
	"net"
	"strconv"

	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

func makeUnixSockaddr(addr parsedAddress) (int, unix.Sockaddr, error) {
	if addr.ipv6 {
		sa := &unix.SockaddrInet6{Port: addr.port, ZoneId: zoneID(addr.zone)}
		copy(sa.Addr[:], addr.ip.To16())
		return unix.AF_INET6, sa, nil
	}
	sa := &unix.SockaddrInet4{Port: addr.port}
	copy(sa.Addr[:], addr.ip.To4())
	return unix.AF_INET, sa, nil
}

func addressToUnixSockaddr(addr Address) (unix.Sockaddr, error) {
	if addr.IP == nil || addr.Port < 0 || addr.Port > 65535 {
		return nil, ErrInvalidAddress
	}
	if ip4 := addr.IP.To4(); ip4 != nil {
		sa := &unix.SockaddrInet4{Port: addr.Port}
		copy(sa.Addr[:], ip4)
		return sa, nil
	}
	ip16 := addr.IP.To16()
	if ip16 == nil {
		return nil, ErrInvalidAddress
	}
	sa := &unix.SockaddrInet6{Port: addr.Port, ZoneId: zoneID(addr.Zone)}
	copy(sa.Addr[:], ip16)
	return sa, nil
}

func unixSockaddrToAddress(sa unix.Sockaddr) Address {
	return socketAddressToAddress(unixSockaddrToSocketAddress(sa))
}

func unixSockaddrToSocketAddress(sa unix.Sockaddr) transport.SocketAddress {
	var addr transport.SocketAddress
	switch v := sa.(type) {
	case *unix.SockaddrInet4:
		addr.Family = transport.SocketFamilyIPv4
		copy(addr.IP[:4], v.Addr[:])
		addr.Port = v.Port
		return addr
	case *unix.SockaddrInet6:
		addr.Family = transport.SocketFamilyIPv6
		copy(addr.IP[:], v.Addr[:])
		addr.Port = v.Port
		addr.ZoneID = v.ZoneId
		return addr
	default:
		return transport.SocketAddress{}
	}
}

func setSocketOptions(fd int, family int, opts socketOptions) error {
	if opts.reuseAddr {
		if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			return err
		}
	}
	if opts.reusePort {
		if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			return err
		}
	}
	if family == unix.AF_INET6 {
		if err := unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 1); err != nil {
			return err
		}
	}
	return nil
}

func closeFD(fd transport.FDRef) error {
	if !fd.Valid() {
		return nil
	}
	err := unix.Close(fd.FD)
	if errors.Is(err, unix.EBADF) {
		return nil
	}
	return err
}

func socketName(fd int, fallback string) string {
	sa, err := unix.Getsockname(fd)
	if err != nil {
		return fallback
	}
	switch v := sa.(type) {
	case *unix.SockaddrInet4:
		return net.JoinHostPort(net.IP(v.Addr[:]).String(), strconv.Itoa(v.Port))
	case *unix.SockaddrInet6:
		return net.JoinHostPort(net.IP(v.Addr[:]).String(), strconv.Itoa(v.Port))
	default:
		return fallback
	}
}

func isAgain(err error) bool {
	return errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK)
}
