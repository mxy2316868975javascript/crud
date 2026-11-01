package service

import (
	"context"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// WordService 处理词汇的查询、创建、更新和删除。
type WordService struct {
	repo WordRepository
}

// NewWordService 创建 WordService。
func NewWordService(repo WordRepository) *WordService {
	return &WordService{repo: repo}
}

// List 返回分页后的词汇和总数。
func (service *WordService) List(ctx context.Context, level *int, limit, offset int) ([]domain.Word, int, error) {
	return service.repo.ListWords(ctx, level, limit, offset)
}

// Get 返回单个词汇。
func (service *WordService) Get(ctx context.Context, id int64) (domain.Word, error) {
	return service.repo.GetWord(ctx, id)
}

// Create 校验并创建词汇。
func (service *WordService) Create(ctx context.Context, word domain.Word) (domain.Word, error) {
	normalizedWord, err := domain.NewWord(word.Word, word.Meaning, word.Level)
	if err != nil {
		return domain.Word{}, err
	}
	return service.repo.CreateWord(ctx, normalizedWord)
}

// Put 校验并全量替换词汇。
func (service *WordService) Put(ctx context.Context, id int64, word domain.Word) (domain.Word, error) {
	normalizedWord, err := domain.NewWord(word.Word, word.Meaning, word.Level)
	if err != nil {
		return domain.Word{}, err
	}
	return service.repo.PutWord(ctx, id, normalizedWord)
}

// Patch 校验并局部更新词汇。
func (service *WordService) Patch(ctx context.Context, id int64, patch domain.WordPatch) (domain.Word, error) {
	normalizedPatch, err := patch.Normalize()
	if err != nil {
		return domain.Word{}, err
	}
	return service.repo.PatchWord(ctx, id, normalizedPatch)
}

// Delete 删除词汇。
func (service *WordService) Delete(ctx context.Context, id int64) error {
	return service.repo.DeleteWord(ctx, id)
}
