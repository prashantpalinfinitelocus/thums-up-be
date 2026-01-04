package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

type WebsiteStatusService interface {
	GetStatus(ctx context.Context) *dtos.WebsiteStatusResponse
}

type websiteStatusService struct {
	db         *gorm.DB
	winnerRepo repository.WinnerRepository
}

func NewWebsiteStatusService(db *gorm.DB, winnerRepo repository.WinnerRepository) WebsiteStatusService {
	return &websiteStatusService{
		db:         db,
		winnerRepo: winnerRepo,
	}
}

func (s *websiteStatusService) GetStatus(ctx context.Context) *dtos.WebsiteStatusResponse {
	launchDate, err := time.Parse(time.RFC3339, constants.WEBSITE_LAUNCH_DATE)
	if err != nil {
		return &dtos.WebsiteStatusResponse{
			Status: "coming_soon",
		}
	}

	now := time.Now()

	if now.After(launchDate) {
		// Check if there are winners in the table
		winnerCount, err := s.winnerRepo.Count(ctx, s.db)
		if err == nil && winnerCount > 1 {
			return &dtos.WebsiteStatusResponse{
				Status: "live_with_winners",
			}
		}

		return &dtos.WebsiteStatusResponse{
			Status: "live",
		}
	}

	timeRemaining := launchDate.Sub(now)
	hoursRemaining := timeRemaining.Hours()

	if hoursRemaining <= 48 {
		return &dtos.WebsiteStatusResponse{
			Status: "live_soon",
		}
	}

	return &dtos.WebsiteStatusResponse{
		Status: "coming_soon",
	}
}
