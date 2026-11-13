# tasks 域 E2E 测试用例

> 用例 ID 与 `internal/tasks/e2e/` 实现一一对应。

| 用例 ID | 场景 | 期望 |
|---------|------|------|
| E2E-TASK-C01 | 创建任务成功 | 201，返回 id/title |
| E2E-TASK-C02 | 列出任务 | 200，返回数组 |
| E2E-TASK-C03 | 创建空标题 | 400，INVALID_ARGUMENT |
| E2E-TASK-C04 | 创建超长标题 | 400，INVALID_ARGUMENT |
| E2E-TASK-C05 | 空列表 | 200，返回空数组 `[]` |
| E2E-TASK-C06 | 健康检查 | 200，返回 ok |
| E2E-TASK-C07 | 标记任务完成 | 200，done=true |
| E2E-TASK-C08 | 标记不存在的任务 | 404，NOT_FOUND |
| E2E-TASK-C09 | 删除任务成功 | 204，无响应体 |
| E2E-TASK-C10 | 删除不存在的任务 | 404，NOT_FOUND |
| E2E-TASK-C11 | 查询任务详情成功 | 200，返回 id/title/done/created_at |
| E2E-TASK-C12 | 查询不存在的任务 | 404，NOT_FOUND |
