---
name: distributed-game-server
description: 架构设计、实现和审查基于 Go 与 Ergo 的大型分布式在线游戏服务端。适用于登录和网关、玩家会话、世界与区服拓扑、场景、战斗、匹配、社交系统、经济、持久化、分片、跨服流程、容量、韧性、安全、可观测性、滚动升级和游戏服务端测试。
---

# 分布式游戏服务端

先根据不变量和状态归属完成设计，再通过 `$ergo-framework` 映射为 Ergo 实现。不要把每个领域对象都做成 Actor，也不要把每个扩展性问题都交给 Pool。

## 必须完成的设计检查

每个非简单功能都必须确定：

1. 限界上下文与明确的排除范围。
2. 权威状态所有者与 Actor 标识。
3. 激活、重连、迁移、空闲、停机和恢复生命周期。
4. 命令、查询、事件、顺序范围、deadline 和背压。
5. 持久化边界、持久化生效点、幂等键与重复处理策略。
6. 路由键、分片映射、热点行为与再平衡策略。
7. 故障模式：节点丢失、网络分区、依赖超时、过期会话、局部提交、重试风暴和过载。
8. 容量预算：并发数、每秒消息、handler 耗时、状态大小、带宽、存储写入和恢复速率。
9. 可观测性：关联 ID、领域指标、追踪、审计记录和安全的运维动作。
10. 版本兼容、灰度、排水、回滚与数据迁移。
11. 视情况实施 unit、stage、性质测试、fuzz、回放、压测、长稳和故障注入测试。

## 核心不变量

- 除非有明确的共识或 CRDT 设计，每个可变聚合只能有一个权威写入者。
- 连接不是身份；使用单调递增 epoch 或等价 token 对旧会话做 fencing。
- 传输确认不等于经济提交。货币和高价值物品变更必须具备业务交易 ID、去重、可审计性和对账机制。
- 只在最小必要键内保持顺序，例如玩家、房间、对局或公会。
- 恢复能力是稳态容量的一部分；节点故障后避免同时激活所有玩家或重放全部状态。
- 必须明确过载策略：丢弃、有限队列、降级、重定向或拒绝；不得让无界邮箱成为默认策略。
- 滚动发布期间，跨版本消息和持久化状态必须存在兼容读取者。

## Reference 导航

| 需求 | 阅读 |
|---|---|
| 领域边界与服务归属 | `references/domain-boundaries.md` |
| Actor 标识、生命周期与状态放置 | `references/actor-modeling.md` |
| 网关、认证、重连与会话 fencing | `references/session-and-gateway.md` |
| 区服、分片、路由、热点与跨服拓扑 | `references/topology-and-sharding.md` |
| 持久化、缓存、幂等、Saga 与对账 | `references/persistence-and-consistency.md` |
| 场景、战斗、匹配、社交与经济模式 | `references/gameplay-services.md` |
| 容量、延迟、过载与性能测试 | `references/capacity-and-performance.md` |
| 故障恢复、分区与灾备规划 | `references/resilience.md` |
| 协议与持久化 schema 演进 | `references/protocol-evolution.md` |
| 部署、排水、灰度与回滚 | `references/deployment.md` |
| 信任边界、重放与运维安全 | `references/security.md` |
| 测试策略与证据等级 | `references/testing.md` |

请求交付系统或功能设计文档时，使用 `assets/architecture-template.md`。
