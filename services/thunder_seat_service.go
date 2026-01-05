package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type ThunderSeatService interface {
	SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string, mediaFile *multipart.FileHeader) (*dtos.ThunderSeatResponse, error)
	GetUserSubmissions(ctx context.Context, userID string) ([]dtos.ThunderSeatResponse, error)
	GetCurrentWeek(ctx context.Context) (*dtos.CurrentWeekResponse, error)
}

type thunderSeatService struct {
	txnManager      *utils.TransactionManager
	thunderSeatRepo repository.ThunderSeatRepository
	contestWeekRepo repository.ContestWeekRepository
	userRepo        repository.UserRepository
	gcsService      utils.GCSService
}

func NewThunderSeatService(
	txnManager *utils.TransactionManager,
	thunderSeatRepo repository.ThunderSeatRepository,
	contestWeekRepo repository.ContestWeekRepository,
	userRepo repository.UserRepository,
	gcsService utils.GCSService,
) ThunderSeatService {
	return &thunderSeatService{
		txnManager:      txnManager,
		thunderSeatRepo: thunderSeatRepo,
		contestWeekRepo: contestWeekRepo,
		userRepo:        userRepo,
		gcsService:      gcsService,
	}
}

func (s *thunderSeatService) SubmitAnswer(ctx context.Context, req dtos.ThunderSeatSubmitRequest, userID string, mediaFile *multipart.FileHeader) (*dtos.ThunderSeatResponse, error) {
	activeWeek, err := s.contestWeekRepo.FindActiveWeek(ctx, s.txnManager.GetDB())
	if err != nil {
		log.WithError(err).Error("Failed to get active contest week from database")
		return nil, errors.NewInternalServerError("Failed to get active contest week", err)
	}
	if activeWeek == nil {
		log.Warn("No active contest week found when attempting to submit answer")
		return nil, errors.NewBadRequestError("No active contest week found. Please check if a contest week is currently active.", nil)
	}

	now := time.Now()
	// Include the full end date by checking if now is after end of day (23:59:59.999)
	endOfDay := time.Date(activeWeek.EndDate.Year(), activeWeek.EndDate.Month(), activeWeek.EndDate.Day(), 23, 59, 59, 999999999, activeWeek.EndDate.Location())

	if now.Before(activeWeek.StartDate) {
		log.WithFields(log.Fields{
			"now":         now,
			"start_date":  activeWeek.StartDate,
			"end_date":    activeWeek.EndDate,
			"week_number": activeWeek.WeekNumber,
		}).Warn("Submission attempted before contest week start date")
		return nil, errors.NewBadRequestError(fmt.Sprintf("Submissions are not allowed before the contest week starts. Contest week %d starts on %s", activeWeek.WeekNumber, activeWeek.StartDate.Format("2006-01-02 15:04:05")), nil)
	}

	if now.After(endOfDay) {
		log.WithFields(log.Fields{
			"now":         now,
			"start_date":  activeWeek.StartDate,
			"end_date":    activeWeek.EndDate,
			"end_of_day":  endOfDay,
			"week_number": activeWeek.WeekNumber,
		}).Warn("Submission attempted after contest week end date")
		return nil, errors.NewBadRequestError(fmt.Sprintf("Submissions are not allowed after the contest week ends. Contest week %d ended on %s", activeWeek.WeekNumber, activeWeek.EndDate.Format("2006-01-02")), nil)
	}

	thunderSeat := &entities.ThunderSeat{
		UserID:     userID,
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
		if err := s.thunderSeatRepo.Create(ctx, tx, thunderSeat); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user_id":     userID,
				"week_number": activeWeek.WeekNumber,
			}).Error("Failed to create thunder seat record in database")
			return err
		}

		if req.SharingPlatform != nil || req.PlatformUserName != nil {
			userUUID, parseErr := uuid.Parse(userID)
			if parseErr != nil {
				log.WithError(parseErr).WithField("user_id", userID).Error("Failed to parse user ID as UUID")
				return parseErr
			}

			user, findErr := s.userRepo.FindById(ctx, tx, userUUID)
			if findErr != nil {
				log.WithError(findErr).WithField("user_id", userID).Error("Failed to find user for updating sharing platform info")
				return findErr
			}

			updateFields := make(map[string]interface{})
			if req.SharingPlatform != nil {
				updateFields["sharing_platform"] = *req.SharingPlatform
			}
			if req.PlatformUserName != nil {
				updateFields["platform_user_name"] = *req.PlatformUserName
			}

			if len(updateFields) > 0 {
				if updateErr := tx.WithContext(ctx).Model(user).Updates(updateFields).Error; updateErr != nil {
					log.WithError(updateErr).WithFields(log.Fields{
						"user_id":       userID,
						"update_fields": updateFields,
					}).Error("Failed to update user sharing platform information")
					return updateErr
				}
			}
		}

		return nil
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id":     userID,
			"week_number": activeWeek.WeekNumber,
			"has_media":   mediaFile != nil,
		}).Error("Failed to submit thunder seat answer in transaction")

		if thunderSeat.MediaURL != nil {
			if deleteErr := s.gcsService.DeleteFile(ctx, *thunderSeat.MediaURL); deleteErr != nil {
				log.WithError(deleteErr).WithField("media_url", *thunderSeat.MediaURL).Error("Failed to cleanup uploaded file after database error")
			}
		}

		errStr := err.Error()
		if contains(errStr, "duplicate") || contains(errStr, "unique constraint") || contains(errStr, "UNIQUE constraint") {
			return nil, errors.NewBadRequestError("You have already submitted an answer for this contest week", err)
		}

		return nil, errors.NewInternalServerError("Failed to submit answer. Please try again later.", err)
	}

	return &dtos.ThunderSeatResponse{
		ID:         thunderSeat.ID,
		UserID:     thunderSeat.UserID,
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
		var avatarURL *string
		var avatarName *string
		if sub.User.Avatar != nil {
			// Reconstruct full path: avatars/{createdBy}/{imageKey}
			fullPath := fmt.Sprintf("avatars/%s/%s", sub.User.Avatar.CreatedBy, sub.User.Avatar.ImageKey)
			url := s.gcsService.GetPublicURL(fullPath)
			avatarURL = &url
			avatarName = &sub.User.Avatar.Name
		}

		responses[i] = dtos.ThunderSeatResponse{
			ID:         sub.ID,
			UserID:     sub.UserID,
			WeekNumber: sub.WeekNumber,
			Answer:     sub.Answer,
			MediaURL:   sub.MediaURL,
			MediaType:  sub.MediaType,
			CreatedOn:  sub.CreatedOn.Format(time.RFC3339),
			Name:       sub.User.Name,
			Email:      sub.User.Email,
			AvatarURL:  avatarURL,
			AvatarName: avatarName,
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
