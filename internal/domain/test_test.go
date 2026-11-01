package domain

import (
	"errors"
	"testing"
)

func TestValidateQuestionCount(t *testing.T) {
	valid := []int{1, 20, 100}
	for _, count := range valid {
		if err := ValidateQuestionCount(count); err != nil {
			t.Errorf("count %d: unexpected error: %v", count, err)
		}
	}

	invalid := []int{-1, 0, 101}
	for _, count := range invalid {
		err := ValidateQuestionCount(count)
		if err == nil || err.Error() != "count must be between 1 and 100" {
			t.Errorf("count %d: error = %v, want count must be between 1 and 100", count, err)
		}
	}
}

func TestValidateAnswerShape(t *testing.T) {
	tests := []struct {
		name    string
		answers []Answer
		wantErr bool
	}{
		{
			name:    "合法答案",
			answers: []Answer{{WordID: 1, SelectedIndex: 0}, {WordID: 2, SelectedIndex: 3}},
		},
		{
			name:    "空答案",
			answers: nil,
		},
		{
			name:    "下标太小",
			answers: []Answer{{WordID: 1, SelectedIndex: -1}},
			wantErr: true,
		},
		{
			name:    "下标太大",
			answers: []Answer{{WordID: 1, SelectedIndex: 4}},
			wantErr: true,
		},
		{
			name:    "重复 word_id",
			answers: []Answer{{WordID: 1, SelectedIndex: 0}, {WordID: 1, SelectedIndex: 1}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAnswerShape(test.answers)
			if test.wantErr {
				if !errors.Is(err, ErrInvalidAnswers) {
					t.Fatalf("error = %v, want ErrInvalidAnswers", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAnswerCoverage(t *testing.T) {
	expected := []int64{1, 2, 3}
	tests := []struct {
		name    string
		answers []Answer
		wantErr bool
	}{
		{
			name: "完整覆盖",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 2, SelectedIndex: 1},
				{WordID: 3, SelectedIndex: 2},
			},
		},
		{
			name: "缺少一题",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 2, SelectedIndex: 1},
			},
			wantErr: true,
		},
		{
			name: "多出一题",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 2, SelectedIndex: 1},
				{WordID: 3, SelectedIndex: 2},
				{WordID: 4, SelectedIndex: 3},
			},
			wantErr: true,
		},
		{
			name: "不属于该测试的题目",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 2, SelectedIndex: 1},
				{WordID: 99, SelectedIndex: 2},
			},
			wantErr: true,
		},
		{
			name: "重复题目",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 1, SelectedIndex: 1},
				{WordID: 2, SelectedIndex: 2},
			},
			wantErr: true,
		},
		{
			name: "下标越界",
			answers: []Answer{
				{WordID: 1, SelectedIndex: 0},
				{WordID: 2, SelectedIndex: 1},
				{WordID: 3, SelectedIndex: 4},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAnswerCoverage(test.answers, expected)
			if test.wantErr {
				if !errors.Is(err, ErrInvalidAnswers) {
					t.Fatalf("error = %v, want ErrInvalidAnswers", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEstimateVocabulary(t *testing.T) {
	tests := []struct {
		name          string
		correctCount  int
		questionCount int
		totalWords    int
		want          int
	}{
		{name: "全错", correctCount: 0, questionCount: 5, totalWords: 10, want: 0},
		{name: "全对", correctCount: 5, questionCount: 5, totalWords: 10, want: 10},
		{name: "三分之一向下取整", correctCount: 1, questionCount: 3, totalWords: 10, want: 3},
		{name: "三分之二向上取整", correctCount: 2, questionCount: 3, totalWords: 10, want: 7},
		{name: "二分之一四舍五入", correctCount: 1, questionCount: 2, totalWords: 5, want: 3},
		{name: "题目数为零", correctCount: 1, questionCount: 0, totalWords: 10, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EstimateVocabulary(test.correctCount, test.questionCount, test.totalWords)
			if got != test.want {
				t.Errorf("EstimateVocabulary(%d, %d, %d) = %d, want %d",
					test.correctCount, test.questionCount, test.totalWords, got, test.want)
			}
		})
	}
}
