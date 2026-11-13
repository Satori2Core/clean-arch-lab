package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Satori2Core/clean-arch-lab/internal/tasks/api"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/app"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/infra/memory"
)

// defaultAddr 是本地练手服务的监听地址。
const defaultAddr = ":8080"

// main 是组合根：创建依赖、注册路由并启动 HTTP 服务。
func main() {
	// 统一 JSON 日志格式（对应 AGENTS.md 日志规范）。
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 组合根：所有依赖在这里显式创建并注入，避免全局单例。
	repo := memory.NewRepository()
	ids := memory.NewIDGenerator(domain.TaskIDPrefix)
	svc := app.NewService(app.ServiceDeps{Repo: repo, IDs: ids})
	handler := api.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	handler.RegisterRoutes(mux)

	slog.Info("clean-arch-lab listening", "addr", defaultAddr)
	if err := http.ListenAndServe(defaultAddr, api.WithRequestID(mux)); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// handleHealthz 返回健康检查 OK。
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok")) // best-effort：健康检查响应写入失败无需处理
}
