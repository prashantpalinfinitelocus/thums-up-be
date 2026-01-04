package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type WinnerService interface {
	SelectWinners(ctx context.Context, req dtos.SelectWinnersRequest) ([]dtos.WinnerResponse, error)
	GetWinnersByWeek(ctx context.Context, weekNumber int) ([]dtos.WinnerResponse, error)
	GetAllWinners(ctx context.Context, limit, offset int) ([]dtos.WinnerResponse, int64, error)
	SubmitWinnerKYC(ctx context.Context, userID string, req dtos.WinnerKYCRequest) error
	CheckUserWinnerStatus(ctx context.Context, userID string) (*dtos.WinnerStatusResponse, error)
	MarkBannerAsViewed(ctx context.Context, userID string) error
}

type winnerService struct {
	txnManager             *utils.TransactionManager
	winnerRepo             repository.WinnerRepository
	thunderSeatRepo        repository.ThunderSeatRepository
	contestWeekRepo        repository.ContestWeekRepository
	userRepo               repository.UserRepository
	userAadharRepo         repository.UserAadharCardRepository
	userAdditionalInfoRepo repository.UserAdditionalInfoRepository
	gcsService             utils.GCSService
}

func NewWinnerService(
	txnManager *utils.TransactionManager,
	winnerRepo repository.WinnerRepository,
	thunderSeatRepo repository.ThunderSeatRepository,
	contestWeekRepo repository.ContestWeekRepository,
	userRepo repository.UserRepository,
	userAadharRepo repository.UserAadharCardRepository,
	userAdditionalInfoRepo repository.UserAdditionalInfoRepository,
	gcsService utils.GCSService,
) WinnerService {
	return &winnerService{
		txnManager:             txnManager,
		winnerRepo:             winnerRepo,
		thunderSeatRepo:        thunderSeatRepo,
		contestWeekRepo:        contestWeekRepo,
		userRepo:               userRepo,
		userAadharRepo:         userAadharRepo,
		userAdditionalInfoRepo: userAdditionalInfoRepo,
		gcsService:             gcsService,
	}
}

func (s *winnerService) SelectWinners(ctx context.Context, req dtos.SelectWinnersRequest) ([]dtos.WinnerResponse, error) {
	contestWeek, err := s.contestWeekRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), req.WeekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get contest week", err)
	}
	if contestWeek == nil {
		return nil, errors.NewNotFoundError("Contest week not found", nil)
	}

	existingWinners, err := s.winnerRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), req.WeekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get existing winners", err)
	}
	if len(existingWinners) >= contestWeek.WinnerCount {
		return nil, errors.NewBadRequestError("Winners already selected for this week", nil)
	}

	existingWinnerUserIDs, err := s.winnerRepo.GetWinnerUserIDs(ctx, s.txnManager.GetDB(), req.WeekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get existing winner IDs", err)
	}

	remainingSlots := contestWeek.WinnerCount - len(existingWinners)
	randomEntries, err := s.thunderSeatRepo.GetRandomEntriesByWeek(ctx, s.txnManager.GetDB(), req.WeekNumber, remainingSlots, existingWinnerUserIDs)
	if err != nil {
		log.WithError(err).Error("Failed to get random entries")
		return nil, errors.NewInternalServerError("Failed to select random entries", err)
	}

	if len(randomEntries) == 0 {
		return nil, errors.NewNotFoundError("No eligible entries found for winner selection", nil)
	}

	now := time.Now()
	winners := make([]entities.ThunderSeatWinner, len(randomEntries))

	for i, entry := range randomEntries {
		qrData := fmt.Sprintf("winner:%s:week:%d:thunder_seat:%d", entry.UserID, req.WeekNumber, entry.ID)
		qrBytes, err := utils.GenerateQRCode(qrData)
		if err != nil {
			log.WithError(err).Error("Failed to generate QR code")
			return nil, errors.NewInternalServerError("Failed to generate QR code", err)
		}

		qrPath := fmt.Sprintf("winners/week_%d/%s.png", req.WeekNumber, entry.UserID)
		_, qrKey, err := s.gcsService.UploadFileFromBytes(ctx, qrBytes, qrPath, "image/png")
		if err != nil {
			log.WithError(err).Error("Failed to upload QR code")
			return nil, errors.NewInternalServerError("Failed to upload QR code", err)
		}

		winners[i] = entities.ThunderSeatWinner{
			UserID:        entry.UserID,
			ThunderSeatID: entry.ID,
			QRCode:        qrKey,
			WeekNumber:    req.WeekNumber,
			HasViewed:     false,
			CreatedBy:     constants.SYSTEM_USER_ID,
			CreatedOn:     now,
		}
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.Create(&winners).Error
	})
	if err != nil {
		log.WithError(err).Error("Failed to create winners")
		return nil, errors.NewInternalServerError("Failed to save winners", err)
	}

	responses := make([]dtos.WinnerResponse, len(winners))
	for i, winner := range winners {
		qrURL := s.gcsService.GetPublicURL(winner.QRCode)
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
			QRCodeURL:     &qrURL,
			CreatedOn:     winner.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *winnerService) GetWinnersByWeek(ctx context.Context, weekNumber int) ([]dtos.WinnerResponse, error) {
	winners, err := s.winnerRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), weekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get winners", err)
	}

	responses := make([]dtos.WinnerResponse, len(winners))
	for i, winner := range winners {
		var qrURL *string
		if winner.QRCode != "" {
			url := s.gcsService.GetPublicURL(winner.QRCode)
			qrURL = &url
		}
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
			QRCodeURL:     qrURL,
			CreatedOn:     winner.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *winnerService) GetAllWinners(ctx context.Context, limit, offset int) ([]dtos.WinnerResponse, int64, error) {
	winners, total, err := s.winnerRepo.FindAllWithPagination(ctx, s.txnManager.GetDB(), limit, offset)
	if err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to get winners", err)
	}

	responses := make([]dtos.WinnerResponse, len(winners))
	for i, winner := range winners {
		var qrURL *string
		if winner.QRCode != "" {
			url := s.gcsService.GetPublicURL(winner.QRCode)
			qrURL = &url
		}
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
			QRCodeURL:     qrURL,
			CreatedOn:     winner.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, total, nil
}

