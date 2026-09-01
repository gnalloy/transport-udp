# API Reference

[简体中文](api.zh-CN.md) | [Docs Index](README.md)

This inventory is generated from `go doc -short` for the packages in this repository. It is a quick public-surface map; source files and tests remain the authority for exact semantics.

## Packages

### `gnalloy.org/transport-udp`

Package name: `udp`

```text
var ErrInvalidAddress = errors.New("gnalloy/transport/udp: invalid address") ...
type Address struct{ ... }
type Addressed struct{ ... }
type AddressedMessageEncoder interface{ ... }
type AllocatorFactory func(loop *transport.EventLoop) (buffer.Allocator, error)
    func NewMmapAllocatorFactory(cfg buffer.MmapAllocatorConfig, fallbackToHeap bool) AllocatorFactory
type Config struct{ ... }
    func DefaultConfig() Config
type Datagram struct{ ... }
type DatagramPayloadDecoder interface{ ... }
type DatagramToMessageDecoder struct{ ... }
    func NewDatagramToMessageDecoder(decoder DatagramPayloadDecoder) *DatagramToMessageDecoder
    func NewDatagramToMessageDecoderFunc(accept func(Datagram) bool, ...) *DatagramToMessageDecoder
type MessageToDatagramEncoder struct{ ... }
    func NewMessageToDatagramEncoder(encoder AddressedMessageEncoder) *MessageToDatagramEncoder
    func NewMessageToDatagramEncoderFunc(accept func(any) bool, ...) *MessageToDatagramEncoder
type Server struct{ ... }
type Transport struct{ ... }
    func NewTransport(cfg Config) *Transport
```
