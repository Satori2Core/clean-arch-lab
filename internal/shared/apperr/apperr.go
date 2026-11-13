// Package apperr 提供带业务错误码、上下文和堆栈的错误类型。
// 本包是练手项目自带的迷你错误库，仅依赖标准库。
package apperr

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

const (
	// CodeInvalidArgument 表示参数校验失败，映射 HTTP 400。
	CodeInvalidArgument = "INVALID_ARGUMENT"
	// CodeNotFound 表示资源不存在，映射 HTTP 404。
	CodeNotFound = "NOT_FOUND"
	// CodeInternal 表示内部错误，映射 HTTP 500。
	CodeInternal = "INTERNAL"

	maxStackFrames   = 32 // maxStackFrames 是捕获调用栈的最大帧数
	callerSkipFrames = 3  // callerSkipFrames 是跳过 Callers 自身与包装函数的帧数
)

// Error 是带业务错误码、用户可见消息、原始错误、包装上下文和调用栈的错误类型。
type Error struct {
	// Code 是稳定业务错误码（INVALID_ARGUMENT / NOT_FOUND / INTERNAL）。
	Code string
	// Message 是用户可见的错误文案。
	Message string
	// Context 是 Wrapf 追加的排查上下文（不面向用户）。
	Context string
	// Cause 是原始错误，供日志排查与 errors.Is/As 链式查找。
	Cause error
	stack string
}

// Error 返回用户可见消息（不含内部 Cause，避免透传内部细节）。
func (e *Error) Error() string {
	return e.Message
}

// Unwrap 支持 errors.Is / errors.As 链式查找。
func (e *Error) Unwrap() error {
	return e.Cause
}

// Detail 返回含业务码、包装上下文、Cause 与调用栈的完整描述，用于日志排查。
func (e *Error) Detail() string {
	parts := []string{e.Message}
	if e.Context != "" {
		parts = append(parts, "context: "+e.Context)
	}
	if e.Cause != nil {
		if inner, ok := e.Cause.(*Error); ok {
			parts = append(parts, "cause: "+inner.Detail())
		} else {
			parts = append(parts, "cause: "+e.Cause.Error())
		}
	}
	return fmt.Sprintf("[%s] %s\n%s", e.Code, strings.Join(parts, "; "), e.stack)
}

// New 创建无原始错误的业务错误，并捕获当前调用栈。
func New(code, message string) error {
	return &Error{Code: code, Message: message, stack: captureStack()}
}

// Wrapf 包装原始错误并带上业务上下文；若原始错误本身是 *Error，则继承其业务码。
// 对 INVALID_ARGUMENT/NOT_FOUND 等业务码保留原始用户 Message，包装上下文存入 Context 供日志排查。
// err 为 nil 时直接 panic（fail-fast，避免掩盖编程错误）。
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		panic("apperr.Wrapf: err is nil")
	}
	contextMsg := fmt.Sprintf(format, args...)
	code := CodeInternal
	message := contextMsg
	var ae *Error
	if errors.As(err, &ae) {
		code = ae.Code
		if code != CodeInternal && ae.Message != "" {
			message = ae.Message
		}
	}
	return &Error{
		Code:    code,
		Message: message,
		Context: contextMsg,
		Cause:   err,
		stack:   captureStack(),
	}
}

// Detail 返回错误的完整排查信息；非 *Error 直接返回 Error()，nil 返回空字符串。
func Detail(err error) string {
	if err == nil {
		return ""
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Detail()
	}
	return err.Error()
}

// captureStack 捕获当前调用栈的简短文本，供日志排查使用。
func captureStack() string {
	pcs := make([]uintptr, maxStackFrames)
	n := runtime.Callers(callerSkipFrames, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	for {
		f, more := frames.Next()
		_, _ = fmt.Fprintf(&b, "%s\n\t%s:%d\n", f.Function, f.File, f.Line) // strings.Builder 写入不会失败
		if !more {
			break
		}
	}
	return b.String()
}
