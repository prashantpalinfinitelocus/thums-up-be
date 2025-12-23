package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"mime/multipart"
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
	SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string, mediaFile *multipart.FileHeader) (*dtos.ThunderSeatResponse, error)
	GetUserSubmissions(ctx context.Context, userID string) ([]dtos.ThunderSeatResponse, error)
	GetCurrentWeek(ctx context.Context) (*dtos.CurrentWeekResponse, error)
}

type thunderSeatService struct {
	txnManager      *utils.TransactionManager
	thunderSeatRepo repository.ThunderSeatRepository
	questionRepo    repository.QuestionRepository
	contestWeekRepo repository.ContestWeekRepository
	gcsService      utils.GCSService
}

func NewThunderSeatService(
	txnManager *utils.TransactionManager,
	thunderSeatRepo repository.ThunderSeatRepository,
	questionRepo repository.QuestionRepository,
	contestWeekRepo repository.ContestWeekRepository,
	gcsService utils.GCSService,
) ThunderSeatService {
	return &thunderSeatService{
		txnManager:      txnManager,
		thunderSeatRepo: thunderSeatRepo,
		questionRepo:    questionRepo,
		contestWeekRepo: contestWeekRepo,
		gcsService:      gcsService,
	}
}

func (s *thunderSeatService) SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string, mediaFile *multipart.FileHeader) (*dtos.ThunderSeatResponse, error) {
	activeWeek, err := s.contestWeekRepo.FindActiveWeek(ctx, s.txnManager.GetDB())
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get active contest week", err)
	}
	if activeWeek == nil {
		return nil, errors.NewBadRequestError("No active contest week found", nil)
	}

	now := time.Now()
	if now.Before(activeWeek.StartDate) || now.After(activeWeek.EndDate) {
		return nil, errors.NewBadRequestError("Submissions are not allowed outside the active contest week period", nil)
	}

	question, err := s.questionRepo.FindByID(ctx, s.txnManager.GetDB(), req.QuestionID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to verify question", err)
	}
	if question == nil {
		return nil, errors.NewNotFoundError("Question not found", nil)
	}

	existing, err := s.thunderSeatRepo.CheckUserSubmission(ctx, s.txnManager.GetDB(), userID, req.QuestionID)
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("Failed to check user submission")
		return nil, errors.NewInternalServerError("Failed to check submission", err)
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("You have already submitted an answer for this question", nil)
	}

	thunderSeat := &entities.ThunderSeat{
		UserID:     userID,
		QuestionID: req.QuestionID,
		WeekNumber: activeWeek.WeekNumber,
		Answer:     req.Answer,
		CreatedBy:  userID,
		CreatedOn:  now,
	}

	// Upload media file to GCS if provided
	if mediaFile != nil {
		folderPath := fmt.Sprintf("thunder-seat/%s/week-%d", userID, activeWeek.WeekNumber)
		mediaURL, mediaKey, err := s.gcsService.UploadFile(ctx, mediaFile, folderPath)
		if err != nil {
			log.WithError(err).Error("Failed to upload media file to GCS")
			return nil, errors.NewInternalServerError("Failed to upload media file", err)
		}

		mediaType := utils.GetMediaType(mediaFile)
		thunderSeat.MediaURL = &mediaURL
		thunderSeat.MediaKey = &mediaKey
		thunderSeat.MediaType = &mediaType
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.thunderSeatRepo.Create(ctx, tx, thunderSeat)
	})
	if err != nil {
		log.WithError(err).Error("Failed to submit thunder seat answer")

		// Cleanup uploaded file if database operation fails
		if thunderSeat.MediaURL != nil {
			if deleteErr := s.gcsService.DeleteFile(ctx, *thunderSeat.MediaURL); deleteErr != nil {
				log.WithError(deleteErr).Error("Failed to cleanup uploaded file after database error")
			}
		}

		return nil, errors.NewInternalServerError("Failed to submit answer", err)
	}

	return &dtos.ThunderSeatResponse{
		ID:         thunderSeat.ID,
		UserID:     thunderSeat.UserID,
		QuestionID: thunderSeat.QuestionID,
		WeekNumber: thunderSeat.WeekNumber,
		Answer:     thunderSeat.Answer,
		MediaURL:   thunderSeat.MediaURL,
		MediaType:  thunderSeat.MediaType,
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
			MediaURL:   sub.MediaURL,
			MediaType:  sub.MediaType,
			CreatedOn:  sub.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *thunderSeatService) GetCurrentWeek(ctx context.Context) (*dtos.CurrentWeekResponse, error) {
	activeWeek, err := s.contestWeekRepo.FindActiveWeek(ctx, s.txnManager.GetDB())
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get active contest week", err)
	}
	if activeWeek == nil {
		return nil, errors.NewNotFoundError("No active contest week found", nil)
	}

	return &dtos.CurrentWeekResponse{
		WeekNumber:  activeWeek.WeekNumber,
		StartDate:   activeWeek.StartDate.Format("2006-01-02"),
		EndDate:     activeWeek.EndDate.Format("2006-01-02"),
		WinnerCount: activeWeek.WinnerCount,
		IsActive:    activeWeek.IsActive,
	}, nil
}