// SubmitWinnerKYC stores winner's aadhar details and optional friends' details.
// It assumes the caller has already verified authentication; this method will
// verify that the user is actually a winner.
func (s *winnerService) SubmitWinnerKYC(ctx context.Context, userID string, req dtos.WinnerKYCRequest) error {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	// Verify user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return errors.NewBadRequestError("Invalid user ID format", err)
	}

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return errors.NewInternalServerError("Failed to fetch user", err)
	}
	if user == nil {
		s.txnManager.AbortTxn(tx)
		return errors.NewNotFoundError("User not found", nil)
	}

	// Ensure user is a winner (any week)
	winners, err := s.winnerRepo.FindByUserID(ctx, tx, userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return errors.NewInternalServerError("Failed to verify winner", err)
	}
	if len(winners) == 0 {
		s.txnManager.AbortTxn(tx)
		return errors.NewBadRequestError("User is not a winner", nil)
	}

	// Update email if different
	if req.UserEmail != "" {
		if user.Email == nil || *user.Email != req.UserEmail {
			fields := map[string]interface{}{
				"email": req.UserEmail,
			}
			if err := s.userRepo.UpdateFields(ctx, tx, userID, fields); err != nil {
				s.txnManager.AbortTxn(tx)
				return errors.NewInternalServerError("Failed to update user email", err)
			}
		}
	}

	// Update name if provided and different
	if req.UserName != "" {
		if user.Name == nil || *user.Name != req.UserName {
			fields := map[string]interface{}{
				"name": req.UserName,
			}
			if err := s.userRepo.UpdateFields(ctx, tx, userID, fields); err != nil {
				s.txnManager.AbortTxn(tx)
				return errors.NewInternalServerError("Failed to update user name", err)
			}
		}
	}

	now := time.Now()

	if req.AadharNumber != nil || req.AadharFront != nil || req.AadharBack != nil {
		existingCard, err := s.userAadharRepo.FindByUserID(ctx, tx, userID)
		if err != nil {
			s.txnManager.AbortTxn(tx)
			return errors.NewInternalServerError("Failed to fetch user aadhar card", err)
		}
		if existingCard == nil {
			aadharNumber := ""
			if req.AadharNumber != nil {
				aadharNumber = *req.AadharNumber
			}
			aadharFront := ""
			if req.AadharFront != nil {
				aadharFront = *req.AadharFront
			}
			aadharBack := ""
			if req.AadharBack != nil {
				aadharBack = *req.AadharBack
			}

			card := &entities.UserAadharCard{
				UserID:         userID,
				AadharNumber:   aadharNumber,
				AadharFrontKey: aadharFront,
				AadharBackKey:  aadharBack,
				IsDeleted:      false,
				CreatedBy:      userID,
				CreatedOn:      now,
			}
			if err := s.userAadharRepo.Create(ctx, tx, card); err != nil {
				s.txnManager.AbortTxn(tx)
				return errors.NewInternalServerError("Failed to save user aadhar card", err)
			}
		} else {
			if req.AadharNumber != nil {
				existingCard.AadharNumber = *req.AadharNumber
			}
			if req.AadharFront != nil {
				existingCard.AadharFrontKey = *req.AadharFront
			}
			if req.AadharBack != nil {
				existingCard.AadharBackKey = *req.AadharBack
			}
			existingCard.LastModifiedBy = &userID
			existingCard.LastModifiedOn = &now

			if err := s.userAadharRepo.Update(ctx, tx, existingCard); err != nil {
				s.txnManager.AbortTxn(tx)
				return errors.NewInternalServerError("Failed to update user aadhar card", err)
			}
		}
	}

	existingInfo, err := s.userAdditionalInfoRepo.FindByUserID(ctx, tx, userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return errors.NewInternalServerError("Failed to fetch user additional info", err)
	}

	if existingInfo == nil {
		info := &entities.UserAdditionalInfo{
			UserID:    userID,
			City1:     req.City1,
			City2:     req.City2,
			City3:     req.City3,
			IsDeleted: false,
			CreatedBy: userID,
			CreatedOn: now,
		}
		if err := s.userAdditionalInfoRepo.Create(ctx, tx, info); err != nil {
			s.txnManager.AbortTxn(tx)
			return errors.NewInternalServerError("Failed to save user additional info", err)
		}
	} else {
		existingInfo.City1 = req.City1
		existingInfo.City2 = req.City2
		existingInfo.City3 = req.City3
		existingInfo.LastModifiedBy = &userID
		existingInfo.LastModifiedOn = &now

		if err := s.userAdditionalInfoRepo.Update(ctx, tx, existingInfo); err != nil {
			s.txnManager.AbortTxn(tx)
			return errors.NewInternalServerError("Failed to update user additional info", err)
		}
	}

	s.txnManager.CommitTxn(tx)
	return nil
}

