package dtos

import (
	"time"
)

type CreateAvatarRequestDTO struct {
	Name        string `form:"name" binding:"required"`
	IsPublished bool   `form:"is_published"`
}

type AvatarResponseDTO struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	ImageURL       string     `json:"image_url"`
	IsPublished    bool       `json:"is_published"`
	PublishedBy    *string    `json:"published_by,omitempty"`
	PublishedOn    *time.Time `json:"published_on,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedOn      time.Time  `json:"created_on"`
	LastModifiedOn *time.Time `json:"last_modified_on,omitempty"`
}

type AddressRequestDTO struct {
	Address1        string  `json:"address1" binding:"required"`
	Address2        *string `json:"address2,omitempty"`
	Pincode         int     `json:"pincode" binding:"required"`
	State           string  `json:"state" binding:"required"`
	City            string  `json:"city" binding:"required"`
	NearestLandmark *string `json:"nearest_landmark,omitempty"`
	ShippingMobile  *string `json:"shipping_mobile,omitempty"`
	IsDefault       bool    `json:"is_default"`
}

type AddressResponseDTO struct {
	ID              int        `json:"id"`
	Address1        string     `json:"address1"`
	Address2        *string    `json:"address2,omitempty"`
	Pincode         int        `json:"pincode"`
	State           string     `json:"state"`
	City            string     `json:"city"`
	NearestLandmark *string    `json:"nearest_landmark,omitempty"`
	ShippingMobile  *string    `json:"shipping_mobile,omitempty"`
	IsDefault       bool       `json:"is_default"`
	IsActive        bool       `json:"is_active"`
	CreatedOn       time.Time  `json:"created_on"`
	LastModifiedOn  *time.Time `json:"last_modified_on,omitempty"`
}

type UpdateProfileRequestDTO struct {
	Name             *string `json:"name,omitempty"`
	Email            *string `json:"email,omitempty"`
	AvatarID         *int    `json:"avatar_id,omitempty"`
	IsViewed         *bool   `json:"is_viewed,omitempty"`
	SharingPlatform  *string `json:"sharing_platform,omitempty"`
	PlatformUserName *string `json:"platform_user_name,omitempty"`
}

type UserProfileDTO struct {
	ID           string    `json:"id"`
	PhoneNumber  string    `json:"phone_number"`
	Name         *string   `json:"name,omitempty"`
	Email        *string   `json:"email,omitempty"`
	AvatarImage  *string   `json:"avatar_image,omitempty"`
	IsActive     bool      `json:"is_active"`
	IsVerified   bool      `json:"is_verified"`
	ReferralCode *string   `json:"referral_code,omitempty"`
	ReferredBy   *string   `json:"referred_by,omitempty"`
	IsViewed     bool      `json:"is_viewed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProfileResponseDTO struct {
	User UserProfileDTO `json:"user"`
}
