package domain

import "strings"

// Word 表示词库中的一个单词。
type Word struct {
	ID      int64  `json:"id" example:"1"`             // 词汇 ID
	Word    string `json:"word" example:"abandon"`     // 单词
	Meaning string `json:"meaning" example:"v. 放弃；抛弃"` // 中文释义
	Level   int    `json:"level" example:"1"`          // 难度等级：1 最常用，5 最生僻
}

// NewWord 去掉首尾空格并校验一个新词汇。
func NewWord(text, meaning string, level int) (Word, error) {
	word := Word{
		Word:    strings.TrimSpace(text),
		Meaning: strings.TrimSpace(meaning),
		Level:   level,
	}
	if err := word.Validate(); err != nil {
		return Word{}, err
	}
	return word, nil
}

// Validate 校验词汇的必填字段和 level 范围。
func (word Word) Validate() error {
	switch {
	case word.Word == "":
		return &ValidationError{Message: "word is required"}
	case word.Meaning == "":
		return &ValidationError{Message: "meaning is required"}
	case word.Level < 1 || word.Level > 5:
		return &ValidationError{Message: "level must be between 1 and 5"}
	}
	return nil
}

// WordPatch 表示词汇的部分更新字段；nil 表示不修改。
type WordPatch struct {
	Word    *string `json:"word" example:"abandon"`     // 新单词
	Meaning *string `json:"meaning" example:"v. 放弃；抛弃"` // 新释义
	Level   *int    `json:"level" example:"1"`          // 新难度等级：1-5
}

// Normalize 去掉首尾空格并校验局部更新字段，返回规范化后的副本。
func (patch WordPatch) Normalize() (WordPatch, error) {
	if patch.Word == nil && patch.Meaning == nil && patch.Level == nil {
		return WordPatch{}, &ValidationError{Message: "at least one field is required"}
	}
	if patch.Word != nil {
		value := strings.TrimSpace(*patch.Word)
		if value == "" {
			return WordPatch{}, &ValidationError{Message: "word cannot be empty"}
		}
		patch.Word = &value
	}
	if patch.Meaning != nil {
		value := strings.TrimSpace(*patch.Meaning)
		if value == "" {
			return WordPatch{}, &ValidationError{Message: "meaning cannot be empty"}
		}
		patch.Meaning = &value
	}
	if patch.Level != nil && (*patch.Level < 1 || *patch.Level > 5) {
		return WordPatch{}, &ValidationError{Message: "level must be between 1 and 5"}
	}
	return patch, nil
}
