# API 参考

[English](api.md) | [文档索引](README.zh-CN.md)

本清单由本仓库 package 的 `go doc -short` 生成，用于快速查看公共面。精确语义以源码和测试为准。

## 包

### `gnalloy.org/transport-udp`

包名：`udp`

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
