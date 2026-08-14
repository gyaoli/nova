---
name: ergo-framework
description: 设计、实现、审查和测试基于 Ergo Framework v3.3 或 v330 的 Go 系统。适用于 Actor、应用、Supervisor、Send 或 Call、meta process、EDF、集群、日志、追踪、cron、Pool、路由、网络和 Ergo 测试工具；不要将其作为游戏领域拆分或线上集群事故响应的唯一依据。
---

# Ergo 框架开发

本 skill 用于 Ergo 框架机制。方法签名和版本相关行为以项目的 `go.mod` 以及模块缓存或检出目录中的精确 Ergo 源码为准。

## 工作流程

1. 阅读仓库 `AGENTS.md`。
2. 处理版本敏感内容前，检查 `go.mod` 并定位实际 Ergo 源码。
3. 仅阅读当前任务相关的 reference。
4. 非简单实现前，说明状态所有者、生命周期、故障边界、消息语义、超时行为和验证计划。
5. 在包边界提供 helper，避免调用者构造内部消息或硬编码已注册进程名。
6. 运行定向测试，并区分编译、单测、stage 与真实运行时证据。

## 不可违反的规则

- 不得在 Actor 回调中无期限等待。使用有界 I/O，优先消息传递或无锁状态。
- 不得在 Actor 间暴露共享可变状态；传递值或不可变快照。
- 节点启动前注册 EDF 类型，通常在 `init()`；先注册嵌套类型，并导出跨节点字段。
- 不得将 Important Delivery 视为业务事务；必须设计应用层幂等和恢复。
- Meta process 不能发起 `Call`，也不能创建 link 或 monitor；它们用于承担阻塞 I/O，而非业务状态。
- 生产集群使用中心化 registrar；内嵌发现只用于开发。
- 未说明顺序、状态亲和性、背压和实测或预估负载时，不得选用 Pool 或 Router。

## Reference 导航

| 需求 | 阅读 |
|---|---|
| Actor 生命周期、回调、link、monitor、alias、event | `references/actors.md` |
| 监督树与重启策略 | `references/supervision.md` |
| Send、Call、优先级、Important Delivery、fallback | `references/messages.md` |
| Application、依赖、tag、map、生命周期 | `references/application.md` |
| Pool、Router、WebWorker 与背压 | `references/pool.md` |
| TCP、UDP、Web、子进程及其他 meta process | `references/meta.md` |
| Node、网络标志、安全、TLS 与生命周期 | `references/node.md` |
| 日志与自定义 logger adapter | `references/logging.md` |
| 追踪与业务 span | `references/tracing.md` |
| Cron 任务与 fallback | `references/cron.md` |
| 错误与终止原因 | `references/errors.md` |
| EDF 注册与 schema 演进 | `references/edf.md` |
| Registrar、发现、路由、tag 与 proxy | `references/cluster.md` |
| unit、stage、check 与 mock 测试 | `references/testing.md` |
| health、leader 与 metrics Actor | `references/actor-lib.md` |
| MCP、observer、pulse 与 radar Application | `references/applications-lib.md` |
| WebSocket 与 SSE 集成 | `references/meta-lib.md` |
| etcd、Saturn 与 logger 集成 | `references/integrations.md` |
| Erlang/OTP 互操作 | `references/erlang-protocol.md` |

## 命名

- 异步消息：`MessageXXX`。
- 同步调用：`XXXRequest` 和 `XXXResponse`。
- 使用 `any`，不要使用 `interface{}`。
- Actor 消息类型应靠近其所属限界上下文，并用 helper 函数作为公开接口。

涉及游戏服务端状态归属、分片、持久化、会话、经济和跨服流程时，同时使用 `$distributed-game-server`。
