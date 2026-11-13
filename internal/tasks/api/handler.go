// Package api 是 tasks 域的 HTTP 适配器。
// 它只做协议转换（JSON ↔ 用例入参/出参），不含业务规则。
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/app"
	"github.com/Satori2Core/clean-arch-lab/internal/tasks/domain"
)

const (
	// routeCreateTask 是创建任务路由。
	routeCreateTask = "POST /api/v1/tasks"
	// routeListTasks 是查询任务列表路由。
	routeListTasks = "GET /api/v1/tasks"
	// routeMarkTaskDone 是标记任务完成路由。
	routeMarkTaskDone = "POST /api/v1/tasks/{id}/done"

	// maxRequestBodyBytes 是创建任务请求体大小上限（1MB）。
	maxRequestBodyBytes = 1 << 20
	// internalErrorMessage 是 5xx 响应的稳定用户文案。
	internalErrorMessage = "internal error"
	// requestIDPlaceholder 是 request id 缺失或生成失败时的占位符。
	requestIDPlaceholder = "-"
	// maxRequestIDLength 是外部传入 request id 的最大长度。
	maxRequestIDLength = 64
	// requestIDRandomBytes 是随机 request id 的字节数。
	requestIDRandomBytes = 8
	// requestIDHeader 是 request id 的 HTTP 头名称。
	requestIDHeader = "X-Request-ID"
)

// ctxKey 是上下文键的私有类型，避免与其他包键冲突。
type ctxKey string

// requestIDKey 是 request id 在 context 中的键。
const requestIDKey ctxKey = "requestID"

// WithRequestID 为请求注入 request id（优先取 X-Request-ID 头，否则生成随机 hex）。
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validRequestID 校验外部传入的 request id：长度 ≤64 且仅允许字母数字与 -_。
func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, c := range id {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}

// requestID 从请求上下文提取 request id，缺失时返回占位符。
func requestID(r *http.Request) string {
	if id, ok := r.Context().Value(requestIDKey).(string); ok && id != "" {
		return id
	}
	return requestIDPlaceholder
}

// newRequestID 生成随机 hex request id；随机数失败时返回占位符。
func newRequestID() string {
	b := make([]byte, requestIDRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return requestIDPlaceholder // best-effort：随机数失败时返回占位
	}
	return hex.EncodeToString(b)
}

// Handler 是任务的 HTTP Handler。
type Handler struct {
	svc *app.Service
}

// NewHandler 构造 HTTP Handler；svc 为 nil 时直接 panic（fail-fast）。
func NewHandler(svc *app.Service) *Handler {
	if svc == nil {
		panic("api.NewHandler: svc is nil")
	}
	return &Handler{svc: svc}
}

// RegisterRoutes 注册任务 HTTP 路由。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(routeCreateTask, h.create)
	mux.HandleFunc(routeListTasks, h.list)
	mux.HandleFunc(routeMarkTaskDone, h.markDone)
}

// createTaskRequest 是创建任务请求体。
type createTaskRequest struct {
	Title string `json:"title"`
}

// taskResponse 是任务响应体。
type taskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// errorResponse 是稳定错误响应体。
type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// create 处理创建任务请求：解析 JSON、调用用例并返回 201。
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, r, apperr.New(apperr.CodeInvalidArgument, "request body too large"), http.StatusBadRequest)
			return
		}
		writeError(w, r, apperr.Wrapf(err, "read request body failed"), http.StatusInternalServerError)
		return
	}
	var req createTaskRequest
	// json.Unmarshal 会拒绝尾随数据与非 JSON 内容；未知字段不在拒绝范围（契约未要求）。
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, r, apperr.New(apperr.CodeInvalidArgument, "invalid json body"), http.StatusBadRequest)
		return
	}
	task, err := h.svc.Create(r.Context(), app.CreateTask{Title: req.Title})
	if err != nil {
		writeError(w, r, err, statusFromErr(err))
		return
	}
	writeJSON(w, r, http.StatusCreated, toResponse(task))
}

// list 处理任务列表请求并返回任务数组。
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, r, err, statusFromErr(err))
		return
	}
	resp := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		resp = append(resp, toResponse(task))
	}
	writeJSON(w, r, http.StatusOK, resp)
}

// markDone 处理标记任务完成请求。
func (h *Handler) markDone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.svc.MarkDone(r.Context(), id)
	if err != nil {
		writeError(w, r, err, statusFromErr(err))
		return
	}
	writeJSON(w, r, http.StatusOK, toResponse(task))
}

// toResponse 将领域 Task 转换为 HTTP 响应结构。
func toResponse(task domain.Task) taskResponse {
	return taskResponse{
		ID:        task.ID,
		Title:     task.Title,
		Done:      task.Done,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	}
}

// writeJSON 统一设置 JSON Content-Type 并编码响应体；编码失败只记录一次日志并返回固定 500。
func writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		slog.Default().ErrorContext(r.Context(), "marshal response failed", "err", err, "request_id", requestID(r), "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(errorResponse{Code: apperr.CodeInternal, Message: internalErrorMessage}) // 简单结构不会编码失败
		_, _ = w.Write(data)                                                                             // best-effort：响应写入失败已无法挽回
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(data) // best-effort：响应写入失败已无法挽回
}

// writeError 按状态码记录日志并输出稳定错误响应：4xx 返回错误码文案，5xx 不泄露内部细节。
func writeError(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status >= http.StatusInternalServerError {
		slog.Default().ErrorContext(r.Context(), "request failed", "err", apperr.Detail(err), "request_id", requestID(r), "method", r.Method, "path", r.URL.Path)
		writeJSON(w, r, status, errorResponse{Code: codeOf(err), Message: internalErrorMessage})
		return
	}
	slog.Default().InfoContext(r.Context(), "request rejected", "err", err, "request_id", requestID(r), "method", r.Method, "path", r.URL.Path)
	writeJSON(w, r, status, errorResponse{Code: codeOf(err), Message: err.Error()})
}

// codeOf 从错误链中提取稳定业务错误码，缺省返回 INTERNAL。
func codeOf(err error) string {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return apperr.CodeInternal
}

// statusFromErr 将业务错误码映射为 HTTP 状态码。
func statusFromErr(err error) int {
	switch codeOf(err) {
	case apperr.CodeInvalidArgument:
		return http.StatusBadRequest
	case apperr.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
