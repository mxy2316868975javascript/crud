package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// Store 是 PostgreSQL 仓储实现。
type Store struct {
	database *pgxpool.Pool
}

// New 用连接池创建 Store。
func New(database *pgxpool.Pool) *Store {
	return &Store{database: database}
}

// Ping 检查数据库连接。
func (store *Store) Ping(ctx context.Context) error {
	return store.database.Ping(ctx)
}

// translateWordWriteError 把 PostgreSQL 唯一约束和外键错误转换成领域错误。
func translateWordWriteError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return domain.ErrWordExists
		case "23503":
			return domain.ErrWordInUse
		}
	}
	return err
}

// CreateWord 插入一个词汇并返回带 id 的结果。
func (store *Store) CreateWord(ctx context.Context, word domain.Word) (domain.Word, error) {
	err := store.database.QueryRow(ctx,
		`INSERT INTO words (word, meaning, level) VALUES ($1, $2, $3) RETURNING id`,
		word.Word, word.Meaning, word.Level,
	).Scan(&word.ID)
	if err != nil {
		return domain.Word{}, translateWordWriteError(err)
	}
	return word, nil
}

// ListWords 分页查询词汇，同时返回符合条件的总数。
func (store *Store) ListWords(ctx context.Context, level *int, limit, offset int) ([]domain.Word, int, error) {
	var levelArgument any
	if level != nil {
		levelArgument = *level
	}

	rows, err := store.database.Query(ctx,
		`SELECT id, word, meaning, level FROM words
		 WHERE ($1::int IS NULL OR level = $1)
		 ORDER BY id LIMIT $2 OFFSET $3`,
		levelArgument, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	words := make([]domain.Word, 0, limit)
	for rows.Next() {
		var word domain.Word
		if err := rows.Scan(&word.ID, &word.Word, &word.Meaning, &word.Level); err != nil {
			return nil, 0, err
		}
		words = append(words, word)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := store.database.QueryRow(ctx,
		`SELECT count(*) FROM words WHERE ($1::int IS NULL OR level = $1)`,
		levelArgument,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	return words, total, nil
}

// GetWord 按 id 查询一个词汇。
func (store *Store) GetWord(ctx context.Context, id int64) (domain.Word, error) {
	var word domain.Word
	err := store.database.QueryRow(ctx,
		`SELECT id, word, meaning, level FROM words WHERE id = $1`, id,
	).Scan(&word.ID, &word.Word, &word.Meaning, &word.Level)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Word{}, domain.ErrNotFound
	}
	return word, err
}

// PutWord 全量替换一个词汇。
func (store *Store) PutWord(ctx context.Context, id int64, word domain.Word) (domain.Word, error) {
	err := store.database.QueryRow(ctx,
		`UPDATE words SET word = $1, meaning = $2, level = $3
		 WHERE id = $4 RETURNING id, word, meaning, level`,
		word.Word, word.Meaning, word.Level, id,
	).Scan(&word.ID, &word.Word, &word.Meaning, &word.Level)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Word{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Word{}, translateWordWriteError(err)
	}
	return word, nil
}

// PatchWord 局部更新一个词汇。
func (store *Store) PatchWord(ctx context.Context, id int64, patch domain.WordPatch) (domain.Word, error) {
	assignments := make([]string, 0, 3)
	arguments := make([]any, 0, 4)

	if patch.Word != nil {
		arguments = append(arguments, *patch.Word)
		assignments = append(assignments, fmt.Sprintf("word = $%d", len(arguments)))
	}
	if patch.Meaning != nil {
		arguments = append(arguments, *patch.Meaning)
		assignments = append(assignments, fmt.Sprintf("meaning = $%d", len(arguments)))
	}
	if patch.Level != nil {
		arguments = append(arguments, *patch.Level)
		assignments = append(assignments, fmt.Sprintf("level = $%d", len(arguments)))
	}
	if len(assignments) == 0 {
		return domain.Word{}, &domain.ValidationError{Message: "at least one field is required"}
	}

	arguments = append(arguments, id)
	// 列名只来自固定的 word/meaning/level，值全部走占位符，避免 SQL 注入。
	updateStatement := fmt.Sprintf(
		`UPDATE words SET %s WHERE id = $%d RETURNING id, word, meaning, level`,
		strings.Join(assignments, ", "), len(arguments),
	)

	var word domain.Word
	err := store.database.QueryRow(ctx, updateStatement, arguments...).Scan(&word.ID, &word.Word, &word.Meaning, &word.Level)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Word{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Word{}, translateWordWriteError(err)
	}
	return word, nil
}

// DeleteWord 删除一个词汇；被测试引用时返回领域错误。
func (store *Store) DeleteWord(ctx context.Context, id int64) error {
	commandTag, err := store.database.Exec(ctx, `DELETE FROM words WHERE id = $1`, id)
	if err != nil {
		return translateWordWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// wordSample 是抽题时从数据库读出的最小字段。
type wordSample struct {
	id      int64
	word    string
	meaning string
}

// CreateTest 创建一次测试：抽题、生成选项并写入数据库。
func (store *Store) CreateTest(ctx context.Context, count int) (*domain.Test, error) {
	transaction, err := store.database.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	// 第一步：确认词库数量足够。
	var totalWords int
	if err := transaction.QueryRow(ctx, `SELECT count(*) FROM words`).Scan(&totalWords); err != nil {
		return nil, err
	}
	if totalWords < count {
		return nil, domain.ErrNotEnoughWords
	}

	// 第二步：随机抽取题目。
	rows, err := transaction.Query(ctx,
		`SELECT id, word, meaning FROM words ORDER BY random() LIMIT $1`, count,
	)
	if err != nil {
		return nil, err
	}
	samples := make([]wordSample, 0, count)
	for rows.Next() {
		var sample wordSample
		if err := rows.Scan(&sample.id, &sample.word, &sample.meaning); err != nil {
			rows.Close()
			return nil, err
		}
		samples = append(samples, sample)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(samples) < count {
		return nil, domain.ErrNotEnoughWords
	}

	// 第三步：创建一次测试记录。
	var attemptID int64
	if err := transaction.QueryRow(ctx,
		`INSERT INTO attempts (question_count, total_words_at_start) VALUES ($1, $2) RETURNING id`,
		count, totalWords,
	).Scan(&attemptID); err != nil {
		return nil, err
	}

	// 第四步：为每道题生成干扰项、打乱选项并保存。
	questions := make([]domain.Question, 0, count)
	for _, sample := range samples {
		// 简化实现：每道题单独查询一次干扰项，题数不超过 100 时足够快。
		distractorRows, err := transaction.Query(ctx,
			`SELECT meaning FROM (
			   SELECT DISTINCT meaning FROM words
			   WHERE id <> $1 AND meaning <> $2
			 ) d ORDER BY random() LIMIT 3`,
			sample.id, sample.meaning,
		)
		if err != nil {
			return nil, err
		}
		distractors := make([]string, 0, 3)
		for distractorRows.Next() {
			var meaning string
			if err := distractorRows.Scan(&meaning); err != nil {
				distractorRows.Close()
				return nil, err
			}
			distractors = append(distractors, meaning)
		}
		distractorRows.Close()
		if err := distractorRows.Err(); err != nil {
			return nil, err
		}
		if len(distractors) < 3 {
			return nil, domain.ErrNotEnoughWords
		}

		options := append([]string{sample.meaning}, distractors...)
		rand.Shuffle(len(options), func(firstIndex, secondIndex int) {
			options[firstIndex], options[secondIndex] = options[secondIndex], options[firstIndex]
		})
		correctIndex := slices.Index(options, sample.meaning)
		if correctIndex < 0 {
			return nil, errors.New("correct option missing")
		}

		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return nil, err
		}
		if _, err := transaction.Exec(ctx,
			`INSERT INTO attempt_questions (attempt_id, word_id, word, options, correct_index)
			 VALUES ($1, $2, $3, $4::jsonb, $5)`,
			attemptID, sample.id, sample.word, string(optionsJSON), correctIndex,
		); err != nil {
			return nil, err
		}

		questions = append(questions, domain.Question{WordID: sample.id, Word: sample.word, Options: options})
	}

	// 全部成功后再提交，避免留下半场测试。
	if err := transaction.Commit(ctx); err != nil {
		return nil, err
	}
	return &domain.Test{ID: attemptID, QuestionCount: count, Questions: questions}, nil
}

// GetTest 按 id 查询测试和题目。
func (store *Store) GetTest(ctx context.Context, id int64) (*domain.Test, error) {
	test := &domain.Test{}
	var completedAt *time.Time
	err := store.database.QueryRow(ctx,
		`SELECT id, question_count, correct_count, estimated_vocabulary, completed_at
		 FROM attempts WHERE id = $1`, id,
	).Scan(&test.ID, &test.QuestionCount, &test.CorrectCount, &test.EstimatedVocabulary, &completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	test.Completed = completedAt != nil
	test.CompletedAt = completedAt

	rows, err := store.database.Query(ctx,
		`SELECT word_id, word, options FROM attempt_questions
		 WHERE attempt_id = $1 ORDER BY word_id`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var question domain.Question
		var optionsJSON []byte
		if err := rows.Scan(&question.WordID, &question.Word, &optionsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(optionsJSON, &question.Options); err != nil {
			return nil, err
		}
		test.Questions = append(test.Questions, question)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return test, nil
}

// SubmitAnswers 在事务里校验并保存答案，然后计算成绩。
func (store *Store) SubmitAnswers(ctx context.Context, attemptID int64, answers []domain.Answer) (*domain.Test, error) {
	transaction, err := store.database.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	// 第一步：锁定该测试，防止并发重复提交。
	var questionCount, totalWords int
	var completedAt *time.Time
	err = transaction.QueryRow(ctx,
		`SELECT question_count, total_words_at_start, completed_at
		 FROM attempts WHERE id = $1 FOR UPDATE`, attemptID,
	).Scan(&questionCount, &totalWords, &completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if completedAt != nil {
		return nil, domain.ErrTestCompleted
	}

	// 第二步：读出每道题的正确答案下标。
	rows, err := transaction.Query(ctx,
		`SELECT word_id, correct_index FROM attempt_questions WHERE attempt_id = $1`,
		attemptID,
	)
	if err != nil {
		return nil, err
	}
	correctIndexes := make(map[int64]int, questionCount)
	for rows.Next() {
		var wordID int64
		var correctIndex int
		if err := rows.Scan(&wordID, &correctIndex); err != nil {
			rows.Close()
			return nil, err
		}
		correctIndexes[wordID] = correctIndex
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 第三步：校验答案是否覆盖全部题目且没有重复。
	expectedWordIDs := make([]int64, 0, len(correctIndexes))
	for wordID := range correctIndexes {
		expectedWordIDs = append(expectedWordIDs, wordID)
	}
	if err := domain.ValidateAnswerShape(answers); err != nil {
		return nil, err
	}
	if err := domain.ValidateAnswerCoverage(answers, expectedWordIDs); err != nil {
		return nil, err
	}

	correctCount := 0
	for _, answer := range answers {
		correctIndex := correctIndexes[answer.WordID]
		if answer.SelectedIndex == correctIndex {
			correctCount++
		}
		if _, err := transaction.Exec(ctx,
			`UPDATE attempt_questions SET selected_index = $1
			 WHERE attempt_id = $2 AND word_id = $3`,
			answer.SelectedIndex, attemptID, answer.WordID,
		); err != nil {
			return nil, err
		}
	}

	// 第四步：计算正确数和估算词汇量。
	estimatedVocabulary := domain.EstimateVocabulary(correctCount, questionCount, totalWords)

	// 第五步：写入成绩和完成时间。
	var completionTime time.Time
	if err := transaction.QueryRow(ctx,
		`UPDATE attempts
		 SET correct_count = $1, estimated_vocabulary = $2, completed_at = now()
		 WHERE id = $3 RETURNING completed_at`,
		correctCount, estimatedVocabulary, attemptID,
	).Scan(&completionTime); err != nil {
		return nil, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, err
	}
	return &domain.Test{
		ID:                  attemptID,
		QuestionCount:       questionCount,
		Completed:           true,
		CorrectCount:        &correctCount,
		EstimatedVocabulary: &estimatedVocabulary,
		CompletedAt:         &completionTime,
	}, nil
}
