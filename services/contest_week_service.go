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

type ContestWeekService interface {
	CreateContestWeek(ctx context.Context, req dtos.ContestWeekRequest, createdBy string) (*dtos.ContestWeekResponse, error)
	GetAllContestWeeks(ctx context.Context) ([]dtos.ContestWeekResponse, error)
	GetContestWeekByNumber(ctx context.Context, weekNumber int) (*dtos.ContestWeekResponse, error)
	ActivateWeek(ctx context.Context, weekNumber int) (*dtos.ContestWeekResponse, error)
	GetActiveWeek(ctx context.Context) (*dtos.ContestWeekResponse, error)
}

type contestWeekService struct {
	txnManager      *utils.TransactionManager
	contestWeekRepo repository.ContestWeekRepository
}

func NewContestWeekService(
	txnManager *utils.TransactionManager,
	contestWeekRepo repository.ContestWeekRepository,
) ContestWeekService {
	return &contestWeekService{
		txnManager:      txnManager,
		contestWeekRepo: contestWeekRepo,
	}
}

func (s *contestWeekService) CreateContestWeek(ctx context.Context, req dtos.ContestWeekRequest, createdBy string) (*dtos.ContestWeekResponse, error) {
	existing, err := s.contestWeekRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), req.WeekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to check existing week", err)
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("Contest week already exists", nil)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid start date format. Use YYYY-MM-DD", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid end date format. Use YYYY-MM-DD", err)
	}

	if endDate.Before(startDate) {
		return nil, errors.NewBadRequestError("End date must be after start date", nil)
	}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)

	now := time.Now()
	contestWeek := &entities.ContestWeek{
		WeekNumber:  req.WeekNumber,
		StartDate:   startDate,
		EndDate:     endDate,
		WinnerCount: req.WinnerCount,
		IsActive:    false,
		CreatedBy:   createdBy,
		CreatedOn:   now,
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.contestWeekRepo.Create(ctx, tx, contestWeek)
	})
	if err != nil {
		log.WithError(err).Error("Failed to create contest week")
		return nil, errors.NewInternalServerError("Failed to create contest week", err)
	}

	return &dtos.ContestWeekResponse{
		ID:          contestWeek.ID,
		WeekNumber:  contestWeek.WeekNumber,
		StartDate:   contestWeek.StartDate.Format("2006-01-02"),
		EndDate:     contestWeek.EndDate.Format("2006-01-02"),
		WinnerCount: contestWeek.WinnerCount,
		IsActive:    contestWeek.IsActive,
		CreatedOn:   contestWeek.CreatedOn.Format(time.RFC3339),
	}, nil
}

func (s *contestWeekService) GetAllContestWeeks(ctx context.Context) ([]dtos.ContestWeekResponse, error) {
	weeks, err := s.contestWeekRepo.FindAll(ctx, s.txnManager.GetDB())
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get contest weeks", err)
	}

	responses := make([]dtos.ContestWeekResponse, len(weeks))
	for i, week := range weeks {
		responses[i] = dtos.ContestWeekResponse{
			ID:          week.ID,
			WeekNumber:  week.WeekNumber,
			StartDate:   week.StartDate.Format("2006-01-02"),
			EndDate:     week.EndDate.Format("2006-01-02"),
			WinnerCount: week.WinnerCount,
			IsActive:    week.IsActive,
			CreatedOn:   week.CreatedOn.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *contestWeekService) GetContestWeekByNumber(ctx context.Context, weekNumber int) (*dtos.ContestWeekResponse, error) {
	week, err := s.contestWeekRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), weekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get contest week", err)
	}
	if week == nil {
		return nil, errors.NewNotFoundError("Contest week not found", nil)
	}

	return &dtos.ContestWeekResponse{
		ID:          week.ID,
		WeekNumber:  week.WeekNumber,
		StartDate:   week.StartDate.Format("2006-01-02"),
		EndDate:     week.EndDate.Format("2006-01-02"),
		WinnerCount: week.WinnerCount,
		IsActive:    week.IsActive,
		CreatedOn:   week.CreatedOn.Format(time.RFC3339),
	}, nil
}

func (s *contestWeekService) ActivateWeek(ctx context.Context, weekNumber int) (*dtos.ContestWeekResponse, error) {
	week, err := s.contestWeekRepo.FindByWeekNumber(ctx, s.txnManager.GetDB(), weekNumber)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get contest week", err)
	}
	if week == nil {
		return nil, errors.NewNotFoundError("Contest week not found", nil)
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.contestWeekRepo.DeactivateAll(ctx, tx); err != nil {
			return err
		}

		week.IsActive = true
		week.UpdatedOn = time.Now()
		return s.contestWeekRepo.Update(ctx, tx, week)
	})
	if err != nil {
		log.WithError(err).Error("Failed to activate contest week")
		return nil, errors.NewInternalServerError("Failed to activate contest week", err)
	}

	return &dtos.ContestWeekResponse{
		ID:          week.ID,
		WeekNumber:  week.WeekNumber,
		StartDate:   week.StartDate.Format("2006-01-02"),
		EndDate:     week.EndDate.Format("2006-01-02"),
		WinnerCount: week.WinnerCount,
		IsActive:    week.IsActive,
		CreatedOn:   week.CreatedOn.Format(time.RFC3339),
	}, nil
}

func (s *contestWeekService) GetActiveWeek(ctx context.Context) (*dtos.ContestWeekResponse, error) {
	week, err := s.contestWeekRepo.FindActiveWeek(ctx, s.txnManager.GetDB())
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get active week", err)
	}
	if week == nil {
		return nil, errors.NewNotFoundError("No active contest week found", nil)
	}

	return &dtos.ContestWeekResponse{
		ID:          week.ID,
		WeekNumber:  week.WeekNumber,
		StartDate:   week.StartDate.Format("2006-01-02"),
		EndDate:     week.EndDate.Format("2006-01-02"),
		WinnerCount: week.WinnerCount,
		IsActive:    week.IsActive,
		CreatedOn:   week.CreatedOn.Format(time.RFC3339),
	}, nil
}
