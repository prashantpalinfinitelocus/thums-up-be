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

type ThunderSeatService interface {
	SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string) (*dtos.ThunderSeatResponse, error)
	GetUserSubmissions(ctx context.Context, userID string) ([]dtos.ThunderSeatResponse, error)
	GetCurrentWeek(ctx context.Context) (*dtos.CurrentWeekResponse, error)
}

type thunderSeatService struct {
	txnManager      *utils.TransactionManager
	thunderSeatRepo repository.ThunderSeatRepository
	questionRepo    repository.QuestionRepository
}

func NewThunderSeatService(
	txnManager *utils.TransactionManager,
	thunderSeatRepo repository.ThunderSeatRepository,
	questionRepo repository.QuestionRepository,
) ThunderSeatService {
	return &thunderSeatService{
		txnManager:      txnManager,
		thunderSeatRepo: thunderSeatRepo,
		questionRepo:    questionRepo,
	}
}

func (s *thunderSeatService) SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string) (*dtos.ThunderSeatResponse, error) {
	question, err := s.questionRepo.FindByID(ctx, s.txnManager.GetDB(), req.QuestionID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to verify question", err)
	}
	if question == nil {
		return nil, errors.NewNotFoundError("Question not found", nil)
	}

	existing, _ := s.thunderSeatRepo.CheckUserSubmission(ctx, s.txnManager.GetDB(), userID, req.QuestionID)
	if existing != nil {
		return nil, errors.NewBadRequestError("You have already submitted an answer for this question", nil)
	}

	now := time.Now()
	thunderSeat := &entities.ThunderSeat{
		UserID:     userID,
		QuestionID: req.QuestionID,
		WeekNumber: req.WeekNumber,
		Answer:     req.Answer,
		CreatedBy:  userID,
		CreatedOn:  now,
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.thunderSeatRepo.Create(ctx, tx, thunderSeat)
	})
	if err != nil {
		log.WithError(err).Error("Failed to submit thunder seat answer")
		return nil, errors.NewInternalServerError("Failed to submit answer", err)
	}

	return &dtos.ThunderSeatResponse{
		ID:         thunderSeat.ID,
		UserID:     thunderSeat.UserID,
		QuestionID: thunderSeat.QuestionID,
		WeekNumber: thunderSeat.WeekNumber,
		Answer:     thunderSeat.Answer,
		CreatedOn:  thunderSeat.CreatedOn.Format(time.RFC3339),
	}, nil
}

func (s *thunderSeatService) GetUserSubmissions(ctx context.Context, userID string) ([]dtos.ThunderSeatResponse, error) {
	submissions, err := s.thunderSeatRepo.FindByUserID(ctx, s.txnManager.GetDB(), userID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get user submissions", err)
	}

	responses := make([]dtos.ThunderSeatResponse, len(submissions))
	for i, sub := range submissions {
		responses[i] = dtos.ThunderSeatResponse{
			ID:         sub.ID,
			UserID:     sub.UserID,
			QuestionID: sub.QuestionID,
			WeekNumber: sub.WeekNumber,
			Answer:     sub.Answer,
			CreatedOn:  sub.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *thunderSeatService) GetCurrentWeek(ctx context.Context) (*dtos.CurrentWeekResponse, error) {
	now := time.Now()

	year, week := now.ISOWeek()
	weekNumber := (year * 100) + week

	startOfWeek := now.AddDate(0, 0, -int(now.Weekday())+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	endOfWeek = time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, endOfWeek.Location())

	return &dtos.CurrentWeekResponse{
		WeekNumber: weekNumber,
		StartDate:  startOfWeek.Format("2006-01-02"),
		EndDate:    endOfWeek.Format("2006-01-02"),
	}, nil
}
