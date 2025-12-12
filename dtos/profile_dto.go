package dtos

import (
	"time"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
)

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
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

type ProfileResponseDTO struct {
	User entities.User `json:"user"`
}
