package dtos

type ThunderSeatSubmitRequest struct {
	WeekNumber int    `json:"week_number" binding:"required"`
	QuestionID int    `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

type ThunderSeatResponse struct {
	ID         int    `json:"id"`
	UserID     string `json:"user_id"`
	QuestionID int    `json:"question_id"`
	WeekNumber int    `json:"week_number"`
	Answer     string `json:"answer"`
	CreatedOn  string `json:"created_on"`
}

type SelectWinnersRequest struct {
	WeekNumber      int `json:"week_number" binding:"required"`
	NumberOfWinners int `json:"number_of_winners" binding:"required,min=1"`
}

type WinnerResponse struct {
	ID            int    `json:"id"`
	UserID        string `json:"user_id"`
	ThunderSeatID int    `json:"thunder_seat_id"`
	WeekNumber    int    `json:"week_number"`
	CreatedOn     string `json:"created_on"`
}

type CurrentWeekResponse struct {
	WeekNumber int    `json:"week_number"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type AllWinnersRequest struct {
	Limit  int `form:"limit" binding:"required,min=1,max=100"`
	Offset int `form:"offset" binding:"min=0"`
}
