package domain

import "errors"

// 领域错误。HTTP 层根据这些错误映射状态码和错误消息。
var (
	// ErrNotFound 表示资源不存在。
	ErrNotFound = errors.New("not found")
	// ErrWordExists 表示单词已经存在。
	ErrWordExists = errors.New("word already exists")
	// ErrWordInUse 表示词汇已被测试引用，不能删除。
	ErrWordInUse = errors.New("word is used by a test")
	// ErrNotEnoughWords 表示词库数量不足，无法抽题。
	ErrNotEnoughWords = errors.New("not enough words in the bank")
	// ErrInvalidAnswers 表示答案没有覆盖全部题目或包含重复项。
	ErrInvalidAnswers = errors.New("answers must cover every question exactly once")
	// ErrTestCompleted 表示测试已经提交过，不能重复提交。
	ErrTestCompleted = errors.New("conflict")
)

// ValidationError 表示用户输入不合法，Message 会直接返回给调用方。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
