package service

import (
	"context"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// TestService 处理创建测试和提交答案。
type TestService struct {
	repo TestRepository
}

// NewTestService 创建 TestService。
func NewTestService(repo TestRepository) *TestService {
	return &TestService{repo: repo}
}

// Create 创建一场测试并抽题。
func (service *TestService) Create(ctx context.Context, count int) (*domain.Test, error) {
	if err := domain.ValidateQuestionCount(count); err != nil {
		return nil, err
	}
	return service.repo.CreateTest(ctx, count)
}

// Get 返回测试详情。
func (service *TestService) Get(ctx context.Context, id int64) (*domain.Test, error) {
	return service.repo.GetTest(ctx, id)
}

// Submit 校验答案并判分。
func (service *TestService) Submit(ctx context.Context, id int64, answers []domain.Answer) (*domain.Test, error) {
	if err := domain.ValidateAnswerShape(answers); err != nil {
		return nil, err
	}
	return service.repo.SubmitAnswers(ctx, id, answers)
}
