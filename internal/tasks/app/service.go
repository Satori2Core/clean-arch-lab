// Package app 是 tasks 域的用例层（Application Service）。
// 它只依赖 domain 端口，负责编排业务规则，不包含 HTTP/DB 细节。
package app

import (
	"context"
	"time"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
)

// Service 是任务应用服务。
type Service struct {
	repo domain.Repository
	ids  domain.IDGenerator
}

// ServiceDeps 收敛依赖注入参数。
type ServiceDeps struct {
	// Repo 是任务存储端口。
	Repo domain.Repository
	// IDs 是任务 ID 生成端口。
	IDs domain.IDGenerator
}

// NewService 构造应用服务；依赖缺失时直接 panic（fail-fast，避免后续 nil 解引用）。
func NewService(deps ServiceDeps) *Service {
	if deps.Repo == nil {
		panic("app.NewService: Repo is nil")
	}
	if deps.IDs == nil {
		panic("app.NewService: IDs is nil")
	}
	return &Service{repo: deps.Repo, ids: deps.IDs}
}

// CreateTask 是创建任务的用例入参。
type CreateTask struct {
	// Title 是任务标题，必填，TrimSpace 后不超过 MaxTaskTitleLength。
	Title string
}

// Create 执行「创建任务」用例：先校验标题 → 生成 ID → 领域构造 → 持久化。
// 校验必须先于任何可失败副作用（ID 生成），确保非法输入稳定返回 INVALID_ARGUMENT。
func (s *Service) Create(ctx context.Context, cmd CreateTask) (domain.Task, error) {
	title, err := domain.ValidateTitle(cmd.Title)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "create task: validate title")
	}
	id, err := s.ids.NewID(ctx)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "generate task id failed")
	}
	task, err := domain.NewTask(id, title, time.Now())
	if err != nil {
		return domain.Task{}, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return domain.Task{}, apperr.Wrapf(err, "save task failed")
	}
	return task, nil
}

// List 返回全部任务；空数据时返回非 nil 空切片，保证 JSON 序列化为 [] 而非 null。
func (s *Service) List(ctx context.Context) ([]domain.Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperr.Wrapf(err, "list tasks failed")
	}
	if tasks == nil {
		tasks = make([]domain.Task, 0)
	}
	return tasks, nil
}

// GetByID 按 ID 查询任务详情：校验并规范化 ID → 按 ID 查询，不存在时返回 NOT_FOUND。
func (s *Service) GetByID(ctx context.Context, id string) (domain.Task, error) {
	id, err := domain.ValidateID(id)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "get task: validate id")
	}
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "find task failed, id=%s", id)
	}
	return task, nil
}

// MarkDone 标记任务为已完成：校验并规范化 ID → 按 ID 查询 → 领域置为完成 → 持久化。
func (s *Service) MarkDone(ctx context.Context, id string) (domain.Task, error) {
	id, err := domain.ValidateID(id)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "mark done: validate id")
	}
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, apperr.Wrapf(err, "find task failed, id=%s", id)
	}
	task.MarkDone()
	if err := s.repo.Save(ctx, task); err != nil {
		return domain.Task{}, apperr.Wrapf(err, "save task failed")
	}
	return task, nil
}

// Delete 删除任务：校验并规范化 ID 后按 ID 删除，不存在时返回 NOT_FOUND。
func (s *Service) Delete(ctx context.Context, id string) error {
	id, err := domain.ValidateID(id)
	if err != nil {
		return apperr.Wrapf(err, "delete task: validate id")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.Wrapf(err, "delete task failed, id=%s", id)
	}
	return nil
}
