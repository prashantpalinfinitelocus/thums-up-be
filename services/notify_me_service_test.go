package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
)

// MockNotifyMeRepository is a mock for the NotifyMeRepository
type MockNotifyMeRepository struct {
	mock.Mock
}

func (m *MockNotifyMeRepository) Create(ctx context.Context, db *gorm.DB, entity *entities.NotifyMe) error {
	args := m.Called(ctx, db, entity)
	return args.Error(0)
}

func (m *MockNotifyMeRepository) FindByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.NotifyMe, error) {
	args := m.Called(ctx, db, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.NotifyMe), args.Error(1)
}

func (m *MockNotifyMeRepository) FindUnnotified(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.NotifyMe, error) {
	args := m.Called(ctx, db, limit, offset)
	return args.Get(0).([]entities.NotifyMe), args.Error(1)
}

func (m *MockNotifyMeRepository) MarkAsNotified(ctx context.Context, db *gorm.DB, id string) error {
	args := m.Called(ctx, db, id)
	return args.Error(0)
}

// MockTransactionManager is a mock for the TransactionManager
type MockTransactionManager struct {
	mock.Mock
	db *gorm.DB
}

func (m *MockTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	// For testing, just execute the function with the mock DB
	return fn(m.db)
}

func (m *MockTransactionManager) GetDB() *gorm.DB {
	return m.db
}

func TestSubscribe_NewSubscription(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotifyMeRepository)
	mockTxnManager := &MockTransactionManager{}
	
	service := &notifyMeService{
		txnManager:   mockTxnManager,
		notifyMeRepo: mockRepo,
	}

	req := dtos.NotifyMeRequest{
		Name:        "Test User",
		PhoneNumber: "1234567890",
		Email:       "test@example.com",
	}

	mockRepo.On("FindByPhoneNumber", mock.Anything, mock.Anything, req.PhoneNumber).
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*entities.NotifyMe")).
		Return(nil)

	// Act
	response, created, err := service.Subscribe(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, response)
	assert.Equal(t, req.PhoneNumber, response.PhoneNumber)
	mockRepo.AssertExpectations(t)
}

func TestSubscribe_ExistingSubscription(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotifyMeRepository)
	mockTxnManager := &MockTransactionManager{}
	
	service := &notifyMeService{
		txnManager:   mockTxnManager,
		notifyMeRepo: mockRepo,
	}

	req := dtos.NotifyMeRequest{
		Name:        "Test User",
		PhoneNumber: "1234567890",
		Email:       "test@example.com",
	}

	existing := &entities.NotifyMe{
		ID:          "test-id",
		PhoneNumber: req.PhoneNumber,
		IsNotified:  false,
	}

	mockRepo.On("FindByPhoneNumber", mock.Anything, mock.Anything, req.PhoneNumber).
		Return(existing, nil)

	// Act
	response, created, err := service.Subscribe(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.False(t, created)
	assert.NotNil(t, response)
	assert.Equal(t, existing.ID, response.ID)
	mockRepo.AssertExpectations(t)
}

