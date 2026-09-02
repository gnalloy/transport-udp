package udp

import (
	"sync"

	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
)

// Datagram 是 UDP Pipeline 的入站和出站消息。
type Datagram struct {
	Payload buffer.ByteBuf
	Addr    Address

	// pool 和 self 仅用于池化入站消息；self 让值接收器能够归还原始指针。
	pool *datagramPool
	self *Datagram
	ip   [16]byte
}

func (d Datagram) Release() {
	if d.Payload != nil {
		d.Payload.Release()
	}
	if d.pool != nil && d.self != nil {
		d.pool.release(d.self)
	}
}

func (d Datagram) Valid() bool {
	return d.Payload != nil && d.Addr.Port >= 0 && d.Addr.Port <= 65535 && d.Addr.IP != nil
}

type datagramPool struct {
	pool sync.Pool
}

func (p *datagramPool) acquire(payload buffer.ByteBuf, addr Address) *Datagram {
	value, _ := p.pool.Get().(*Datagram)
	if value == nil {
		value = &Datagram{}
	}
	value.Payload = payload
	value.Addr = addr
	value.pool = p
	value.self = value
	return value
}

func (p *datagramPool) acquireSocketAddress(payload buffer.ByteBuf, addr transport.SocketAddress) *Datagram {
	value := p.acquire(payload, Address{})
	value.Addr = socketAddressToAddressInto(addr, &value.ip)
	return value
}

func (p *datagramPool) release(value *Datagram) {
	value.Payload = nil
	value.Addr = Address{}
	value.pool = nil
	value.self = nil
	p.pool.Put(value)
}
