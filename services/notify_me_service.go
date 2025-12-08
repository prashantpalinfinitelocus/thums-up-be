package services

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type NotifyMeService interface {
	Subscribe(ctx context.Context, req dtos.NotifyMeRequest) (*dtos.NotifyMeResponse, bool, error)
	GetSubscription(ctx context.Context, phoneNumber string) (*dtos.NotifyMeResponse, error)
	GetAllUnnotified(ctx context.Context) ([]dtos.NotifyMeResponse, error)
	MarkAsNotified(ctx context.Context, id string) error
}

type notifyMeService struct {
	txnManager   *utils.TransactionManager
	notifyMeRepo repository.NotifyMeRepository
}

func NewNotifyMeService(
	txnManager *utils.TransactionManager,
	notifyMeRepo repository.NotifyMeRepository,
) NotifyMeService {
	return &notifyMeService{
		txnManager:   txnManager,
		notifyMeRepo: notifyMeRepo,
	}
}

func (s *notifyMeService) Subscribe(ctx context.Context, req dtos.NotifyMeRequest) (*dtos.NotifyMeResponse, bool, error) {
	existing, _ := s.notifyMeRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), req.PhoneNumber)
	if existing != nil {
		return &dtos.NotifyMeResponse{
			ID:          existing.ID,
			PhoneNumber: existing.PhoneNumber,
			Email:       existing.Email,
			IsNotified:  existing.IsNotified,
		}, false, nil
	}

	notifyMe := &entities.NotifyMe{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		IsNotified:  false,
	}

	err := s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.notifyMeRepo.Create(ctx, tx, notifyMe)
	})
	if err != nil {
		log.WithError(err).Error("Failed to create notify me subscription")
		return nil, false, errors.NewInternalServerError("Failed to subscribe", err)
	}

	return &dtos.NotifyMeResponse{
		ID:          notifyMe.ID,
		PhoneNumber: notifyMe.PhoneNumber,
		Email:       notifyMe.Email,
		IsNotified:  notifyMe.IsNotified,
	}, true, nil
}

func (s *notifyMeService) GetSubscription(ctx context.Context, phoneNumber string) (*dtos.NotifyMeResponse, error) {
	notifyMe, err := s.notifyMeRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), phoneNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Subscription not found", err)
		}
		return nil, errors.NewInternalServerError("Failed to get subscription", err)
	}

	return &dtos.NotifyMeResponse{
		ID:          notifyMe.ID,
		PhoneNumber: notifyMe.PhoneNumber,
		Email:       notifyMe.Email,
		IsNotified:  notifyMe.IsNotified,
	}, nil
}

func (s *notifyMeService) GetAllUnnotified(ctx context.Context) ([]dtos.NotifyMeResponse, error) {
	records, err := s.notifyMeRepo.FindUnnotified(ctx, s.txnManager.GetDB())
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get unnotified subscriptions", err)
	}

	responses := make([]dtos.NotifyMeResponse, len(records))
	for i, record := range records {
		responses[i] = dtos.NotifyMeResponse{
			ID:          record.ID,
			PhoneNumber: record.PhoneNumber,
			Email:       record.Email,
			IsNotified:  record.IsNotified,
		}
	}

	return responses, nil
}

func (s *notifyMeService) MarkAsNotified(ctx context.Context, id string) error {
	err := s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.notifyMeRepo.MarkAsNotified(ctx, tx, id)
	})
	if err != nil {
		return errors.NewInternalServerError("Failed to mark as notified", err)
	}
	return nil
}
