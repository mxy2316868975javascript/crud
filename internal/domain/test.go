package domain

import (
	"math"
	"time"
)

const (
	minQuestionCount = 1
	maxQuestionCount = 100
)

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

// ValidateQuestionCount 校验题目数量是否在 1–100 之间。
func ValidateQuestionCount(count int) error {
	if count < minQuestionCount || count > maxQuestionCount {
		return &ValidationError{Message: "count must be between 1 and 100"}
	}
	return nil
}

// ValidateAnswerShape 校验答案下标和重复的 word_id。
func ValidateAnswerShape(answers []Answer) error {
	seenWordIDs := make(map[int64]bool, len(answers))
	for _, answer := range answers {
		if answer.SelectedIndex < 0 || answer.SelectedIndex > 3 || seenWordIDs[answer.WordID] {
			return ErrInvalidAnswers
		}
		seenWordIDs[answer.WordID] = true
	}
	return nil
}

// ValidateAnswerCoverage 校验答案是否正好覆盖给定的题目集合。
func ValidateAnswerCoverage(answers []Answer, expectedWordIDs []int64) error {
	if len(answers) != len(expectedWordIDs) {
		return ErrInvalidAnswers
	}

	expected := make(map[int64]bool, len(expectedWordIDs))
	for _, wordID := range expectedWordIDs {
		expected[wordID] = true
	}

	seenWordIDs := make(map[int64]bool, len(answers))
	for _, answer := range answers {
		if answer.SelectedIndex < 0 || answer.SelectedIndex > 3 ||
			!expected[answer.WordID] || seenWordIDs[answer.WordID] {
			return ErrInvalidAnswers
		}
		seenWordIDs[answer.WordID] = true
	}
	return nil
}

// EstimateVocabulary 按正确率估算词汇量，题目数为 0 时返回 0。
func EstimateVocabulary(correctCount, questionCount, totalWords int) int {
	if questionCount <= 0 {
		return 0
	}
	return int(math.Round(float64(correctCount) / float64(questionCount) * float64(totalWords)))
}
