package services

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type QuestionService interface {
	SubmitQuestion(ctx context.Context, req dtos.QuestionSubmitRequest, userID string) (*dtos.QuestionResponse, error)
	GetActiveQuestions(ctx context.Context) ([]dtos.QuestionResponse, error)
	GetQuestionsByLanguage(ctx context.Context, languageID int) ([]dtos.QuestionResponse, error)
}

type questionService struct {
	txnManager   *utils.TransactionManager
	questionRepo repository.QuestionRepository
}

func NewQuestionService(
	txnManager *utils.TransactionManager,
	questionRepo repository.QuestionRepository,
) QuestionService {
	return &questionService{
		txnManager:   txnManager,
		questionRepo: questionRepo,
	}
}

func (s *questionService) SubmitQuestion(ctx context.Context, req dtos.QuestionSubmitRequest, userID string) (*dtos.QuestionResponse, error) {
	now := time.Now()
	question := &entities.QuestionMaster{
		QuestionText: req.QuestionText,
		LanguageID:   req.LanguageID,
		IsActive:     true,
		IsDeleted:    false,
		CreatedBy:    userID,
		CreatedOn:    now,
	}

	err := s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.questionRepo.Create(ctx, tx, question)
	})
	if err != nil {
		log.WithError(err).Error("Failed to create question")
		return nil, errors.NewInternalServerError("Failed to submit question", err)
	}

	return &dtos.QuestionResponse{
		ID:           question.ID,
		QuestionText: question.QuestionText,
		LanguageID:   question.LanguageID,
		IsActive:     question.IsActive,
	}, nil
}

func (s *questionService) GetActiveQuestions(ctx context.Context) ([]dtos.QuestionResponse, error) {
	questions, err := s.questionRepo.FindActiveQuestions(ctx, s.txnManager.GetDB(), 100, 0)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get active questions", err)
	}

	responses := make([]dtos.QuestionResponse, len(questions))
	for i, q := range questions {
		responses[i] = dtos.QuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			LanguageID:   q.LanguageID,
			IsActive:     q.IsActive,
		}
	}

	return responses, nil
}

func (s *questionService) GetQuestionsByLanguage(ctx context.Context, languageID int) ([]dtos.QuestionResponse, error) {
	questions, err := s.questionRepo.FindByLanguageID(ctx, s.txnManager.GetDB(), languageID, 100, 0)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get questions by language", err)
	}

	responses := make([]dtos.QuestionResponse, len(questions))
	for i, q := range questions {
		responses[i] = dtos.QuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			LanguageID:   q.LanguageID,
			IsActive:     q.IsActive,
		}
	}

	return responses, nil
}
