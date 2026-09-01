# 概览

[English](overview.md) | [文档索引](README.zh-CN.md)

## 目标

Gnalloy 原生 UDP datagram server 与 connected client 传输。

该模块负责一个具体 I/O 边界，为具体传输创建或适配 Gnalloy Channel。协议解析、业务 handler、TLS 策略和观测能力由其他模块负责。

## 仓库身份

- 模块路径：`gnalloy.org/transport-udp`
- GitHub 仓库：`github.com/gnalloy/transport-udp`
- 默认分支：`dev`
- 许可证：Apache-2.0

## 包结构
- `gnalloy.org/transport-udp`（`udp`）

## 直接 Gnalloy 依赖

- `gnalloy.org/gnalloy`

## 当前仓库集合中的直接下游

- `gnalloy.org/examples`
- `gnalloy.org/handler-ipfilter`
- `gnalloy.org/handler-pcap`
- `gnalloy.org/protocol`

## 架构位置

Gnalloy 保持核心小而轻依赖。本仓库围绕单一职责形成可替换模块，通过显式 Go package 连接，而不是依靠运行时发现。

## 兼容性

公共导入路径是 `gnalloy.org/transport-udp`。首个稳定 tag 发布前，请按依赖策略使用 `@dev` 或明确的 pseudo-version。
