package services

import (
	"context"
	"fmt"
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
	CreateQuestions(ctx context.Context, userID string, req dtos.CreateQuestionsRequestDTO) error
}

type questionService struct {
	txnManager         *utils.TransactionManager
	questionRepo       repository.QuestionRepository
	questionAnswerRepo repository.UserQuestionAnswerRepository
	optionMasterRepo   repository.OptionMasterRepository
}

func NewQuestionService(
	txnManager *utils.TransactionManager,
	questionRepo repository.QuestionRepository,
	questionAnswerRepo repository.UserQuestionAnswerRepository,
	optionMasterRepo repository.OptionMasterRepository,
) QuestionService {
	return &questionService{
		txnManager:         txnManager,
		questionRepo:       questionRepo,
		questionAnswerRepo: questionAnswerRepo,
		optionMasterRepo:   optionMasterRepo,
	}
}

func (s *questionService) SubmitQuestion(ctx context.Context, req dtos.QuestionSubmitRequest, userID string) (*dtos.QuestionResponse, error) {
	now := time.Now()
	question := &entities.QuestionMaster{
		QuestionText: req.QuestionText,
		QuesPoint:    0, // Default value when not provided
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

func (s *questionService) CreateQuestions(ctx context.Context, userID string, req dtos.CreateQuestionsRequestDTO) error {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	for _, qDTO := range req.Questions {
		var questionID int
		isActive := true
		if qDTO.IsActive != nil {
			isActive = *qDTO.IsActive
		}

		if qDTO.ID != nil && *qDTO.ID > 0 {
			q, err := s.questionRepo.FindByIDTx(ctx, tx, *qDTO.ID)
			if err != nil {
				s.txnManager.AbortTxn(tx)
				return fmt.Errorf("failed to find question %d: %w", *qDTO.ID, err)
			}
			if q == nil {
				s.txnManager.AbortTxn(tx)
				return fmt.Errorf("question with id %d not found", *qDTO.ID)
			}
			q.QuestionText = qDTO.QuestionText
			q.QuesPoint = qDTO.QuesPoint
			q.LanguageID = qDTO.LanguageID
			q.IsActive = isActive

			now := time.Now()
			q.LastModifiedOn = &now
			q.LastModifiedBy = &userID

			if err := s.questionRepo.Update(ctx, tx, q); err != nil {
				s.txnManager.AbortTxn(tx)
				return fmt.Errorf("failed to update question %d: %w", *qDTO.ID, err)
			}
			questionID = q.ID
		} else {
			q := &entities.QuestionMaster{
				QuestionText: qDTO.QuestionText,
				QuesPoint:    qDTO.QuesPoint,
				LanguageID:   qDTO.LanguageID,
				IsActive:     isActive,
				IsDeleted:    false,
				CreatedOn:    time.Now(),
				CreatedBy:    userID,
			}
			if err := s.questionRepo.Create(ctx, tx, q); err != nil {
				s.txnManager.AbortTxn(tx)
				return fmt.Errorf("failed to create question: %w", err)
			}
			questionID = q.ID
		}

		for _, oDTO := range qDTO.Options {
			optActive := true
			if oDTO.IsActive != nil {
				optActive = *oDTO.IsActive
			}
			if oDTO.ID != nil && *oDTO.ID > 0 {
				opt, err := s.optionMasterRepo.FindByID(ctx, tx, *oDTO.ID)
				if err != nil {
					s.txnManager.AbortTxn(tx)
					return fmt.Errorf("failed to find option %d: %w", *oDTO.ID, err)
				}
				if opt == nil {
					s.txnManager.AbortTxn(tx)
					return fmt.Errorf("option with id %d not found", *oDTO.ID)
				}
				opt.OptionText = oDTO.OptionText
				opt.DisplayOrder = oDTO.DisplayOrder
				opt.IsActive = optActive

				now := time.Now()
				opt.LastModifiedOn = &now
				opt.LastModifiedBy = &userID

				if err := s.optionMasterRepo.Update(ctx, tx, opt); err != nil {
					s.txnManager.AbortTxn(tx)
					return fmt.Errorf("failed to update option %d: %w", *oDTO.ID, err)
				}
			} else {
				opt := &entities.OptionMaster{
					QuestionMasterID: questionID,
					OptionText:       oDTO.OptionText,
					DisplayOrder:     oDTO.DisplayOrder,
					IsActive:         optActive,
					IsDeleted:        false,
					CreatedOn:        time.Now(),
					CreatedBy:        userID,
				}
				if err := s.optionMasterRepo.Create(ctx, tx, opt); err != nil {
					s.txnManager.AbortTxn(tx)
					return fmt.Errorf("failed to create option: %w", err)
				}
			}
		}
	}

	s.txnManager.CommitTxn(tx)
	return nil
}
