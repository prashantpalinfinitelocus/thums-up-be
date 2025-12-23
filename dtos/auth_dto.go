package dtos

type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,min=10,max=10,numeric"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,min=10,max=10,numeric"`
	OTP         string `json:"otp" binding:"required,min=6,max=6,numeric"`
}

type SignUpRequest struct {
	PhoneNumber  string  `json:"phone_number" binding:"required,min=10,max=10,numeric"`
	Name         string  `json:"name" binding:"required"`
	Email        *string `json:"email,omitempty" binding:"omitempty,email"`
	ReferralCode *string `json:"referral_code,omitempty"`
	DeviceToken  *string `json:"device_token,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserResponse struct {
	ID           string  `json:"id"`
	PhoneNumber  string  `json:"phone_number"`
	Name         *string `json:"name,omitempty"`
	Email        *string `json:"email,omitempty"`
	ReferralCode *string `json:"referral_code,omitempty"`
	IsVerified   bool    `json:"is_verified"`
}
