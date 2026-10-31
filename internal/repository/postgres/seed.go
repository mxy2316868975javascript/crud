package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// SeedWords 批量写入或更新词汇，返回处理的词数。
func (store *Store) SeedWords(ctx context.Context, words []domain.Word) (int, error) {
	batch := &pgx.Batch{}
	flush := func() error {
		if batch.Len() == 0 {
			return nil
		}
		results := store.database.SendBatch(ctx, batch)
		for index := 0; index < batch.Len(); index++ {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return err
			}
		}
		if err := results.Close(); err != nil {
			return err
		}
		batch = &pgx.Batch{}
		return nil
	}

	for _, word := range words {
		batch.Queue(
			`INSERT INTO words (word, meaning, level) VALUES ($1, $2, $3)
			 ON CONFLICT (word) DO UPDATE
			 SET meaning = EXCLUDED.meaning, level = EXCLUDED.level`,
			word.Word, word.Meaning, word.Level,
		)
		if batch.Len() >= 1000 {
			if err := flush(); err != nil {
				return 0, err
			}
		}
	}
	if err := flush(); err != nil {
		return 0, err
	}
	return len(words), nil
}
