## CR 审查报告

**审查范围**：clean-arch-lab tasks Delete 功能增量（domain/repository.go、domain/task.go、infra/memory/repository.go、app/service.go、api/handler.go、测试、e2e）
**审查模式**：全文件审查；subagent 并行审查（2 个 subagent）
**涉及规则**：clean-arch-lab/AGENTS.md（命名与注释、函数结构、错误处理、日志、输入校验、测试、禁止）

### 问题列表

#### [必须]

无（本轮发现的必须项已在修复记录中闭环）。

#### [建议]

1. [建议] 测试 [service_test.go:104](../internal/tasks/app/service_test.go#L104) **service 层测试命名未与用例 ID 对齐**。建议重命名或注释标注。
规则：AGENTS.md > 测试 > 测试命名与用例 ID 对齐
忽略：否

```go
func TestCreateTask(t *testing.T) {
```

2. [建议] 函数结构 [repository.go:54](../internal/tasks/infra/memory/repository.go#L54) **List 按字符串字典序排序，`task-10` 会排在 `task-2` 前**。建议按数字后缀或 CreatedAt 排序。
规则：AGENTS.md > 函数结构 > 输出确定性
忽略：否

```go
sort.Slice(tasks, func(i, j int) bool {
```

### 总结

| 级别 | 数量 |
|------|------|
| 必须修复 | 0 项 |
| 建议优化 | 2 项 |
| 已忽略 | 0 项 |

**CR 结论**：通过

---

### 本轮修复记录

- ✅ Create/MarkDone/Delete 对领域校验错误统一 `apperr.Wrapf` 补上下文
- ✅ `fakeRepo` 增加 `findErr`，新增 `TestMarkDoneFindError`
- ✅ `ValidateID` 改为按 `utf8.RuneCountInString` 计算长度，补多字节边界测试
- ✅ 仓储删除成功断言 NOT_FOUND 业务码
- ✅ E2E C07 校验 id/title/created_at；C09 删除后二次删除验证已删除
- ✅ handler 新增非法 ID 400 测试（markDone/delete）
- ✅ handler C07-C10 测试命名对齐用例 ID
- ✅ 超大请求体测试提取 `oversizedRequestBodyBytes` 常量
