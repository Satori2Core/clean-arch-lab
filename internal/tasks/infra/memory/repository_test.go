package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/infra/memory"
)

// newTask 构造测试任务。
func newTask(id, title string) domain.Task {
	task, err := domain.NewTask(id, title, time.Now())
	if err != nil {
		panic(err)
	}
	return task
}

// TestRepositorySaveAndFindByID 验证保存后可按 ID 查询。
func TestRepositorySaveAndFindByID(t *testing.T) {
	repo := memory.NewRepository()
	task := newTask("task-1", "买牛奶")

	if err := repo.Save(context.Background(), task); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := repo.FindByID(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Title != "买牛奶" {
		t.Fatalf("got.Title = %q, want 买牛奶", got.Title)
	}
}

// TestRepositoryFindByIDNotFound 验证不存在时返回 NOT_FOUND。
func TestRepositoryFindByIDNotFound(t *testing.T) {
	repo := memory.NewRepository()
	_, err := repo.FindByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("FindByID() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

// TestRepositorySaveOverwritesByID 验证同 ID 保存会覆盖旧值。
func TestRepositorySaveOverwritesByID(t *testing.T) {
	repo := memory.NewRepository()
	if err := repo.Save(context.Background(), newTask("task-1", "old")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := repo.Save(context.Background(), newTask("task-1", "new")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := repo.FindByID(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Title != "new" {
		t.Fatalf("got.Title = %q, want new（同 ID 覆盖写）", got.Title)
	}
}

// TestRepositoryListStableOrder 验证 List 返回稳定升序结果。
func TestRepositoryListStableOrder(t *testing.T) {
	repo := memory.NewRepository()
	if err := repo.Save(context.Background(), newTask("task-2", "b")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := repo.Save(context.Background(), newTask("task-1", "a")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	tasks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("List() len = %d, want 2", len(tasks))
	}
	if tasks[0].ID != "task-1" || tasks[1].ID != "task-2" {
		t.Fatalf("List() 顺序 = [%s, %s], want [task-1, task-2]", tasks[0].ID, tasks[1].ID)
	}
}

// TestIDGeneratorNewID 验证 ID 顺序递增。
func TestIDGeneratorNewID(t *testing.T) {
	gen := memory.NewIDGenerator(domain.TaskIDPrefix)
	id1, err := gen.NewID(context.Background())
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	id2, err := gen.NewID(context.Background())
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if id1 != "task-1" || id2 != "task-2" {
		t.Fatalf("ids = [%s, %s], want [task-1, task-2]", id1, id2)
	}
}

// TestRepositoryListEmpty 验证空仓储时 List 返回非 nil 空切片。
func TestRepositoryListEmpty(t *testing.T) {
	repo := memory.NewRepository()
	tasks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if tasks == nil || len(tasks) != 0 {
		t.Fatalf("List() = %#v, want 非 nil 空切片", tasks)
	}
}

// TestRepositoryDeleteSuccess 验证删除后任务不存在。
func TestRepositoryDeleteSuccess(t *testing.T) {
	repo := memory.NewRepository()
	if err := repo.Save(context.Background(), newTask("task-1", "a")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := repo.Delete(context.Background(), "task-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err := repo.FindByID(context.Background(), "task-1")
	if err == nil {
		t.Fatal("FindByID() error = nil, want NOT_FOUND")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

// TestRepositoryDeleteNotFound 验证删除不存在任务返回 NOT_FOUND。
func TestRepositoryDeleteNotFound(t *testing.T) {
	repo := memory.NewRepository()
	err := repo.Delete(context.Background(), "missing")
	if err == nil {
		t.Fatal("Delete() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}
