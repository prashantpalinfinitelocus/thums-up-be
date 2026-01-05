package dtos

type ThunderSeatSubmitRequest struct {
	Answer           string  `form:"description" binding:"required"`
	SharingPlatform  *string `form:"social_media" binding:"omitempty,oneof=instagram snapchat facebook twitter tiktok youtube"`
	PlatformUserName *string `form:"user_name" binding:"omitempty,min=3,max=255"`
}

type ThunderSeatResponse struct {
	ID         int     `json:"id"`
	UserID     string  `json:"user_id"`
	WeekNumber int     `json:"week_number"`
	Answer     string  `json:"answer"`
	MediaURL   *string `json:"media_url,omitempty"`
	MediaType  *string `json:"media_type,omitempty"`
	CreatedOn  string  `json:"created_on"`
	Name       *string `json:"name,omitempty"`
	Email      *string `json:"email,omitempty"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
	AvatarName *string `json:"avatar_name,omitempty"`
}

type SelectWinnersRequest struct {
	WeekNumber int `json:"week_number" binding:"required"`
}

type WinnerResponse struct {
	ID            int     `json:"id"`
	UserID        string  `json:"user_id"`
	ThunderSeatID int     `json:"thunder_seat_id"`
	WeekNumber    int     `json:"week_number"`
	QRCodeURL     *string `json:"qr_code_url,omitempty"`
	CreatedOn     string  `json:"created_on"`
	Name          *string `json:"name,omitempty"`
	Email         *string `json:"email,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	AvatarName    *string `json:"avatar_name,omitempty"`
}

type CurrentWeekResponse struct {
	WeekNumber  int    `json:"week_number"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	WinnerCount int    `json:"winner_count"`
	IsActive    bool   `json:"is_active"`
}

type AllWinnersRequest struct {
	Limit  int `form:"limit" binding:"required,min=1,max=100"`
	Offset int `form:"offset" binding:"min=0"`
}

type ContestWeekRequest struct {
	WeekNumber  int    `json:"week_number" binding:"required,min=1"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	WinnerCount int    `json:"winner_count" binding:"required,min=1"`
}

type ContestWeekResponse struct {
	ID          int    `json:"id"`
	WeekNumber  int    `json:"week_number"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	WinnerCount int    `json:"winner_count"`
	IsActive    bool   `json:"is_active"`
	CreatedOn   string `json:"created_on"`
}

type ActivateWeekRequest struct {
	WeekNumber int `json:"week_number" binding:"required"`
}

type WinnerStatusResponse struct {
	HasWon         bool    `json:"has_won"`
	HasViewed      bool    `json:"has_viewed"`
	HasParticipated bool    `json:"has_participated"`
	WeekNumber     *int    `json:"week_number,omitempty"`
	QRCodeURL      *string `json:"qr_code_url,omitempty"`
}
