# 用法

[English](usage.md) | [文档索引](README.zh-CN.md)

## 要求

- Go 1.25 或更新版本，并与 module 的 `go` 指令一致。
- 由 Gnalloy 应用、recipe、example 或 benchmark harness 负责生命周期与部署配置。
- 独立模块复验应设置 `GOWORK=off`，确保通过已发布依赖图测试。

## 安装
```bash
go get gnalloy.org/transport-udp@dev
```

## 导入
```go
import "gnalloy.org/transport-udp"
```

## 集成模式
- address、listener、dialer、buffer allocator、event loop 与 channel initializer 都属于 transport 边界配置。
- 平台相关 transport 必须返回显式 unsupported 错误，不能静默降级。
- raw socket、L2 capture 等特权传输需要 Go 模块之外的操作系统能力。
- protocol、TLS、proxy 与 observability handler 应通过 Channel pipeline 安装。
- `Write` 将消息所有权转移到出站队列；调用 `Flush` 后才提交已排队的 datagram。Linux 使用 `sendmmsg` 批量提交一次 flush，其他平台保持有序逐包回退。

## API 选择

通过 API 清单选择当前协议路径需要的具体构造函数或 option 类型：

```bash
go doc gnalloy.org/transport-udp
```

当前常用入口：
- `var ErrInvalidAddress = errors.New("gnalloy/transport/udp: invalid address") ...`
- `type Config struct{ ... }`

## 跨模块装配

多个 Gnalloy 仓库一起开发时，在自己选择的 workspace 中创建本地 `go.work` 文件。不要把应用本地 `replace` 指令提交到发布用 library module，除非它是明确的临时变更且不会进入提交。

## 错误处理

网络输入、对端行为、平台能力和超时失败都必须作为普通错误处理。不要用 panic 恢复协议正确性。返回或传播模块错误，并在所有权要求时关闭受影响的 Channel。
