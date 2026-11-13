package domain

import "context"

// Repository 是任务存储端口（Port）。
// 具体实现（内存/MySQL/Redis）在 infra 层提供，领域层只依赖此接口。
type Repository interface {
	// Save 保存任务（按 ID 覆盖）。
	Save(ctx context.Context, task Task) error
	// FindByID 按 ID 查询任务，不存在时返回 NOT_FOUND。
	FindByID(ctx context.Context, id string) (Task, error)
	// List 返回任务列表。
	List(ctx context.Context) ([]Task, error)
}

// IDGenerator 是 ID 生成端口（Port），隔离具体 ID 生成方案。
type IDGenerator interface {
	// NewID 生成新任务 ID。
	NewID(ctx context.Context) (string, error)
}
