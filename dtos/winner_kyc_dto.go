package dtos

type WinnerKYCRequest struct {
	UserName     string  `json:"user_name" binding:"required"`
	UserEmail    string  `json:"user_email" binding:"required,email"`
	AadharNumber *string `json:"aadhar_number,omitempty"`
	AadharFront  *string `json:"aadhar_front,omitempty"`
	AadharBack   *string `json:"aadhar_back,omitempty"`
	City1        string  `json:"city1"`
	City2        string  `json:"city2"`
	City3        string  `json:"city3"`
}
