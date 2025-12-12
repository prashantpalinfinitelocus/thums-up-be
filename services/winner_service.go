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

type WinnerService interface {
	SelectWinners(ctx context.Context, req dtos.SelectWinnersRequest) ([]dtos.WinnerResponse, error)
	GetWinnersByWeek(ctx context.Context, weekNumber int) ([]dtos.WinnerResponse, error)
	GetAllWinners(ctx context.Context, limit, offset int) ([]dtos.WinnerResponse, int64, error)
}

type winnerService struct {
	txnManager      *utils.TransactionManager
	winnerRepo      repository.WinnerRepository
	thunderSeatRepo repository.ThunderSeatRepository
}

func NewWinnerService(
	txnManager *utils.TransactionManager,
	winnerRepo repository.WinnerRepository,
	thunderSeatRepo repository.ThunderSeatRepository,
) WinnerService {
	return &winnerService{
		txnManager:      txnManager,
		winnerRepo:      winnerRepo,
		thunderSeatRepo: thunderSeatRepo,
	}
}

func (s *winnerService) SelectWinners(ctx context.Context, req dtos.SelectWinnersRequest) ([]dtos.WinnerResponse, error) {
	existingWinnerUserIDs, err := s.winnerRepo.GetWinnerUserIDs(ctx, s.txnManager.GetDB(), req.WeekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get existing winners", err)
	}

	randomEntries, err := s.thunderSeatRepo.GetRandomEntries(ctx, s.txnManager.GetDB(), req.NumberOfWinners, existingWinnerUserIDs)
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
		winners[i] = entities.ThunderSeatWinner{
			UserID:        entry.UserID,
			ThunderSeatID: entry.ID,
			WeekNumber:    req.WeekNumber,
			CreatedBy:     "system",
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
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
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
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
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
		responses[i] = dtos.WinnerResponse{
			ID:            winner.ID,
			UserID:        winner.UserID,
			ThunderSeatID: winner.ThunderSeatID,
			WeekNumber:    winner.WeekNumber,
			CreatedOn:     winner.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, total, nil
}
