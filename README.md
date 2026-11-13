# clean-arch-lab

一个用于练习「模块化单体 + 领域分包 + 端口适配器」的 Go 练手项目。
业务域采用通用虚构场景：**个人任务管理（tasks）**，不含任何特定业务领域。

## 目录结构

```
cmd/server/              组合根：唯一装配点
docs/
  architecture.md        目标架构说明
  decisions/             ADR（架构决策记录）
  ai-workflow.md         AI 协作工作流
  e2e-testcases.md       E2E 行为契约
internal/
  shared/apperr/         跨域通用错误类型
  tasks/
    domain/              领域模型 + 端口（接口）
    app/                 用例编排
    api/                 HTTP 适配器
    infra/memory/        内存仓储适配器
    e2e/                 E2E 测试（harness + 用例）
```

## 设计原则

- **依赖方向单向**：`api → app → domain ← infra`
- **domain 不依赖任何框架/数据库**，只定义端口（接口）
- **组合根唯一装配**，消灭全局单例
- **契约先行**：新接口先写 `docs/<domain>/api-contract.md` 与 `docs/e2e-testcases.md`，再实现
- **小步演进**：一个用例一个 diff，配合 ADR 记录决策过程

## 运行

```bash
go run ./cmd/server
```

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"买牛奶"}'

curl http://localhost:8080/api/v1/tasks
```

## 测试

```bash
go test ./...
```
