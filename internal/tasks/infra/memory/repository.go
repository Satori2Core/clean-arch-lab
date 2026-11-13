// Package memory 是 tasks 域的「内存适配器」。
// 它实现 domain.Repository / domain.IDGenerator，用于演示业务不依赖真实数据库。
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
)

// Repository 是基于内存 map 的仓储实现。
type Repository struct {
	mu    sync.RWMutex
	tasks map[string]domain.Task
}

// NewRepository 创建内存仓储。
func NewRepository() *Repository {
	return &Repository{tasks: make(map[string]domain.Task)}
}

// Save 保存任务（按 ID 覆盖）；ID 合法性由 domain.NewTask 统一保证。
func (r *Repository) Save(_ context.Context, task domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

// FindByID 按 ID 查询任务。
func (r *Repository) FindByID(_ context.Context, id string) (domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	if !ok {
		return domain.Task{}, apperr.New(apperr.CodeNotFound, "task not found")
	}
	return task, nil
}

// List 返回全部任务，按 ID 升序保证稳定顺序。
func (r *Repository) List(_ context.Context) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]domain.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

// IDGenerator 是自增 ID 生成器实现（练手用，真实场景用 idgen/雪花等）。
type IDGenerator struct {
	prefix  string
	counter atomic.Int64
}

// NewIDGenerator 创建 ID 生成器。
func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

// NewID 生成自增 ID。
func (g *IDGenerator) NewID(_ context.Context) (string, error) {
	n := g.counter.Add(1)
	return fmt.Sprintf("%s%d", g.prefix, n), nil
}
