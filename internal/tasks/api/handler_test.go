package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/api"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/app"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/infra/memory"
)

// oversizedRequestBodyBytes 是测试用超限请求体大小（对应 api 层 1MB 上限 +1）。
const oversizedRequestBodyBytes = 1<<20 + 1

// newTestHandler 构造基于内存仓储的 HTTP Handler。
func newTestHandler() *api.Handler {
	repo := memory.NewRepository()
	ids := memory.NewIDGenerator(domain.TaskIDPrefix)
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: ids})
	return api.NewHandler(svc)
}

// newTestMux 注册 Handler 路由。
func newTestMux(h *api.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// failingRepo 是测试用失败仓储，用于验证 5xx 不泄露内部错误细节。
type failingRepo struct{}

// Save 返回带敏感细节的内部错误。
func (failingRepo) Save(context.Context, domain.Task) error {
	return apperr.New(apperr.CodeInternal, "secret db detail")
}

// FindByID 返回未找到。
func (failingRepo) FindByID(context.Context, string) (domain.Task, error) {
	return domain.Task{}, apperr.New(apperr.CodeNotFound, domain.TaskNotFoundMessage)
}

// List 返回带敏感细节的内部错误。
func (failingRepo) List(context.Context) ([]domain.Task, error) {
	return nil, apperr.New(apperr.CodeInternal, "secret db detail")
}

// Delete 返回带敏感细节的内部错误。
func (failingRepo) Delete(context.Context, string) error {
	return apperr.New(apperr.CodeInternal, "secret db detail")
}

// TestCreateTaskHTTPInvalidJSON 覆盖非法 JSON 返回 400 与 INVALID_ARGUMENT。
func TestCreateTaskHTTPInvalidJSON(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`not-json`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestCreateTaskHTTPTrailingData 覆盖合法 JSON 后跟尾随数据返回 400。
func TestCreateTaskHTTPTrailingData(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"} extra`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestCreateTaskHTTPC03EmptyTitle 覆盖 C03 空标题返回 400 与 INVALID_ARGUMENT。
func TestCreateTaskHTTPC03EmptyTitle(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":""}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestCreateTaskHTTPC04TooLongTitle 覆盖 C04 超长标题返回 400 与 INVALID_ARGUMENT。
func TestCreateTaskHTTPC04TooLongTitle(t *testing.T) {
	mux := newTestMux(newTestHandler())
	longTitle := strings.Repeat("长", domain.MaxTaskTitleLength+1)
	body := `{"title":"` + longTitle + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestCreateTaskHTTPInternalErrorHidesDetail 覆盖 5xx 响应不泄露内部错误细节。
func TestCreateTaskHTTPInternalErrorHidesDetail(t *testing.T) {
	svc := app.NewService(app.ServiceDeps{Repo: failingRepo{}, IDs: memory.NewIDGenerator(domain.TaskIDPrefix)})
	mux := newTestMux(api.NewHandler(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret db detail") {
		t.Fatalf("5xx 响应不应泄露内部错误细节: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "internal error") {
		t.Fatalf("5xx 响应应返回通用文案: %s", rec.Body.String())
	}
}

// TestListTasksHTTPInternalErrorHidesDetail 覆盖 GET 列表 5xx 不泄露内部错误细节。
func TestListTasksHTTPInternalErrorHidesDetail(t *testing.T) {
	svc := app.NewService(app.ServiceDeps{Repo: failingRepo{}, IDs: memory.NewIDGenerator(domain.TaskIDPrefix)})
	mux := newTestMux(api.NewHandler(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret db detail") {
		t.Fatalf("5xx 响应不应泄露内部错误细节: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "internal error") {
		t.Fatalf("5xx 响应应返回通用文案: %s", rec.Body.String())
	}
}

// TestListTasksHTTPC05Empty 覆盖 C05 空列表返回 200 与空数组。
func TestListTasksHTTPC05Empty(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if list == nil || len(list) != 0 {
		t.Fatalf("list = %#v, want 非 nil 空数组", list)
	}
}

// TestCreateTaskHTTPC01Success 覆盖 C01 创建成功返回 201 与响应字段。
func TestCreateTaskHTTPC01Success(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"买牛奶"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Done      bool   `json:"done"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp.ID == "" || resp.Title != "买牛奶" || resp.Done || resp.CreatedAt == "" {
		t.Fatalf("resp = %+v", resp)
	}
}

// TestListTasksHTTPC02Success 覆盖 C02 列表成功返回 200 且包含已创建任务。
func TestListTasksHTTPC02Success(t *testing.T) {
	mux := newTestMux(newTestHandler())
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body=%s", createRec.Code, createRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
}

// TestCreateTaskHTTPC01BoundaryMaxLength 覆盖标题长度恰好等于 MaxTaskTitleLength 时创建成功。
func TestCreateTaskHTTPC01BoundaryMaxLength(t *testing.T) {
	mux := newTestMux(newTestHandler())
	title := strings.Repeat("长", domain.MaxTaskTitleLength)
	body := `{"title":"` + title + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
}

// TestCreateTaskHTTPRequestBodyTooLarge 覆盖请求体超过上限返回 400 与 INVALID_ARGUMENT。
func TestCreateTaskHTTPRequestBodyTooLarge(t *testing.T) {
	mux := newTestMux(newTestHandler())
	hugeBody := `{"title":"` + strings.Repeat("a", oversizedRequestBodyBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(hugeBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestWithRequestIDValidHeader 验证合法 X-Request-ID 会被回写到响应头。
func TestWithRequestIDValidHeader(t *testing.T) {
	var handler http.Handler = api.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "abc123" {
		t.Fatalf("X-Request-ID = %q, want abc123", got)
	}
}

// TestWithRequestIDInvalidHeaderGeneratesNew 验证非法 X-Request-ID 会被替换为新生成 ID。
func TestWithRequestIDInvalidHeaderGeneratesNew(t *testing.T) {
	var handler http.Handler = api.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "bad id!!")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if got == "bad id!!" || got == "" {
		t.Fatalf("X-Request-ID = %q, want 新生成的合法 ID", got)
	}
}

// TestWithRequestIDTooLongHeaderGeneratesNew 验证超长 X-Request-ID 会被替换为新生成 ID。
func TestWithRequestIDTooLongHeaderGeneratesNew(t *testing.T) {
	var handler http.Handler = api.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", strings.Repeat("a", domain.MaxTaskIDLength+1))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if len(got) > domain.MaxTaskIDLength || got == "" {
		t.Fatalf("X-Request-ID = %q, want 新的合法 ID", got)
	}
}

// TestWithRequestIDSetsResponseHeaderWhenMissing 验证缺失请求头时也会生成并回写 request id。
func TestWithRequestIDSetsResponseHeaderWhenMissing(t *testing.T) {
	var handler http.Handler = api.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("X-Request-ID 为空，应生成新 ID")
	}
}

// TestMarkTaskDoneHTTPC07 覆盖标记任务完成返回 200 且 done=true。
func TestMarkTaskDoneHTTPC07(t *testing.T) {
	mux := newTestMux(newTestHandler())
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+created.ID+"/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Done bool `json:"done"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response error = %v", err)
	}
	if !resp.Done {
		t.Fatal("done = false, want true")
	}
}

// TestMarkTaskDoneHTTPC08 覆盖标记不存在的任务返回 404 与 NOT_FOUND。
func TestMarkTaskDoneHTTPC08(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/missing/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeNotFound {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeNotFound)
	}
}

// TestDeleteTaskHTTPC09 覆盖删除任务成功返回 204。
func TestDeleteTaskHTTPC09(t *testing.T) {
	mux := newTestMux(newTestHandler())
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create error = %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+created.ID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}

// TestDeleteTaskHTTPC10 覆盖删除不存在的任务返回 404 与 NOT_FOUND。
func TestDeleteTaskHTTPC10(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeNotFound {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeNotFound)
	}
}

// TestMarkTaskDoneHTTPInvalidID 覆盖非法 ID 返回 400 与 INVALID_ARGUMENT。
func TestMarkTaskDoneHTTPInvalidID(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/%20%20/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}

// TestDeleteTaskHTTPInvalidID 覆盖非法 ID 返回 400 与 INVALID_ARGUMENT。
func TestDeleteTaskHTTPInvalidID(t *testing.T) {
	mux := newTestMux(newTestHandler())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/%20%20", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if resp["code"] != apperr.CodeInvalidArgument {
		t.Fatalf("code = %v, want %s", resp["code"], apperr.CodeInvalidArgument)
	}
}
