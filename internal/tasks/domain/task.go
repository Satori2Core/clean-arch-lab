// Package domain 是 tasks 域的领域模型与端口（接口）。
// 本包不依赖任何框架/数据库/HTTP，只依赖标准库与 shared/apperr。
package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Satori2Core/clean-arch-lab/internal/shared/apperr"
)

// MaxTaskTitleLength 是任务标题最大长度（按字符数）。
const (
	MaxTaskTitleLength = 200
	// MaxTaskIDLength 是任务 ID 最大长度（内部生成器产生，对外接口校验用）。
	MaxTaskIDLength = 64
	// TaskIDPrefix 是任务 ID 的统一前缀，供内存 ID 生成器使用。
	TaskIDPrefix = "task-"
	// TaskNotFoundMessage 是任务不存在的稳定用户文案。
	TaskNotFoundMessage = "task not found"
)

// Task 是个人任务的领域实体。
type Task struct {
	// ID 是任务唯一标识，由内部 ID 生成器产生。
	ID string
	// Title 是任务标题，已 TrimSpace，长度不超过 MaxTaskTitleLength。
	Title string
	// Done 表示任务是否完成。
	Done bool
	// CreatedAt 是任务创建时间。
	CreatedAt time.Time
}

// NewTask 创建 Task 并做领域规则校验：ID 必填（ID 仅由内部生成器产生，此处只做非空校验）、标题必填且长度受限、创建时间必填。
func NewTask(id, title string, createdAt time.Time) (Task, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Task{}, apperr.New(apperr.CodeInvalidArgument, "id is required")
	}
	title, err := ValidateTitle(title)
	if err != nil {
		return Task{}, err
	}
	if createdAt.IsZero() {
		return Task{}, apperr.New(apperr.CodeInvalidArgument, "created_at is required")
	}
	return Task{ID: id, Title: title, CreatedAt: createdAt}, nil
}

// ValidateTitle 校验并规范化标题；返回 TrimSpace 后的标题，供用例层在任何副作用前先校验。
func ValidateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", apperr.New(apperr.CodeInvalidArgument, "title is required")
	}
	if utf8.RuneCountInString(title) > MaxTaskTitleLength {
		return "", apperr.New(apperr.CodeInvalidArgument, "title too long")
	}
	return title, nil
}

// MarkDone 标记任务完成。
func (t *Task) MarkDone() {
	t.Done = true
}

// ValidateID 校验外部传入的任务 ID：非空且长度受限。
func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return apperr.New(apperr.CodeInvalidArgument, "id is required")
	}
	if len(id) > MaxTaskIDLength {
		return apperr.New(apperr.CodeInvalidArgument, "id too long")
	}
	return nil
}
