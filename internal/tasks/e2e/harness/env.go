// Package harness 提供 E2E 共享环境：装配内存仓储 + 服务 + HTTP 路由。
// 非测试包，供各 E2E 场景复用。
package harness

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Satori2Core/clean-arch-lab/internal/tasks/api"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/app"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/infra/memory"
)

// Env 是 E2E 共享环境。
type Env struct {
	// Mux 是已注册任务路由与健康检查路由的 Handler，供 E2E 场景直接发送请求。
	Mux http.Handler
}

// NewEnv 创建共享环境：内存仓储 + 应用服务 + HTTP 路由。
func NewEnv() *Env {
	repo := memory.NewRepository()
	ids := memory.NewIDGenerator(domain.TaskIDPrefix)
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: ids})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok")) // best-effort：健康检查响应写入失败无需处理
	})
	api.NewHandler(svc).RegisterRoutes(mux)
	return &Env{Mux: api.WithRequestID(mux)}
}

// DoJSON 发送 JSON 请求并返回 recorder。
func (e *Env) DoJSON(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body error = %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	e.Mux.ServeHTTP(rec, req)
	return rec
}

// Decode 解析 JSON 响应。
func (e *Env) Decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("unmarshal response error = %v, body=%s", err, rec.Body.String())
	}
}
