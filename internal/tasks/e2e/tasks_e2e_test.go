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

// TestE2ETaskC01CreateTask 覆盖 C01：创建任务成功返回 201 与 id/title。
func TestE2ETaskC01CreateTask(t *testing.T) {
	env := harness.NewEnv()
	rec := env.DoJSON(t, http.MethodPost, "/api/v1/tasks", map[string]any{"title": "买牛奶"})

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
	createRec := env.DoJSON(t, http.MethodPost, "/api/v1/tasks", map[string]any{"title": "a"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	env.Decode(t, createRec, &created)

	rec := env.DoJSON(t, http.MethodGet, "/api/v1/tasks", nil)
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
	rec := env.DoJSON(t, http.MethodPost, "/api/v1/tasks", map[string]any{"title": ""})

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
	rec := env.DoJSON(t, http.MethodPost, "/api/v1/tasks", map[string]any{"title": longTitle})

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
	rec := env.DoJSON(t, http.MethodGet, "/api/v1/tasks", nil)

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
	rec := env.DoJSON(t, http.MethodGet, "/healthz", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}
