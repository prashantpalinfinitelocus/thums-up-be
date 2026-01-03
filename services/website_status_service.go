package services

import (
	"time"

	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
)

type WebsiteStatusService interface {
	GetStatus() *dtos.WebsiteStatusResponse
}

type websiteStatusService struct{}

func NewWebsiteStatusService() WebsiteStatusService {
	return &websiteStatusService{}
}

func (s *websiteStatusService) GetStatus() *dtos.WebsiteStatusResponse {
	launchDate, err := time.Parse(time.RFC3339, constants.WEBSITE_LAUNCH_DATE)
	if err != nil {
		return &dtos.WebsiteStatusResponse{
			Status: "coming_soon",
		}
	}

	now := time.Now()

	if now.After(launchDate) {
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
