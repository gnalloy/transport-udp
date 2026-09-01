# ADR-001: Keep transport-udp as an Independent Gnalloy Module

[简体中文](0001-module-boundary.zh-CN.md) | [Docs Index](../README.md)

## Status

Accepted

## Context

Gnalloy is published as focused repositories under `github.com/gnalloy` with public Go import paths under `gnalloy.org`. The core must stay small enough for hot network paths and for applications that only need a subset of protocols or transports.

## Decision

`transport-udp` is published as `gnalloy.org/transport-udp` and owns this responsibility: Native UDP datagram server and connected client transport for Gnalloy.

The core module remains `gnalloy.org/gnalloy`. Protocol codecs, concrete transports, handlers, resolvers, observability adapters, examples, and benchmark tooling depend on the core contracts instead of living inside the core repository.

## Consequences

- Consumers can depend on only the protocol, transport, handler, resolver, or tooling module they need.
- Hot-path core code avoids optional protocol, compression, QUIC, OpenTelemetry, example, and benchmark dependencies.
- Cross-module behavior is assembled by applications, recipes, examples, or benchmark harnesses.
- Cross-module changes require standalone module verification plus integrated workspace verification before push.
