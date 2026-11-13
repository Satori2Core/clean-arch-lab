package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/app"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
)

// fakeRepo 是测试用内存仓储，可注入 Save/List 错误。
type fakeRepo struct {
	tasks   []domain.Task
	saveErr error
	listErr error
}

// Save 保存任务；注入 saveErr 时返回错误。
func (f *fakeRepo) Save(_ context.Context, task domain.Task) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.tasks = append(f.tasks, task)
	return nil
}

// FindByID 按 ID 查找任务。
func (f *fakeRepo) FindByID(_ context.Context, id string) (domain.Task, error) {
	for _, task := range f.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return domain.Task{}, apperr.New(apperr.CodeNotFound, "task not found")
}

// List 返回任务列表；注入 listErr 时返回错误。
func (f *fakeRepo) List(_ context.Context) ([]domain.Task, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.tasks, nil
}

// fakeIDGen 是测试用 ID 生成器，可注入错误。
type fakeIDGen struct {
	n   int
	err error
}

// NewID 生成递增 ID；注入 err 时返回错误。
func (f *fakeIDGen) NewID(_ context.Context) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.n++
	return fmt.Sprintf("id-%d", f.n), nil
}

// newTestService 构造测试 Service 与 fakeRepo。
func newTestService() (*app.Service, *fakeRepo) {
	repo := &fakeRepo{}
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: &fakeIDGen{}})
	return svc, repo
}

// TestCreateTask 验证创建任务成功并 TrimSpace 标题。
func TestCreateTask(t *testing.T) {
	svc, repo := newTestService()

	task, err := svc.Create(context.Background(), app.CreateTask{Title: "  买牛奶  "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.ID != "id-1" {
		t.Fatalf("task.ID = %q, want id-1", task.ID)
	}
	if task.Title != "买牛奶" {
		t.Fatalf("task.Title = %q, want 买牛奶（应 TrimSpace）", task.Title)
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("repo.tasks len = %d, want 1", len(repo.tasks))
	}
}

// TestCreateTaskBoundaryMaxLength 验证标题恰好等于 MaxTaskTitleLength 时创建成功。
func TestCreateTaskBoundaryMaxLength(t *testing.T) {
	svc, _ := newTestService()

	title := strings.Repeat("长", domain.MaxTaskTitleLength)
	task, err := svc.Create(context.Background(), app.CreateTask{Title: title})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.Title != title {
		t.Fatalf("task.Title 长度 = %d, want %d", len([]rune(task.Title)), domain.MaxTaskTitleLength)
	}
}

// TestCreateTaskRejectsEmptyTitle 验证空标题返回 INVALID_ARGUMENT。
func TestCreateTaskRejectsEmptyTitle(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.Create(context.Background(), app.CreateTask{Title: "  "})
	if err == nil {
		t.Fatal("Create() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidArgument {
		t.Fatalf("error = %v, want INVALID_ARGUMENT", err)
	}
}

// TestCreateTaskRejectsTooLongTitle 验证超长标题返回 INVALID_ARGUMENT。
func TestCreateTaskRejectsTooLongTitle(t *testing.T) {
	svc, _ := newTestService()

	longTitle := strings.Repeat("长", domain.MaxTaskTitleLength+1)
	_, err := svc.Create(context.Background(), app.CreateTask{Title: longTitle})
	if err == nil {
		t.Fatal("Create() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidArgument {
		t.Fatalf("error = %v, want INVALID_ARGUMENT", err)
	}
}

// TestList 验证 List 返回全部任务。
func TestList(t *testing.T) {
	svc, _ := newTestService()
	if _, err := svc.Create(context.Background(), app.CreateTask{Title: "a"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(context.Background(), app.CreateTask{Title: "b"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tasks, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("List() len = %d, want 2", len(tasks))
	}
}

// TestCreateTaskIDGenError 验证 ID 生成失败时返回错误且业务码为 INTERNAL。
func TestCreateTaskIDGenError(t *testing.T) {
	repo := &fakeRepo{}
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: &fakeIDGen{err: errors.New("idgen down")}})

	_, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err == nil {
		t.Fatal("Create() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("error = %v, want INTERNAL", err)
	}
}

// TestCreateTaskSaveError 验证仓储保存失败时返回错误且业务码为 INTERNAL。
func TestCreateTaskSaveError(t *testing.T) {
	repo := &fakeRepo{saveErr: errors.New("db down")}
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: &fakeIDGen{}})

	_, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err == nil {
		t.Fatal("Create() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("error = %v, want INTERNAL", err)
	}
}

// TestListError 验证仓储列表失败时返回错误且业务码为 INTERNAL。
func TestListError(t *testing.T) {
	repo := &fakeRepo{listErr: errors.New("db down")}
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: &fakeIDGen{}})

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("List() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("error = %v, want INTERNAL", err)
	}
}

// TestListEmpty 验证空仓储时 List 返回空切片（非 nil）。
func TestListEmpty(t *testing.T) {
	svc, _ := newTestService()
	tasks, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if tasks == nil || len(tasks) != 0 {
		t.Fatalf("List() = %#v, want 非 nil 空切片", tasks)
	}
}

// TestNewServiceNilRepoPanics 验证 Repo 为 nil 时构造 Service 直接 panic。
func TestNewServiceNilRepoPanics(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("NewService(Repo=nil) 未 panic")
		}
	}()
	_ = app.NewService(app.ServiceDeps{Repo: nil, IDs: &fakeIDGen{}}) // 预期 panic，返回值不可达
}

// TestNewServiceNilIDsPanics 验证 IDs 为 nil 时构造 Service 直接 panic。
func TestNewServiceNilIDsPanics(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("NewService(IDs=nil) 未 panic")
		}
	}()
	_ = app.NewService(app.ServiceDeps{Repo: &fakeRepo{}, IDs: nil}) // 预期 panic，返回值不可达
}
