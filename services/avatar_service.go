package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type AvatarService interface {
	CreateAvatar(ctx context.Context, req dtos.CreateAvatarRequestDTO, imageFile *multipart.FileHeader, createdBy string) (*dtos.AvatarResponseDTO, error)
	GetAllAvatars(ctx context.Context, isPublished *bool) ([]dtos.AvatarResponseDTO, error)
	GetAvatarByID(ctx context.Context, avatarID int) (*dtos.AvatarResponseDTO, error)
}

type avatarService struct {
	txnManager *utils.TransactionManager
	avatarRepo repository.GenericRepository[entities.Avatar]
	gcsService utils.GCSService
}

func NewAvatarService(
	txnManager *utils.TransactionManager,
	avatarRepo repository.GenericRepository[entities.Avatar],
	gcsService utils.GCSService,
) AvatarService {
	return &avatarService{
		txnManager: txnManager,
		avatarRepo: avatarRepo,
		gcsService: gcsService,
	}
}

func (s *avatarService) CreateAvatar(ctx context.Context, req dtos.CreateAvatarRequestDTO, imageFile *multipart.FileHeader, createdBy string) (*dtos.AvatarResponseDTO, error) {
	if s.gcsService == nil {
		return nil, fmt.Errorf("GCS service is not initialized")
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	// Upload image file to GCS
	folderPath := fmt.Sprintf("avatars/%s", createdBy)
	imageURL, imageKey, err := s.gcsService.UploadFile(ctx, imageFile, folderPath)
	if err != nil {
		log.WithError(err).Error("Failed to upload avatar image to GCS")
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to upload avatar image: %w", err)
	}

	now := time.Now()
	avatar := &entities.Avatar{
		Name:        req.Name,
		ImageKey:    imageKey,
		IsPublished: req.IsPublished,
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   createdBy,
		CreatedOn:   now,
	}

	if req.IsPublished {
		avatar.PublishedBy = &createdBy
		avatar.PublishedOn = &now
	}

	if err := s.avatarRepo.Create(ctx, tx, avatar); err != nil {
		s.txnManager.AbortTxn(tx)
		// Cleanup uploaded file if database operation fails
		if deleteErr := s.gcsService.DeleteFile(ctx, imageURL); deleteErr != nil {
			log.WithError(deleteErr).Error("Failed to cleanup uploaded avatar image after database error")
		}
		return nil, fmt.Errorf("failed to create avatar: %w", err)
	}

	response := &dtos.AvatarResponseDTO{
		ID:             avatar.ID,
		Name:           avatar.Name,
		ImageURL:       imageURL,
		IsPublished:    avatar.IsPublished,
		PublishedBy:    avatar.PublishedBy,
		PublishedOn:    avatar.PublishedOn,
		IsActive:       avatar.IsActive,
		CreatedOn:      avatar.CreatedOn,
		LastModifiedOn: avatar.LastModifiedOn,
	}

	s.txnManager.CommitTxn(tx)
	return response, nil
}

func (s *avatarService) GetAllAvatars(ctx context.Context, isPublished *bool) ([]dtos.AvatarResponseDTO, error) {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	conditions := map[string]interface{}{
		"is_deleted": false,
		"is_active":  true,
	}

	if isPublished != nil {
		conditions["is_published"] = *isPublished
	}

	avatars, err := s.avatarRepo.FindWithConditions(ctx, tx, conditions)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to fetch avatars: %w", err)
	}

	response := make([]dtos.AvatarResponseDTO, 0, len(avatars))
	for _, avatar := range avatars {
		var imageURL string
		if s.gcsService != nil {
			// Reconstruct full path: avatars/{userID}/{filename}
			fullPath := fmt.Sprintf("avatars/%s/%s", avatar.CreatedBy, avatar.ImageKey)
			imageURL = s.gcsService.GetPublicURL(fullPath)
		}
		response = append(response, dtos.AvatarResponseDTO{
			ID:             avatar.ID,
			Name:           avatar.Name,
			ImageURL:       imageURL,
			IsPublished:    avatar.IsPublished,
			PublishedBy:    avatar.PublishedBy,
			PublishedOn:    avatar.PublishedOn,
			IsActive:       avatar.IsActive,
			CreatedOn:      avatar.CreatedOn,
			LastModifiedOn: avatar.LastModifiedOn,
		})
	}

	s.txnManager.CommitTxn(tx)
	return response, nil
}

func (s *avatarService) GetAvatarByID(ctx context.Context, avatarID int) (*dtos.AvatarResponseDTO, error) {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	avatar, err := s.avatarRepo.FindByID(ctx, tx, avatarID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}

	if avatar == nil {
		s.txnManager.AbortTxn(tx)
		return nil, gorm.ErrRecordNotFound
	}

	if avatar.IsDeleted {
		s.txnManager.AbortTxn(tx)
		return nil, gorm.ErrRecordNotFound
	}

	var imageURL string
	if s.gcsService != nil {
		fullPath := fmt.Sprintf("avatars/%s/%s", avatar.CreatedBy, avatar.ImageKey)
		imageURL = s.gcsService.GetPublicURL(fullPath)
	}

	response := &dtos.AvatarResponseDTO{
		ID:             avatar.ID,
		Name:           avatar.Name,
		ImageURL:       imageURL,
		IsPublished:    avatar.IsPublished,
		PublishedBy:    avatar.PublishedBy,
		PublishedOn:    avatar.PublishedOn,
		IsActive:       avatar.IsActive,
		CreatedOn:      avatar.CreatedOn,
		LastModifiedOn: avatar.LastModifiedOn,
	}

	s.txnManager.CommitTxn(tx)
	return response, nil
}
