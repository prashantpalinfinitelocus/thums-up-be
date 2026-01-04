package services

import (
	"context"

	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

type StateService interface {
	GetAllStates(ctx context.Context) ([]entities.State, error)
}

type stateService struct {
	db        *gorm.DB
	stateRepo repository.StateRepository
}

func NewStateService(db *gorm.DB, stateRepo repository.StateRepository) StateService {
	return &stateService{
		db:        db,
		stateRepo: stateRepo,
	}
}

func (s *stateService) GetAllStates(ctx context.Context) ([]entities.State, error) {
	return s.stateRepo.FindAllActive(ctx, s.db)
}

