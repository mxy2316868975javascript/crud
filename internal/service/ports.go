package service

import (
	"context"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// WordRepository 是词汇数据访问接口，由 postgres.Store 实现。
type WordRepository interface {
	ListWords(ctx context.Context, level *int, limit, offset int) ([]domain.Word, int, error)
	GetWord(ctx context.Context, id int64) (domain.Word, error)
	CreateWord(ctx context.Context, word domain.Word) (domain.Word, error)
	PutWord(ctx context.Context, id int64, word domain.Word) (domain.Word, error)
	PatchWord(ctx context.Context, id int64, patch domain.WordPatch) (domain.Word, error)
	DeleteWord(ctx context.Context, id int64) error
}

// TestRepository 是测试数据访问接口，由 postgres.Store 实现。
type TestRepository interface {
	CreateTest(ctx context.Context, count int) (*domain.Test, error)
	GetTest(ctx context.Context, id int64) (*domain.Test, error)
	SubmitAnswers(ctx context.Context, id int64, answers []domain.Answer) (*domain.Test, error)
}
