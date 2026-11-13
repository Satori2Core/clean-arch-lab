# AI 协作工作流

> 目标：让 AI 产出可控，而不是「先写再修」。

## 核心原则

1. **契约先行**：先定义接口契约与 E2E 行为，再让 AI 实现
2. **小步生成**：一个用例一个 diff，禁止大爆炸式提交
3. **生成期规则**：AGENTS.md 写明分层/依赖/禁止项，AI 动手前先读
4. **自动门禁**：`go test ./...`、`go vet ./...`、E2E 全部通过才算完成
5. **评审闭环**：AI 写完必须走 Code Review，报告归档，逐条确认修/忽略

## 工作流

```
1. 写契约     docs/<domain>/api-contract.md + e2e-testcases.md
2. 写 ADR     （有架构决策时）
3. 小步实现   domain → app → infra → api → e2e
4. 跑门禁     go test ./... && go vet ./...
5. AI CR     subagent 并行审查 + 主 agent 复核
6. 归档       cr/ 报告，Owner 确认「修 / 忽略 / 稍后」
7. 提交       commit 信息说明改动原因
```

## 上下文控制（防污染）

- 根 `AGENTS.md` 只做导航，不写业务细节
- 每个域一份短索引 + 长文档按任务选读
- 评审只看**本次 diff**，不看全仓历史

## AI Code Review 约定

- **增量审查**：`git diff <base>...HEAD`，排除生成物/第三方
- **subagent 并行审查**：按规则域拆分（错误处理、SQL、安全、日志）
- **主 agent 复核**：剔除误报，避免“AI 审 AI”空转
- **忽略机制**：确认忽略的项后续不再重复报