func (s *winnerService) CheckUserWinnerStatus(ctx context.Context, userID string) (*dtos.WinnerStatusResponse, error) {
	thunderSeatEntries, err := s.thunderSeatRepo.FindByUserID(ctx, s.txnManager.GetDB(), userID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to check user participation", err)
	}
	hasParticipated := len(thunderSeatEntries) > 0

	winner, err := s.winnerRepo.FindLatestByUserID(ctx, s.txnManager.GetDB(), userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return &dtos.WinnerStatusResponse{
				HasWon:          false,
				HasViewed:       false,
				HasParticipated: hasParticipated,
			}, nil
		}
		return nil, errors.NewInternalServerError("Failed to check winner status", err)
	}

	var qrURL *string
	if winner.QRCode != "" {
		url := s.gcsService.GetPublicURL(winner.QRCode)
		qrURL = &url
	}

	weekNumber := winner.WeekNumber
	return &dtos.WinnerStatusResponse{
		HasWon:          true,
		HasViewed:       winner.HasViewed,
		HasParticipated: hasParticipated,
		WeekNumber:      &weekNumber,
		QRCodeURL:       qrURL,
	}, nil
}

func (s *winnerService) MarkBannerAsViewed(ctx context.Context, userID string) error {
	winner, err := s.winnerRepo.FindLatestByUserID(ctx, s.txnManager.GetDB(), userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NewNotFoundError("User is not a winner", nil)
		}
		return errors.NewInternalServerError("Failed to find winner", err)
	}

	if winner.HasViewed {
		return nil
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.winnerRepo.UpdateHasViewed(ctx, tx, winner.ID)
	})
	if err != nil {
		return errors.NewInternalServerError("Failed to update banner status", err)
	}

	return nil
}
