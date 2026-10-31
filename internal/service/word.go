package service

import (
	"context"
	"strings"

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
	normalizedWord, err := normalizeWord(word)
	if err != nil {
		return domain.Word{}, err
	}
	return service.repo.CreateWord(ctx, normalizedWord)
}

// Put 校验并全量替换词汇。
func (service *WordService) Put(ctx context.Context, id int64, word domain.Word) (domain.Word, error) {
	normalizedWord, err := normalizeWord(word)
	if err != nil {
		return domain.Word{}, err
	}
	return service.repo.PutWord(ctx, id, normalizedWord)
}

// Patch 校验并局部更新词汇。
func (service *WordService) Patch(ctx context.Context, id int64, patch domain.WordPatch) (domain.Word, error) {
	if patch.Word == nil && patch.Meaning == nil && patch.Level == nil {
		return domain.Word{}, &domain.ValidationError{Message: "at least one field is required"}
	}
	if patch.Word != nil {
		value := strings.TrimSpace(*patch.Word)
		if value == "" {
			return domain.Word{}, &domain.ValidationError{Message: "word cannot be empty"}
		}
		patch.Word = &value
	}
	if patch.Meaning != nil {
		value := strings.TrimSpace(*patch.Meaning)
		if value == "" {
			return domain.Word{}, &domain.ValidationError{Message: "meaning cannot be empty"}
		}
		patch.Meaning = &value
	}
	if patch.Level != nil && (*patch.Level < 1 || *patch.Level > 5) {
		return domain.Word{}, &domain.ValidationError{Message: "level must be between 1 and 5"}
	}
	return service.repo.PatchWord(ctx, id, patch)
}

// Delete 删除词汇。
func (service *WordService) Delete(ctx context.Context, id int64) error {
	return service.repo.DeleteWord(ctx, id)
}

// normalizeWord 去掉首尾空格，并校验必填字段和 level 范围。
func normalizeWord(word domain.Word) (domain.Word, error) {
	word.Word = strings.TrimSpace(word.Word)
	word.Meaning = strings.TrimSpace(word.Meaning)
	switch {
	case word.Word == "":
		return domain.Word{}, &domain.ValidationError{Message: "word is required"}
	case word.Meaning == "":
		return domain.Word{}, &domain.ValidationError{Message: "meaning is required"}
	case word.Level < 1 || word.Level > 5:
		return domain.Word{}, &domain.ValidationError{Message: "level must be between 1 and 5"}
	}
	return word, nil
}
