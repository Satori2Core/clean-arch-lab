# tasks API 契约

> 契约先行：新增/修改接口必须先更新本文件与 `docs/e2e-testcases.md`。

## 接口总览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/tasks` | 创建任务 |
| GET | `/api/v1/tasks` | 查询任务列表 |
| POST | `/api/v1/tasks/{id}/done` | 标记任务完成 |

## GET /healthz

### 成功响应（200）

```text
ok
```

## POST /api/v1/tasks

### 请求

```json
{
  "title": "买牛奶"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| title | string | 是 | 非空，≤ 200 字符，服务端会 TrimSpace |

### 成功响应（201）

```json
{
  "id": "task-1",
  "title": "买牛奶",
  "done": false,
  "created_at": "2026-01-15T00:00:00+08:00"
}
```

### 错误

| HTTP | code | 场景 |
|------|------|------|
| 400 | `INVALID_ARGUMENT` | 标题为空 / 超长 / JSON 格式错误 |
| 500 | `INTERNAL` | 内部错误（不返回内部细节） |

## GET /api/v1/tasks

### 成功响应（200）

```json
[
  {
    "id": "task-1",
    "title": "买牛奶",
    "done": false,
    "created_at": "2026-01-15T00:00:00+08:00"
  }
]
```

空列表返回 `[]`。

## POST /api/v1/tasks/{id}/done

### 请求

无请求体。`id` 为任务 ID。

### 成功响应（200）

```json
{
  "id": "task-1",
  "title": "买牛奶",
  "done": true,
  "created_at": "2026-01-15T00:00:00+08:00"
}
```

### 错误

| HTTP | code | 场景 |
|------|------|------|
| 404 | `NOT_FOUND` | 任务不存在 |
| 500 | `INTERNAL` | 内部错误 |
