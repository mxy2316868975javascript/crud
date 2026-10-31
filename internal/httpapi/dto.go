package httpapi

import "github.com/mxy2316868975javascript/crud/internal/domain"

// WordInput 是新增或全量更新词汇的请求体。
type WordInput struct {
	Word    string `json:"word" example:"abandon"`     // 单词
	Meaning string `json:"meaning" example:"v. 放弃；抛弃"` // 中文释义
	Level   int    `json:"level" example:"1"`          // 难度等级：1-5
}

// WordPatchInput 是局部更新词汇的请求体，nil 表示不修改该字段。
type WordPatchInput struct {
	Word    *string `json:"word" example:"abandon"`     // 新单词
	Meaning *string `json:"meaning" example:"v. 放弃；抛弃"` // 新释义
	Level   *int    `json:"level" example:"1"`          // 新难度等级：1-5
}

// WordListResponse 是词汇列表的响应体。
type WordListResponse struct {
	Items []domain.Word `json:"items"`                // 当前页词汇
	Total int           `json:"total" example:"3849"` // 符合条件的总数
}

// CreateTestRequest 是创建测试的请求体。
type CreateTestRequest struct {
	Count int `json:"count" example:"20"` // 题目数量，省略或为 0 时默认 20
}

// SubmitAnswersRequest 是提交答案的请求体。
type SubmitAnswersRequest struct {
	Answers []domain.Answer `json:"answers"` // 答案列表，必须覆盖全部题目
}

// ErrorResponse 是所有错误响应的统一结构。
type ErrorResponse struct {
	Error string `json:"error" example:"not found"` // 错误信息
}

// HealthResponse 是健康检查的响应体。
type HealthResponse struct {
	Status string `json:"status" example:"ok"` // ok 或 unavailable
}

func (input WordInput) toDomain() domain.Word {
	return domain.Word{Word: input.Word, Meaning: input.Meaning, Level: input.Level}
}

func (input WordPatchInput) toDomain() domain.WordPatch {
	return domain.WordPatch{Word: input.Word, Meaning: input.Meaning, Level: input.Level}
}
