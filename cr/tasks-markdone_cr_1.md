## CR 审查报告

**审查范围**：clean-arch-lab tasks MarkDone 功能增量（domain/task.go、app/service.go、api/handler.go、service_test.go、handler_test.go、e2e）
**审查模式**：全文件审查；subagent 并行审查（2 个 subagent）
**涉及规则**：clean-arch-lab/AGENTS.md（命名与注释、函数结构、错误处理、日志、输入校验、测试、禁止）

### 问题列表

#### [必须]

无（本轮发现的必须项已在修复记录中闭环；C06 健康检查单测已存在于 `cmd/server/main_test.go`，subagent 审查范围未覆盖，属误报）。

#### [建议]

1. [建议] 测试 [service_test.go:77](../internal/tasks/app/service_test.go#L77) **service 层单测命名未带用例 ID**。建议命名或注释中标注对应用例 ID。
规则：AGENTS.md > 测试 > 测试命名与用例 ID 对齐
忽略：否

```go
func TestCreateTask(t *testing.T) {
```

2. [建议] 测试 [env.go:10](../internal/tasks/e2e/harness/env.go#L10) **非测试包 harness 导入 `testing`**。建议测试辅助放入 `_test.go` 或独立 testutil 包。
规则：AGENTS.md > 测试 > 通用质量问题
忽略：否

```go
"testing"
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

- ✅ `domain.ValidateID` 改为返回规范化 ID，MarkDone 使用规范化 ID 查询
- ✅ 新增 `TestMarkDoneRejectsInvalidID`（空 ID / 超长 ID）
- ✅ `TestMarkDoneSaveError` 断言业务码 INTERNAL
- ✅ `TestMarkTaskDoneHTTPNotFound` 断言 code=NOT_FOUND
- ✅ 新增 `TestWithRequestIDTooLongHeaderGeneratesNew`
- ✅ `maxRequestIDLength` 复用 `domain.MaxTaskIDLength`
- ✅ `TaskNotFoundMessage` 常量接入仓储/fake 使用
- ✅ 新增 `TestValidateIDBoundaryMaxLength`
