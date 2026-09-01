# transport-udp

[简体中文](README.zh-CN.md) | [Documentation](docs/README.md)

Native UDP datagram server and connected client transport for Gnalloy.

This module owns an I/O boundary. It creates or adapts Gnalloy Channels for a concrete transport while protocol parsing, business handlers, TLS policy, and observability remain in other modules.

## Status

- Import path: `gnalloy.org/transport-udp`
- Repository: `github.com/gnalloy/transport-udp`
- Default branch: `dev`
- Preview install: `go get gnalloy.org/transport-udp@dev`
- License: Apache-2.0

## Install
```bash
go get gnalloy.org/transport-udp@dev
go doc gnalloy.org/transport-udp
GOWORK=off GOTOOLCHAIN=local go test ./... -count=1
```

## Documentation
- [Overview](docs/overview.md) ([中文](docs/overview.zh-CN.md))
- [Usage](docs/usage.md) ([中文](docs/usage.zh-CN.md))
- [Examples](docs/examples.md) ([中文](docs/examples.zh-CN.md))
- [Configuration](docs/configuration.md) ([中文](docs/configuration.zh-CN.md))
- [Testing and Performance](docs/testing.md) ([中文](docs/testing.zh-CN.md))
- [API Reference](docs/api.md) ([中文](docs/api.zh-CN.md))
- [Notes and Caveats](docs/notes.md) ([中文](docs/notes.zh-CN.md))
- [ADR-001 Module Boundary](docs/decisions/0001-module-boundary.md) ([中文](docs/decisions/0001-module-boundary.zh-CN.md))

## Module Boundary

This repository owns: Native UDP datagram server and connected client transport for Gnalloy.

It does not absorb neighboring module responsibilities. Core primitives stay in `gnalloy.org/gnalloy`; protocol codecs, transports, handlers, resolvers, examples, and benchmarks stay in their own repositories.

## Packages
- `gnalloy.org/transport-udp` (`udp`)

## Gnalloy Dependencies
- `gnalloy.org/gnalloy`

## Common Integration Pattern
- Address, listener, dialer, buffer allocator, event loop, and channel initializer choices are part of the transport boundary.
- Platform-specific transports must return explicit unsupported errors rather than silently falling back.
- Privileged transports such as raw sockets and L2 capture need operating-system capabilities outside the Go module.
- Protocol, TLS, proxy, and observability handlers should be installed through the Channel pipeline.

## Current Public Entry Points

The generated API reference lists the full public surface. Common constructors or option types currently include:
- `var ErrInvalidAddress = errors.New("gnalloy/transport/udp: invalid address") ...`
- `type Config struct{ ... }`

## Verification

```bash
GOWORK=off GOTOOLCHAIN=local go test ./... -count=1
GOWORK=off GOTOOLCHAIN=local go vet ./...
GOWORK=off GOTOOLCHAIN=local go test ./... -run '^$' -bench . -benchmem -count=1
```

For pressure tests, assemble this module with the relevant transport, codec, and handler stack and run the scenario from `gnalloy.org/benchmarks` or `gnalloy.org/examples`. Keep host, operating system, payload, concurrency, warmup, and repetitions in the report.

## Caveats
- This repository is intentionally narrow. Cross-module behavior should be assembled in applications, recipes, examples, or benchmark harnesses.
- Public APIs should remain Go-native and explicit; avoid runtime scanning, hidden global registries, and reflection-heavy behavior in hot paths.
- Treat network input as untrusted. Configure parser limits and return typed errors instead of panics.
- Keep benchmark claims tied to a concrete host, operating system, protocol, payload, concurrency, warmup, and repetition count.
- Transport modules do not parse application protocols by themselves; install the right codec and handler modules in the pipeline.
