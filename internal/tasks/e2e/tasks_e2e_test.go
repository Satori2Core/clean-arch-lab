// Package e2e_test 是 tasks 域 HTTP E2E 测试，用例 ID 与 docs/e2e-testcases.md 对齐。
package e2e_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/e2e/harness"
)

const (
	tasksBase   = "/api/v1/tasks"
	healthzPath = "/healthz"
)

// TestE2ETaskC01CreateTask 覆盖 C01：创建任务成功返回 201 与 id/title。
func TestE2ETaskC01CreateTask(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": "买牛奶"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Done      bool   `json:"done"`
		CreatedAt string `json:"created_at"`
	}
	env.Decode(t, rec, &resp)
	if resp.ID == "" || resp.Title != "买牛奶" {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Done {
		t.Fatalf("新创建任务 Done 应为 false")
	}
	if resp.CreatedAt == "" {
		t.Fatalf("created_at 不应为空")
	}
}

// TestE2ETaskC02ListTasks 覆盖 C02：创建任务后列表返回 200 且包含该任务。
func TestE2ETaskC02ListTasks(t *testing.T) {
	env := harness.NewEnv()
	createRec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": "a"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	env.Decode(t, createRec, &created)

	rec := env.DoJSON(t, http.MethodGet, tasksBase, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var list []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	env.Decode(t, rec, &list)
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	if list[0].ID != created.ID || list[0].Title != created.Title {
		t.Fatalf("list[0] = %+v, want id=%s title=%s", list[0], created.ID, created.Title)
	}
}

// TestE2ETaskC03CreateEmptyTitle 覆盖 C03：空标题创建返回 400 与 INVALID_ARGUMENT。
func TestE2ETaskC03CreateEmptyTitle(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": ""})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	env.Decode(t, rec, &resp)
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestE2ETaskC04CreateTooLongTitle 覆盖 C04：超长标题创建返回 400 与 INVALID_ARGUMENT。
func TestE2ETaskC04CreateTooLongTitle(t *testing.T) {
	env := harness.NewEnv()
	longTitle := strings.Repeat("长", domain.MaxTaskTitleLength+1)
	rec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": longTitle})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	env.Decode(t, rec, &resp)
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestE2ETaskC05ListEmpty 覆盖 C05：空列表返回 200 且响应体为空数组。
func TestE2ETaskC05ListEmpty(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodGet, tasksBase, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var list []map[string]any
	env.Decode(t, rec, &list)
	if list == nil || len(list) != 0 {
		t.Fatalf("list = %#v, want 非 nil 空数组", list)
	}
}

// TestE2ETaskC06HealthCheck 覆盖 C06：健康检查返回 200 与 ok。
func TestE2ETaskC06HealthCheck(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodGet, healthzPath, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}

// TestE2ETaskC07MarkDone 覆盖 C07：标记任务完成返回 200 且 done=true。
func TestE2ETaskC07MarkDone(t *testing.T) {
	env := harness.NewEnv()
	createRec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": "a"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	env.Decode(t, createRec, &created)

	rec := env.DoJSON(t, http.MethodPost, tasksBase+"/"+created.ID+"/done", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Done      bool   `json:"done"`
		CreatedAt string `json:"created_at"`
	}
	env.Decode(t, rec, &resp)
	if resp.ID != created.ID || resp.Title != "a" || !resp.Done || resp.CreatedAt == "" {
		t.Fatalf("resp = %+v", resp)
	}
}

// TestE2ETaskC08MarkDoneNotFound 覆盖 C08：标记不存在的任务返回 404 与 NOT_FOUND。
func TestE2ETaskC08MarkDoneNotFound(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodPost, tasksBase+"/missing/done", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	env.Decode(t, rec, &resp)
	if resp["code"] != apperr.CodeNotFound {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeNotFound)
	}
}

// TestE2ETaskC09DeleteTask 覆盖 C09：删除任务成功返回 204。
func TestE2ETaskC09DeleteTask(t *testing.T) {
	env := harness.NewEnv()
	createRec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": "a"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	env.Decode(t, createRec, &created)

	rec := env.DoJSON(t, http.MethodDelete, tasksBase+"/"+created.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	// 验证已删除：再次删除同一 ID 应返回 404。
	again := env.DoJSON(t, http.MethodDelete, tasksBase+"/"+created.ID, nil)
	if again.Code != http.StatusNotFound {
		t.Fatalf("again status = %d, want 404; body=%s", again.Code, again.Body.String())
	}
}

// TestE2ETaskC10DeleteNotFound 覆盖 C10：删除不存在的任务返回 404 与 NOT_FOUND。
func TestE2ETaskC10DeleteNotFound(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodDelete, tasksBase+"/missing", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	env.Decode(t, rec, &resp)
	if resp["code"] != apperr.CodeNotFound {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeNotFound)
	}
}

// TestE2ETaskC11GetByID 覆盖 C11：查询任务详情成功返回 200。
func TestE2ETaskC11GetByID(t *testing.T) {
	env := harness.NewEnv()
	createRec := env.DoJSON(t, http.MethodPost, tasksBase, map[string]any{"title": "a"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	env.Decode(t, createRec, &created)

	rec := env.DoJSON(t, http.MethodGet, tasksBase+"/"+created.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Done      bool   `json:"done"`
		CreatedAt string `json:"created_at"`
	}
	env.Decode(t, rec, &resp)
	if resp.ID != created.ID || resp.Title != "a" || resp.Done || resp.CreatedAt == "" {
		t.Fatalf("resp = %+v", resp)
	}
}

// TestE2ETaskC12GetByIDNotFound 覆盖 C12：查询不存在任务返回 404 与 NOT_FOUND。
func TestE2ETaskC12GetByIDNotFound(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodGet, tasksBase+"/missing", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	env.Decode(t, rec, &resp)
	if resp["code"] != apperr.CodeNotFound {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeNotFound)
	}
}
