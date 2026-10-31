package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS words (
		id BIGSERIAL PRIMARY KEY,
		word TEXT NOT NULL UNIQUE,
		meaning TEXT NOT NULL,
		level SMALLINT NOT NULL CHECK (level BETWEEN 1 AND 5)
	)`,
	`CREATE INDEX IF NOT EXISTS words_level_idx ON words(level)`,
	`CREATE TABLE IF NOT EXISTS attempts (
		id BIGSERIAL PRIMARY KEY,
		question_count INT NOT NULL,
		total_words_at_start INT NOT NULL,
		correct_count INT,
		estimated_vocabulary INT,
		completed_at TIMESTAMPTZ
	)`,
	`CREATE TABLE IF NOT EXISTS attempt_questions (
		attempt_id BIGINT NOT NULL REFERENCES attempts(id) ON DELETE CASCADE,
		word_id BIGINT NOT NULL REFERENCES words(id) ON DELETE RESTRICT,
		word TEXT NOT NULL,
		options JSONB NOT NULL,
		correct_index SMALLINT NOT NULL CHECK (correct_index BETWEEN 0 AND 3),
		selected_index SMALLINT CHECK (selected_index BETWEEN 0 AND 3),
		PRIMARY KEY (attempt_id, word_id)
	)`,
	`COMMENT ON TABLE words IS 'CET-4 词库'`,
	`COMMENT ON COLUMN words.id IS '主键'`,
	`COMMENT ON COLUMN words.word IS '单词'`,
	`COMMENT ON COLUMN words.meaning IS '中文释义'`,
	`COMMENT ON COLUMN words.level IS '难度等级：1 最常用，5 最生僻'`,
	`COMMENT ON TABLE attempts IS '词汇量测试记录'`,
	`COMMENT ON COLUMN attempts.id IS '测试 ID'`,
	`COMMENT ON COLUMN attempts.question_count IS '题目数量'`,
	`COMMENT ON COLUMN attempts.total_words_at_start IS '创建测试时的词库总量快照'`,
	`COMMENT ON COLUMN attempts.correct_count IS '答对题数，提交后写入'`,
	`COMMENT ON COLUMN attempts.estimated_vocabulary IS '估算词汇量，提交后写入'`,
	`COMMENT ON COLUMN attempts.completed_at IS '提交时间，NULL 表示未完成'`,
	`COMMENT ON TABLE attempt_questions IS '测试题目与作答'`,
	`COMMENT ON COLUMN attempt_questions.attempt_id IS '所属测试 ID'`,
	`COMMENT ON COLUMN attempt_questions.word_id IS '词汇 ID'`,
	`COMMENT ON COLUMN attempt_questions.word IS '题目单词快照'`,
	`COMMENT ON COLUMN attempt_questions.options IS '四个选项的 JSONB 数组'`,
	`COMMENT ON COLUMN attempt_questions.correct_index IS '正确选项下标，0-3'`,
	`COMMENT ON COLUMN attempt_questions.selected_index IS '用户选择的下标，0-3；NULL 表示未作答'`,
}

// EnsureSchema 创建表、索引和字段注释，可重复执行。
func EnsureSchema(ctx context.Context, database *pgxpool.Pool) error {
	for _, statement := range schemaStatements {
		if _, err := database.Exec(ctx, statement); err != nil {
			return fmt.Errorf("schema: %w", err)
		}
	}
	return nil
}
