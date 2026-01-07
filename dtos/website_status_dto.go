package dtos

import "time"

type WebsiteStatusResponse struct {
	Status                 string     `json:"status"`
	WinnerAnnouncement     *bool      `json:"winner_announcement,omitempty"`
	WinnerAnnouncementDate *time.Time `json:"winner_announcement_date,omitempty"`
}
