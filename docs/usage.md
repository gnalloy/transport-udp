# Usage

[简体中文](usage.zh-CN.md) | [Docs Index](README.md)

## Requirements

- Go 1.25 or newer, matching the module `go` directive.
- A Gnalloy application, recipe, example, or benchmark harness that owns lifecycle and deployment configuration.
- Standalone module verification should set `GOWORK=off` so the module is tested through its published dependency graph.

## Install
```bash
go get gnalloy.org/transport-udp@dev
```

## Import
```go
import "gnalloy.org/transport-udp"
```

## Integration Pattern
- Address, listener, dialer, buffer allocator, event loop, and channel initializer choices are part of the transport boundary.
- Platform-specific transports must return explicit unsupported errors rather than silently falling back.
- Privileged transports such as raw sockets and L2 capture need operating-system capabilities outside the Go module.
- Protocol, TLS, proxy, and observability handlers should be installed through the Channel pipeline.
- `Write` transfers message ownership into the outbound queue; call `Flush` to submit queued datagrams. Linux batches one flush with `sendmmsg`, while other platforms preserve ordered fallback writes.

## API Selection

Use the API inventory to choose the exact constructor or option type for your protocol path:

```bash
go doc gnalloy.org/transport-udp
```

Common current entry points:
- `var ErrInvalidAddress = errors.New("gnalloy/transport/udp: invalid address") ...`
- `type Config struct{ ... }`

## Cross-Module Assembly

When multiple Gnalloy repositories are developed together, create a local `go.work` file in your chosen workspace. Keep application-local `replace` directives out of published library modules unless the change is intentionally temporary and never committed.

## Error Handling

Network input, peer behavior, platform capability, and timeout failures must be handled as normal errors. Do not recover protocol correctness by panicking. Return or propagate the module error and close the affected Channel when ownership requires it.
