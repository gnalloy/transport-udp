# transport-udp

[English](README.md) | [文档](docs/README.zh-CN.md)

Gnalloy 原生 UDP datagram server 与 connected client 传输。

该模块负责一个具体 I/O 边界，为具体传输创建或适配 Gnalloy Channel。协议解析、业务 handler、TLS 策略和观测能力由其他模块负责。

## 状态

- 导入路径：`gnalloy.org/transport-udp`
- 仓库：`github.com/gnalloy/transport-udp`
- 默认分支：`dev`
- 预览安装：`go get gnalloy.org/transport-udp@dev`
- 许可证：Apache-2.0

## 安装
```bash
go get gnalloy.org/transport-udp@dev
go doc gnalloy.org/transport-udp
GOWORK=off GOTOOLCHAIN=local go test ./... -count=1
```

## 文档
- [概览](docs/overview.zh-CN.md) ([English](docs/overview.md))
- [用法](docs/usage.zh-CN.md) ([English](docs/usage.md))
- [案例](docs/examples.zh-CN.md) ([English](docs/examples.md))
- [配置说明](docs/configuration.zh-CN.md) ([English](docs/configuration.md))
- [测试与性能](docs/testing.zh-CN.md) ([English](docs/testing.md))
- [API 参考](docs/api.zh-CN.md) ([English](docs/api.md))
- [注意事项与备注](docs/notes.zh-CN.md) ([English](docs/notes.md))
- [ADR-001 模块边界](docs/decisions/0001-module-boundary.zh-CN.md) ([English](docs/decisions/0001-module-boundary.md))

## 模块边界

本仓库负责：Gnalloy 原生 UDP datagram server 与 connected client 传输。

它不吸收相邻模块职责。核心基础能力保留在 `gnalloy.org/gnalloy`；协议 codec、transport、handler、resolver、examples 与 benchmarks 分别由独立仓库负责。

## 包结构
- `gnalloy.org/transport-udp`（`udp`）

## Gnalloy 依赖

- `gnalloy.org/gnalloy`

## 常见集成方式
- address、listener、dialer、buffer allocator、event loop 与 channel initializer 都属于 transport 边界配置。
- 平台相关 transport 必须返回显式 unsupported 错误，不能静默降级。
- raw socket、L2 capture 等特权传输需要 Go 模块之外的操作系统能力。
- protocol、TLS、proxy 与 observability handler 应通过 Channel pipeline 安装。

## 当前公共入口

生成的 API 参考列出了完整公共面。当前常用构造函数或 option 类型包括：
- `var ErrInvalidAddress = errors.New("gnalloy/transport/udp: invalid address") ...`
- `type Config struct{ ... }`

## 验证

```bash
GOWORK=off GOTOOLCHAIN=local go test ./... -count=1
GOWORK=off GOTOOLCHAIN=local go vet ./...
GOWORK=off GOTOOLCHAIN=local go test ./... -run '^$' -bench . -benchmem -count=1
```

压测时，将该模块和相应 transport、codec、handler 栈装配后，使用 `gnalloy.org/benchmarks` 或 `gnalloy.org/examples` 中的场景运行。报告必须保留主机、操作系统、payload、并发度、warmup 和 repetition。

## 注意事项
- 本仓库保持窄边界。跨模块行为应在应用、recipes、examples 或 benchmark harness 中装配。
- 公共 API 必须保持 Go 原生和显式；热路径避免运行时扫描、隐藏全局注册表和重反射。
- 网络输入一律视为不可信。配置解析上限，返回类型化错误，不使用 panic 处理输入错误。
- 性能结论必须绑定具体主机、操作系统、协议、payload、并发度、warmup 和 repetition。
- transport 模块本身不解析应用协议；需要在 pipeline 中安装合适的 codec 与 handler。
