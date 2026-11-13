## CR 审查报告

**审查范围**：clean-arch-lab tasks GetByID 功能增量（app/service.go、api/handler.go、service_test.go、handler_test.go、e2e、domain 常量）
**审查模式**：全文件审查；subagent 并行审查（2 个 subagent）
**涉及规则**：clean-arch-lab/AGENTS.md（命名与注释、函数结构、错误处理、日志、输入校验、测试、禁止）

### 问题列表

#### [必须]

无（本轮发现的必须项已在修复记录中闭环）。

#### [建议]

1. [建议] 测试 [service_test.go:190](../internal/tasks/app/service_test.go#L190) **service 层测试命名未与用例 ID 对齐**。建议重命名或注释标注。
规则：AGENTS.md > 测试 > 测试命名与用例 ID 对齐
忽略：否

```go
func TestCreateTaskIDGenError(t *testing.T) {
```

2. [建议] 日志 [handler.go:92](../internal/tasks/api/handler.go#L92) **`newRequestID` 随机数失败时未记录日志**。建议在失败分支记录日志后再返回占位符。
规则：AGENTS.md > 日志 > 结构化 JSON；错误处理 > 禁止吞错
忽略：否

```go
if _, err := rand.Read(b); err != nil {
```

3. [建议] 函数结构 [repository.go:54](../internal/tasks/infra/memory/repository.go#L54) **List 字符串字典序排序，`task-10` 会排在 `task-2` 前**。建议按数字后缀或 CreatedAt 排序。
规则：AGENTS.md > 函数结构 > 输出确定性
忽略：否

```go
sort.Slice(tasks, func(i, j int) bool {
```

### 总结

| 级别 | 数量 |
|------|------|
| 必须修复 | 0 项 |
| 建议优化 | 3 项 |
| 已忽略 | 0 项 |

**CR 结论**：通过

---

### 本轮修复记录

- ✅ `domain.NewTask` 错误统一 `apperr.Wrapf` 补上下文
- ✅ 新增 `TestCreateTaskNewTaskError`（空 ID 触发领域构造失败）
- ✅ `failingRepo` 支持注入 `findErr`，新增 GetByID / MarkDone / Delete 三个 5xx 不泄露内部细节测试
- ✅ `TaskIDRequiredMessage` 常量复用（NewTask / ValidateID）
- ✅ E2E `tasksBase` / `healthzPath` 补注释
