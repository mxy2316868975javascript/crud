package domain

// Word 表示词库中的一个单词。
type Word struct {
	ID      int64  `json:"id" example:"1"`             // 词汇 ID
	Word    string `json:"word" example:"abandon"`     // 单词
	Meaning string `json:"meaning" example:"v. 放弃；抛弃"` // 中文释义
	Level   int    `json:"level" example:"1"`          // 难度等级：1 最常用，5 最生僻
}

// WordPatch 表示词汇的部分更新字段；nil 表示不修改。
type WordPatch struct {
	Word    *string `json:"word" example:"abandon"`     // 新单词
	Meaning *string `json:"meaning" example:"v. 放弃；抛弃"` // 新释义
	Level   *int    `json:"level" example:"1"`          // 新难度等级：1-5
}
