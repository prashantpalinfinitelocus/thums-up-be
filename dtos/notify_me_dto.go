package dtos

type NotifyMeRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required,min=10,max=10,numeric"`
	Email string `json:"email,omitempty" binding:"omitempty,email"`
}

type NotifyMeResponse struct {
	ID          string  `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	Email       *string `json:"email,omitempty"`
	IsNotified  bool    `json:"is_notified"`
}

