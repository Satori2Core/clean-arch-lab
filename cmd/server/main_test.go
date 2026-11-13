package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleHealthzC06 覆盖 C06 健康检查单测：返回 200 与 ok。
func TestHandleHealthzC06(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handleHealthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}
