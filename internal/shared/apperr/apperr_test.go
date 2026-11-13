package apperr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
)

// TestNewReturnsErrorInterface 验证 New 返回 error 接口而非具体类型。
func TestNewReturnsErrorInterface(t *testing.T) {
	var err error = apperr.New(apperr.CodeInvalidArgument, "bad")
	if err == nil {
		t.Fatal("New() error = nil, want non-nil error")
	}
}

// TestWrapfInheritsCode 验证 Wrapf 继承内层业务码。
func TestWrapfInheritsCode(t *testing.T) {
	base := apperr.New(apperr.CodeNotFound, "not found")
	wrapped := apperr.Wrapf(base, "load failed")

	var ae *apperr.Error
	if !errors.As(wrapped, &ae) || ae.Code != apperr.CodeNotFound {
		t.Fatalf("wrapped code = %v, want NOT_FOUND", ae)
	}
}

// TestWrapfDefaultInternal 验证无业务码的原始错误包装后默认 INTERNAL。
func TestWrapfDefaultInternal(t *testing.T) {
	wrapped := apperr.Wrapf(errors.New("boom"), "op failed")

	var ae *apperr.Error
	if !errors.As(wrapped, &ae) || ae.Code != apperr.CodeInternal {
		t.Fatalf("wrapped code = %v, want INTERNAL", ae)
	}
}

// TestErrorHidesCause 验证 Error() 不暴露内部 Cause，避免用户可见消息泄露内部细节。
func TestErrorHidesCause(t *testing.T) {
	wrapped := apperr.Wrapf(errors.New("secret detail"), "user message")
	if strings.Contains(wrapped.Error(), "secret detail") {
		t.Fatal("Error() 不应包含内部 Cause")
	}
}

// TestDetailContainsCauseAndStack 验证 Detail() 包含 Cause 与调用栈，供日志排查。
func TestDetailContainsCauseAndStack(t *testing.T) {
	wrapped := apperr.Wrapf(errors.New("secret detail"), "user message")
	detail := apperr.Detail(wrapped)
	if !strings.Contains(detail, "secret detail") {
		t.Fatal("Detail() 应包含 Cause")
	}
	if !strings.Contains(detail, "apperr_test") {
		t.Fatal("Detail() 应包含调用栈")
	}
}

// TestDetailNil 验证 Detail(nil) 返回空字符串。
func TestDetailNil(t *testing.T) {
	if got := apperr.Detail(nil); got != "" {
		t.Fatalf("Detail(nil) = %q, want empty", got)
	}
}

// TestDetailPlainError 验证非 *Error 的普通错误返回其 Error() 文本。
func TestDetailPlainError(t *testing.T) {
	plain := errors.New("plain error")
	if got := apperr.Detail(plain); got != "plain error" {
		t.Fatalf("Detail(plain) = %q, want plain error", got)
	}
}

// TestWrapfNilPanics 验证 Wrapf(nil, ...) 直接 panic（fail-fast）。
func TestWrapfNilPanics(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("Wrapf(nil) 未 panic")
		}
	}()
	_ = apperr.Wrapf(nil, "should panic") // 预期 panic，返回值不可达
}

// TestUnwrap 验证 errors.Is 能沿错误链找到原始错误。
func TestUnwrap(t *testing.T) {
	base := errors.New("root")
	wrapped := apperr.Wrapf(base, "ctx")
	if !errors.Is(wrapped, base) {
		t.Fatal("errors.Is(wrapped, base) = false, want true")
	}
}

// TestDetailContainsCode 验证 Detail() 包含业务码，便于日志区分错误类型。
func TestDetailContainsCode(t *testing.T) {
	err := apperr.New(apperr.CodeNotFound, "not found")
	detail := apperr.Detail(err)
	if !strings.Contains(detail, "[NOT_FOUND]") {
		t.Fatalf("Detail() 应包含业务码，got: %s", detail)
	}
}

// TestWrapfPreservesBusinessMessage 验证包装业务码错误时保留原始用户文案，上下文只进 Detail。
func TestWrapfPreservesBusinessMessage(t *testing.T) {
	base := apperr.New(apperr.CodeNotFound, "not found")
	wrapped := apperr.Wrapf(base, "load failed")

	if got := wrapped.Error(); got != "not found" {
		t.Fatalf("Error() = %q, want 原始文案 not found", got)
	}
	detail := apperr.Detail(wrapped)
	if !strings.Contains(detail, "context: load failed") {
		t.Fatalf("Detail() 应包含包装上下文，got: %s", detail)
	}
}
