package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
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

// TestNewTaskTrimsID 验证 ID 会被 TrimSpace。
func TestNewTaskTrimsID(t *testing.T) {
	task, err := domain.NewTask(" task-1 ", "标题", time.Now())
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	if task.ID != "task-1" {
		t.Fatalf("task.ID = %q, want task-1", task.ID)
	}
}

// TestNewTaskTrimsTitle 验证标题会被 TrimSpace。
func TestNewTaskTrimsTitle(t *testing.T) {
	task, err := domain.NewTask("task-1", "  买牛奶  ", time.Now())
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	if task.Title != "买牛奶" {
		t.Fatalf("task.Title = %q, want 买牛奶", task.Title)
	}
}

// TestNewTaskRejectsEmptyID 验证空 ID 返回 INVALID_ARGUMENT。
func TestNewTaskRejectsEmptyID(t *testing.T) {
	_, err := domain.NewTask("  ", "标题", time.Now())
	assertInvalidArgument(t, err)
}

// TestNewTaskRejectsEmptyTitle 验证空标题返回 INVALID_ARGUMENT。
func TestNewTaskRejectsEmptyTitle(t *testing.T) {
	_, err := domain.NewTask("task-1", "  ", time.Now())
	assertInvalidArgument(t, err)
}

// TestNewTaskBoundaryMaxLength 验证标题长度恰好等于 MaxTaskTitleLength 时创建成功。
func TestNewTaskBoundaryMaxLength(t *testing.T) {
	title := strings.Repeat("长", domain.MaxTaskTitleLength)
	task, err := domain.NewTask("task-1", title, time.Now())
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	if task.Title != title {
		t.Fatalf("task.Title 长度 = %d, want %d", len([]rune(task.Title)), domain.MaxTaskTitleLength)
	}
}

// TestNewTaskRejectsTooLongTitle 验证超长标题返回 INVALID_ARGUMENT。
func TestNewTaskRejectsTooLongTitle(t *testing.T) {
	longTitle := strings.Repeat("长", domain.MaxTaskTitleLength+1)
	_, err := domain.NewTask("task-1", longTitle, time.Now())
	assertInvalidArgument(t, err)
}

// TestNewTaskRejectsZeroCreatedAt 验证零值创建时间返回 INVALID_ARGUMENT。
func TestNewTaskRejectsZeroCreatedAt(t *testing.T) {
	_, err := domain.NewTask("task-1", "标题", time.Time{})
	assertInvalidArgument(t, err)
}

// TestMarkDone 验证 MarkDone 将 Done 置为 true。
func TestMarkDone(t *testing.T) {
	task, err := domain.NewTask("task-1", "a", time.Now())
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	task.MarkDone()
	if !task.Done {
		t.Fatal("MarkDone() 后 Done 仍为 false")
	}
}

// TestValidateIDValid 验证合法任务 ID 通过校验并返回规范化 ID。
func TestValidateIDValid(t *testing.T) {
	got, err := domain.ValidateID(" task-1 ")
	if err != nil {
		t.Fatalf("ValidateID() error = %v, want nil", err)
	}
	if got != "task-1" {
		t.Fatalf("ValidateID() = %q, want task-1", got)
	}
}

// TestValidateIDEmpty 验证空任务 ID 返回 INVALID_ARGUMENT。
func TestValidateIDEmpty(t *testing.T) {
	_, err := domain.ValidateID("  ")
	assertInvalidArgument(t, err)
}

// TestValidateIDTooLong 验证超长任务 ID 返回 INVALID_ARGUMENT。
func TestValidateIDTooLong(t *testing.T) {
	longID := strings.Repeat("a", domain.MaxTaskIDLength+1)
	_, err := domain.ValidateID(longID)
	assertInvalidArgument(t, err)
}

// TestValidateIDBoundaryMaxLength 验证 ID 长度恰好等于 MaxTaskIDLength 时通过。
func TestValidateIDBoundaryMaxLength(t *testing.T) {
	borderID := strings.Repeat("a", domain.MaxTaskIDLength)
	got, err := domain.ValidateID(borderID)
	if err != nil {
		t.Fatalf("ValidateID() error = %v, want nil", err)
	}
	if got != borderID {
		t.Fatalf("ValidateID() = %q, want %q", got, borderID)
	}
}

// TestValidateIDMultiByteBoundary 验证多字节 ID 按字符数（而非字节数）计算长度。
func TestValidateIDMultiByteBoundary(t *testing.T) {
	borderID := strings.Repeat("任", domain.MaxTaskIDLength)
	got, err := domain.ValidateID(borderID)
	if err != nil {
		t.Fatalf("ValidateID() error = %v, want nil", err)
	}
	if got != borderID {
		t.Fatalf("ValidateID() = %q, want %q", got, borderID)
	}

	tooLong := strings.Repeat("任", domain.MaxTaskIDLength+1)
	_, err = domain.ValidateID(tooLong)
	assertInvalidArgument(t, err)
}
