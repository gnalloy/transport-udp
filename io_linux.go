//go:build linux

package udp

import (
	"net"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"

	"gnalloy.org/gnalloy/transport"
	"golang.org/x/sys/unix"
)

// 直接解析内核 sockaddr，避免通用转换路径对每个数据报查询 SO_PROTOCOL。
func recvDatagram(fd transport.FDRef, dst []byte) (int, Address, bool, error) {
	var raw unix.RawSockaddrAny
	length := uint32(unsafe.Sizeof(raw))
	var data unsafe.Pointer
	if len(dst) != 0 {
		data = unsafe.Pointer(unsafe.SliceData(dst))
	}
	for {
		n, _, errno := syscall.Syscall6(
			unix.SYS_RECVFROM,
			uintptr(fd.FD),
			uintptr(data),
			uintptr(len(dst)),
			0,
			uintptr(unsafe.Pointer(&raw)),
			uintptr(unsafe.Pointer(&length)),
		)
		runtime.KeepAlive(dst)
		if errno == unix.EINTR {
			continue
		}
		if rawWouldBlock(errno) {
			return 0, Address{}, true, nil
		}
		if errno != 0 {
			return 0, Address{}, false, errno
		}
		addr, err := addressFromRawSockaddr(&raw, length)
		if err != nil {
			return int(n), Address{}, false, err
		}
		return int(n), addr, false, nil
	}
}

func sendDatagram(fd transport.FDRef, datagram Datagram) (bool, error) {
	raw, length, err := rawSockaddrFromAddress(datagram.Addr)
	if err != nil {
		return false, err
	}
	payload := datagram.Payload.Bytes()
	var data unsafe.Pointer
	if len(payload) != 0 {
		data = unsafe.Pointer(unsafe.SliceData(payload))
	}
	for {
		_, _, errno := syscall.Syscall6(
			unix.SYS_SENDTO,
			uintptr(fd.FD),
			uintptr(data),
			uintptr(len(payload)),
			0,
			uintptr(unsafe.Pointer(&raw)),
			uintptr(length),
		)
		runtime.KeepAlive(payload)
		runtime.KeepAlive(datagram)
		if errno == unix.EINTR {
			continue
		}
		if rawWouldBlock(errno) {
			return true, nil
		}
		if errno != 0 {
			return false, errno
		}
		return false, nil
	}
}

func rawSockaddrFromAddress(addr Address) (unix.RawSockaddrAny, uint32, error) {
	var raw unix.RawSockaddrAny
	if addr.IP == nil || addr.Port < 0 || addr.Port > 65535 {
		return raw, 0, ErrInvalidAddress
	}
	if ip4 := addr.IP.To4(); ip4 != nil {
		sa := (*unix.RawSockaddrInet4)(unsafe.Pointer(&raw))
		sa.Family = unix.AF_INET
		sa.Port = hostToNetwork16(uint16(addr.Port))
		copy(sa.Addr[:], ip4)
		return raw, unix.SizeofSockaddrInet4, nil
	}
	ip16 := addr.IP.To16()
	if ip16 == nil {
		return raw, 0, ErrInvalidAddress
	}
	sa := (*unix.RawSockaddrInet6)(unsafe.Pointer(&raw))
	sa.Family = unix.AF_INET6
	sa.Port = hostToNetwork16(uint16(addr.Port))
	sa.Scope_id = zoneID(addr.Zone)
	copy(sa.Addr[:], ip16)
	return raw, unix.SizeofSockaddrInet6, nil
}

func addressFromRawSockaddr(raw *unix.RawSockaddrAny, length uint32) (Address, error) {
	if raw == nil {
		return Address{}, ErrInvalidAddress
	}
	switch raw.Addr.Family {
	case unix.AF_INET:
		if length < unix.SizeofSockaddrInet4 {
			return Address{}, ErrInvalidAddress
		}
		sa := (*unix.RawSockaddrInet4)(unsafe.Pointer(raw))
		ip := make(net.IP, net.IPv4len)
		copy(ip, sa.Addr[:])
		return Address{IP: ip, Port: int(networkToHost16(sa.Port))}, nil
	case unix.AF_INET6:
		if length < unix.SizeofSockaddrInet6 {
			return Address{}, ErrInvalidAddress
		}
		sa := (*unix.RawSockaddrInet6)(unsafe.Pointer(raw))
		ip := make(net.IP, net.IPv6len)
		copy(ip, sa.Addr[:])
		zone := ""
		if sa.Scope_id != 0 {
			zone = strconv.FormatUint(uint64(sa.Scope_id), 10)
		}
		return Address{IP: ip, Port: int(networkToHost16(sa.Port)), Zone: zone}, nil
	default:
		return Address{}, ErrInvalidAddress
	}
}

func hostToNetwork16(value uint16) uint16 {
	return value<<8 | value>>8
}

func networkToHost16(value uint16) uint16 {
	return hostToNetwork16(value)
}

func rawWouldBlock(errno syscall.Errno) bool {
	return errno == unix.EAGAIN || errno == unix.EWOULDBLOCK
}
