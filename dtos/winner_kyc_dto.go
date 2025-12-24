package dtos

type WinnerFriendDTO struct {
	Name         string `json:"name" binding:"required"`
	UUID         string `json:"uuid" binding:"required"`
	AadharNumber string `json:"aadhar_number" binding:"required"`
	AadharFront  string `json:"aadhar_front" binding:"required"`
	AadharBack   string `json:"aadhar_back" binding:"required"`
}

type WinnerKYCRequest struct {
	UserName      string           `json:"user_name" binding:"required"`
	UserEmail     string           `json:"user_email" binding:"required,email"`
	AadharNumber  string           `json:"aadhar_number" binding:"required"`
	AadharFront   string           `json:"aadhar_front" binding:"required"`
	AadharBack    string           `json:"aadhar_back" binding:"required"`
	Friends       []WinnerFriendDTO `json:"friends" binding:"max=10"`
}



