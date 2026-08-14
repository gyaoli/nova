---
name: ergo-operations
description: 诊断 Ergo Framework 集群与分布式游戏服务的线上或疑似生产问题。适用于进程泄漏、邮箱压力、延迟、重启循环、内存或 CPU 增长、网络故障、事件扇出、profile、sampler 和集群健康检查；默认只读，未获明确授权不得使用变更型工具。
---

# Ergo 运维诊断

以有界、假设驱动的证据调查运行中的 Ergo 系统。Ergo MCP 端点可能提供 reference 中描述的工具，但必须在当前会话中核验工具可用性和 schema。

## 安全边界

- 从只读开始，明确目标节点、时间窗口、部署版本和症状。
- 执行 `network_connect`、`network_disconnect`、`send_message`、`call_process`、`send_exit`、`process_kill`、`log_level_set` 或同类变更前，必须获得明确授权。
- 限制并排序进程、事件、sampler 和 profile 结果。
- 每个 sampler 都设置有限 `duration_sec` 和足够的 linger 窗口。
- 过滤远程 profile；不得通过 proxy 传输无界 goroutine 或 heap dump。
- 检查网络连接两端并关联时间戳。
- 脱敏集群 cookie、token、凭据、玩家标识、聊天内容和个人数据。

## 调查循环

1. 提出一个或多个可证伪假设。
2. 低成本采集基线：node、application、process、runtime 和 network 摘要。
3. 按服务角色、注册名、PID、邮箱、错误计数或时间窗口逐步收窄。
4. 仅在计数器无法判断假设时使用 sampler 或 profile。
5. 将发现与发布、重启、流量和下游依赖关联。
6. 报告已确认观察、可能解释、已排除的替代原因、风险和下一步安全动作。

MCP 不可用时，使用仓库源码、日志、指标、配置、OS 进程信息和既有 profile；明确标注结果并非实时证据或证据不完整。

## Reference 导航

| 需求 | 阅读 |
|---|---|
| 当前 MCP 工具目录与参数 | `references/tools.md` |
| Actor 状态、队列、延迟与活性 | `references/process-model.md` |
| node、process、network 与 event 计数器含义 | `references/counters.md` |
| Important 错误、重启强度、Pool 与网络内部机制 | `references/framework-internals.md` |
| 标准诊断流程 | `references/playbooks.md` |
| 主动、被动与追踪 sampler | `references/samplers.md` |
| pprof、latency、verbose、typestats 与 norecover tag | `references/build-tags.md` |

## 游戏服务检查

对于网关风暴、玩家 Actor 泄漏、场景热点、战斗卡顿、匹配延迟、跨服路由和持久化滞后，还应阅读 `$distributed-game-server` 中相应领域的 reference。未判定瓶颈是 CPU、邮箱服务时间、键亲和性、下游延迟、锁竞争、网络饱和还是恢复抖动前，不得建议扩容。
