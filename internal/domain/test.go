package domain

import "time"

// Question 是一道四选一题目，Options 中不包含正确选项下标。
type Question struct {
	WordID  int64    `json:"word_id" example:"1"`    // 词汇 ID
	Word    string   `json:"word" example:"abandon"` // 题目单词
	Options []string `json:"options"`                // 四个选项，正确答案下标不返回
}

// Test 表示一次词汇量测试。
type Test struct {
	ID                  int64      `json:"id" example:"1"`                                // 测试 ID
	QuestionCount       int        `json:"question_count" example:"20"`                   // 题目数量
	Completed           bool       `json:"completed" example:"false"`                     // 是否已提交答案
	Questions           []Question `json:"questions,omitempty"`                           // 题目列表
	CorrectCount        *int       `json:"correct_count,omitempty" example:"15"`          // 答对题数，未完成时为 null
	EstimatedVocabulary *int       `json:"estimated_vocabulary,omitempty" example:"2887"` // 估算词汇量，未完成时为 null
	CompletedAt         *time.Time `json:"completed_at,omitempty"`                        // 提交时间，未完成时为 null
}

// Answer 表示用户对一道题的作答。
type Answer struct {
	WordID        int64 `json:"word_id" example:"1"`                                // 题目对应的词汇 ID
	SelectedIndex int   `json:"selected_index" example:"0" minimum:"0" maximum:"3"` // 用户选择的选项下标 0-3
}
