package udp

import (
	"gnalloy.org/gnalloy/buffer"
	"gnalloy.org/gnalloy/transport"
)

const defaultReadBufferSize = 2048

// AllocatorFactory 为每个 UDP endpoint 创建专属 ByteBuf 分配器。
type AllocatorFactory func(loop *transport.EventLoop) (buffer.Allocator, error)

type Config struct {
	ReuseAddr bool
	ReusePort bool
	// PooledInboundDatagrams 使用可复用指针承载入站报文，处理器接收 *Datagram 并负责转移或释放其所有权。
	PooledInboundDatagrams bool
	ReadBufferSize         int
	WriteBufferWatermark   transport.WriteBufferWatermark

	AllocatorFactory AllocatorFactory
}

func DefaultConfig() Config {
	return Config{ReuseAddr: true, ReadBufferSize: defaultReadBufferSize}
}

func normalizeConfig(cfg Config) Config {
	def := DefaultConfig()
	if !cfg.ReuseAddr {
		cfg.ReuseAddr = def.ReuseAddr
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = def.ReadBufferSize
	}
	cfg.WriteBufferWatermark = transport.NormalizeWriteBufferWatermark(cfg.WriteBufferWatermark)
	return cfg
}

type socketOptions struct {
	reuseAddr              bool
	reusePort              bool
	pooledInboundDatagrams bool
	readBufferSize         int
	writeBufferWatermark   transport.WriteBufferWatermark
}

func (c Config) socketOptions() socketOptions {
	return socketOptions{
		reuseAddr:              c.ReuseAddr,
		reusePort:              c.ReusePort,
		pooledInboundDatagrams: c.PooledInboundDatagrams,
		readBufferSize:         c.ReadBufferSize,
		writeBufferWatermark:   c.WriteBufferWatermark,
	}
}
