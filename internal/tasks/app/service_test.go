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

// assertInvalidArgument 断言错误非 nil 且业务码为 INVALID_ARGUMENT。
func assertInvalidArgument(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("err = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidArgument {
		t.Fatalf("err = %v, want INVALID_ARGUMENT", err)
	}
}

// fakeRepo 是测试用内存仓储，可注入 Save/List/Find/Delete 错误。
type fakeRepo struct {
	tasks     []domain.Task
	saveErr   error
	listErr   error
	findErr   error
	deleteErr error
}

// Save 保存任务（按 ID 覆盖）；注入 saveErr 时返回错误。
func (f *fakeRepo) Save(_ context.Context, task domain.Task) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	for i := range f.tasks {
		if f.tasks[i].ID == task.ID {
			f.tasks[i] = task
			return nil
		}
	}
	f.tasks = append(f.tasks, task)
	return nil
}

// FindByID 按 ID 查找任务；注入 findErr 时返回错误。
func (f *fakeRepo) FindByID(_ context.Context, id string) (domain.Task, error) {
	if f.findErr != nil {
		return domain.Task{}, f.findErr
	}
	for _, task := range f.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return domain.Task{}, apperr.New(apperr.CodeNotFound, domain.TaskNotFoundMessage)
}

// List 返回任务列表；注入 listErr 时返回错误。
func (f *fakeRepo) List(_ context.Context) ([]domain.Task, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.tasks, nil
}

// Delete 删除任务；注入 deleteErr 时返回错误。
func (f *fakeRepo) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks = append(f.tasks[:i], f.tasks[i+1:]...)
			return nil
		}
	}
	return apperr.New(apperr.CodeNotFound, domain.TaskNotFoundMessage)
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

// TestMarkDoneSuccess 验证标记任务完成后 Done 变为 true。
func TestMarkDoneSuccess(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	done, err := svc.MarkDone(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("MarkDone() error = %v", err)
	}
	if !done.Done {
		t.Fatal("MarkDone() 后 Done 仍为 false")
	}
}

// TestMarkDoneNotFound 验证标记不存在的任务返回 NOT_FOUND。
func TestMarkDoneNotFound(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.MarkDone(context.Background(), "missing")
	if err == nil {
		t.Fatal("MarkDone() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

// TestMarkDoneRejectsInvalidID 验证空 ID/超长 ID 返回 INVALID_ARGUMENT。
func TestMarkDoneRejectsInvalidID(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.MarkDone(context.Background(), "  ")
	assertInvalidArgument(t, err)

	longID := strings.Repeat("a", domain.MaxTaskIDLength+1)
	_, err = svc.MarkDone(context.Background(), longID)
	assertInvalidArgument(t, err)
}

// TestMarkDoneSaveError 验证标记完成时仓储 Save 失败返回错误且业务码为 INTERNAL。
func TestMarkDoneSaveError(t *testing.T) {
	svc, repo := newTestService()
	created, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	repo.saveErr = errors.New("db down")

	_, err = svc.MarkDone(context.Background(), created.ID)
	if err == nil {
		t.Fatal("MarkDone() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("err = %v, want INTERNAL", err)
	}
}

// TestDeleteSuccess 验证删除任务成功。
func TestDeleteSuccess(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// TestDeleteNotFound 验证删除不存在的任务返回 NOT_FOUND。
func TestDeleteNotFound(t *testing.T) {
	svc, _ := newTestService()

	err := svc.Delete(context.Background(), "missing")
	if err == nil {
		t.Fatal("Delete() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

// TestDeleteRejectsInvalidID 验证空 ID/超长 ID 返回 INVALID_ARGUMENT。
func TestDeleteRejectsInvalidID(t *testing.T) {
	svc, _ := newTestService()

	err := svc.Delete(context.Background(), "  ")
	assertInvalidArgument(t, err)

	longID := strings.Repeat("a", domain.MaxTaskIDLength+1)
	err = svc.Delete(context.Background(), longID)
	assertInvalidArgument(t, err)
}

// TestDeleteRepoError 验证仓储 Delete 失败时返回错误且业务码为 INTERNAL。
func TestDeleteRepoError(t *testing.T) {
	svc, repo := newTestService()
	repo.deleteErr = errors.New("db down")

	err := svc.Delete(context.Background(), "any")
	if err == nil {
		t.Fatal("Delete() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("err = %v, want INTERNAL", err)
	}
}

// TestMarkDoneFindError 验证 MarkDone 仓储 FindByID 失败时返回错误且业务码为 INTERNAL。
func TestMarkDoneFindError(t *testing.T) {
	svc, repo := newTestService()
	repo.findErr = errors.New("db down")

	_, err := svc.MarkDone(context.Background(), "task-1")
	if err == nil {
		t.Fatal("MarkDone() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("err = %v, want INTERNAL", err)
	}
}

// TestGetByIDSuccess 验证按 ID 查询任务成功。
func TestGetByIDSuccess(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.Create(context.Background(), app.CreateTask{Title: "a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != created.ID || got.Title != created.Title {
		t.Fatalf("got = %+v, want id=%s title=%s", got, created.ID, created.Title)
	}
}

// TestGetByIDNotFound 验证查询不存在的任务返回 NOT_FOUND。
func TestGetByIDNotFound(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("GetByID() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

// TestGetByIDRejectsInvalidID 验证空 ID/超长 ID 返回 INVALID_ARGUMENT。
func TestGetByIDRejectsInvalidID(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.GetByID(context.Background(), "  ")
	assertInvalidArgument(t, err)

	longID := strings.Repeat("a", domain.MaxTaskIDLength+1)
	_, err = svc.GetByID(context.Background(), longID)
	assertInvalidArgument(t, err)
}

// TestGetByIDFindError 验证 GetByID 仓储 FindByID 失败时返回错误且业务码为 INTERNAL。
func TestGetByIDFindError(t *testing.T) {
	svc, repo := newTestService()
	repo.findErr = errors.New("db down")

	_, err := svc.GetByID(context.Background(), "task-1")
	if err == nil {
		t.Fatal("GetByID() error = nil, want error")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("err = %v, want INTERNAL", err)
	}
}
