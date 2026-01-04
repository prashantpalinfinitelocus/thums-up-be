package dtos

type WinnerKYCRequest struct {
	UserName     string `json:"user_name" binding:"required"`
	UserEmail    string `json:"user_email" binding:"required,email"`
	AadharNumber string `json:"aadhar_number" binding:"required"`
	AadharFront  string `json:"aadhar_front" binding:"required"`
	AadharBack   string `json:"aadhar_back" binding:"required"`
	City1        string `json:"city1"`
	City2        string `json:"city2"`
	City3        string `json:"city3"`
}
