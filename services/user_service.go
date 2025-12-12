package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	UpdateUser(ctx context.Context, userID string, req dtos.UpdateProfileRequestDTO) (*entities.User, error)
	GetUserAddresses(ctx context.Context, userID string) ([]dtos.AddressResponseDTO, error)
	AddUserAddress(ctx context.Context, userID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error)
	UpdateUserAddress(ctx context.Context, userID string, addressID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error)
	DeleteUserAddress(ctx context.Context, userID string, addressID string) error
}

type userService struct {
	txnManager  *utils.TransactionManager
	userRepo    repository.UserRepository
	addressRepo repository.GenericRepository[entities.Address]
	stateRepo   repository.StateRepository
	cityRepo    repository.CityRepository
	pinCodeRepo repository.PinCodeRepository
}

func NewUserService(
	txnManager *utils.TransactionManager,
	userRepo repository.UserRepository,
	addressRepo repository.GenericRepository[entities.Address],
	stateRepo repository.StateRepository,
	cityRepo repository.CityRepository,
	pinCodeRepo repository.PinCodeRepository,
) UserService {
	return &userService{
		txnManager:  txnManager,
		userRepo:    userRepo,
		addressRepo: addressRepo,
		stateRepo:   stateRepo,
		cityRepo:    cityRepo,
		pinCodeRepo: pinCodeRepo,
	}
}

func (s *userService) GetUser(ctx context.Context, userID string) (*entities.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindById(ctx, s.txnManager.GetDB(), userUUID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, req dtos.UpdateProfileRequestDTO) (*entities.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("user not found")
	}

	updateFields := make(map[string]interface{})

	if req.Name != nil {
		updateFields["name"] = *req.Name
	}

	if req.Email != nil {
		existingUser, err := s.userRepo.FindByEmail(ctx, tx, *req.Email)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to check email uniqueness: %v", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("email already in use")
		}
		updateFields["email"] = *req.Email
	}

	if len(updateFields) > 0 {
		if err := s.userRepo.UpdateFields(ctx, tx, userID, updateFields); err != nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to update user: %v", err)
		}
	}

	updatedUser, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	s.txnManager.CommitTxn(tx)
	return updatedUser, nil
}

func (s *userService) GetUserAddresses(ctx context.Context, userID string) ([]dtos.AddressResponseDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("user not found")
	}

	addresses, err := s.addressRepo.FindWithConditions(ctx, tx, map[string]interface{}{
		"user_id":    userID,
		"is_deleted": false,
	})
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	stateIDs := make([]int, 0, len(addresses))
	cityIDs := make([]int, 0, len(addresses))
	pinCodeIDs := make([]int, 0, len(addresses))

	for _, addr := range addresses {
		stateIDs = append(stateIDs, addr.StateID)
		cityIDs = append(cityIDs, addr.CityID)
		pinCodeIDs = append(pinCodeIDs, addr.PinCodeID)
	}

	states, err := s.stateRepo.FindByIDs(ctx, tx, stateIDs)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get state details: %v", err)
	}

	cities, err := s.cityRepo.FindByIDs(ctx, tx, cityIDs)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get city details: %v", err)
	}

	pincodes, err := s.pinCodeRepo.FindByIDs(ctx, tx, pinCodeIDs)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get pincode details: %v", err)
	}

	stateMap := make(map[int]*entities.State)
	for _, state := range states {
		stateCopy := state
		stateMap[state.ID] = &stateCopy
	}

	cityMap := make(map[int]*entities.City)
	for _, city := range cities {
		cityCopy := city
		cityMap[city.ID] = &cityCopy
	}

	pincodeMap := make(map[int]*entities.PinCode)
	for _, pincode := range pincodes {
		pincodeCopy := pincode
		pincodeMap[pincode.ID] = &pincodeCopy
	}

	addressDTOs := make([]dtos.AddressResponseDTO, 0, len(addresses))
	for _, addr := range addresses {
		state := stateMap[addr.StateID]
		city := cityMap[addr.CityID]
		pincode := pincodeMap[addr.PinCodeID]

		if state == nil || city == nil || pincode == nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("missing location details for address %d", addr.ID)
		}

		addressDTO := dtos.AddressResponseDTO{
			ID:              addr.ID,
			Address1:        addr.Address1,
			Address2:        addr.Address2,
			Pincode:         pincode.Pincode,
			State:           state.Name,
			City:            city.Name,
			NearestLandmark: addr.NearestLandmark,
			ShippingMobile:  addr.ShippingMobile,
			IsDefault:       addr.IsDefault,
			IsActive:        addr.IsActive,
			CreatedOn:       addr.CreatedOn,
			LastModifiedOn:  addr.LastModifiedOn,
		}
		addressDTOs = append(addressDTOs, addressDTO)
	}

	s.txnManager.CommitTxn(tx)
	return addressDTOs, nil
}

func (s *userService) AddUserAddress(ctx context.Context, userID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("user not found")
	}

	state, err := s.stateRepo.FindByName(ctx, tx, req.State)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find state: %v", err)
	}
	if state == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("state '%s' not found", req.State)
	}

	city, err := s.cityRepo.FindByNameAndStateID(ctx, tx, req.City, state.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find city: %v", err)
	}
	if city == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("city '%s' not found in state '%s'", req.City, req.State)
	}

	pincode, err := s.pinCodeRepo.FindByPincodeAndCityID(ctx, tx, req.Pincode, city.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find pincode: %v", err)
	}
	if pincode == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("pincode '%d' not found in city '%s'", req.Pincode, req.City)
	}

	isDeliverable, err := s.pinCodeRepo.IsDeliverable(ctx, tx, req.Pincode, city.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to check pincode deliverability: %v", err)
	}
	if !isDeliverable {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("pincode '%d' is not deliverable in city '%s'", req.Pincode, req.City)
	}

	if req.IsDefault {
		err := tx.Model(&entities.Address{}).
			Where("user_id = ? AND is_default = ? AND is_deleted = ?", userID, true, false).
			Update("is_default", false).Error
		if err != nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	now := time.Now()
	address := &entities.Address{
		UserID:          userID,
		Address1:        req.Address1,
		Address2:        req.Address2,
		Pincode:         req.Pincode,
		PinCodeID:       pincode.ID,
		CityID:          city.ID,
		StateID:         state.ID,
		NearestLandmark: req.NearestLandmark,
		ShippingMobile:  req.ShippingMobile,
		IsDefault:       req.IsDefault,
		IsActive:        true,
		IsDeleted:       false,
		CreatedBy:       userID,
		CreatedOn:       now,
	}

	if err := s.addressRepo.Create(ctx, tx, address); err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	addressDTO := &dtos.AddressResponseDTO{
		ID:              address.ID,
		Address1:        address.Address1,
		Address2:        address.Address2,
		Pincode:         pincode.Pincode,
		State:           state.Name,
		City:            city.Name,
		NearestLandmark: address.NearestLandmark,
		ShippingMobile:  address.ShippingMobile,
		IsDefault:       address.IsDefault,
		IsActive:        address.IsActive,
		CreatedOn:       address.CreatedOn,
		LastModifiedOn:  address.LastModifiedOn,
	}

	s.txnManager.CommitTxn(tx)
	return addressDTO, nil
}

func (s *userService) UpdateUserAddress(ctx context.Context, userID string, addressID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("user not found")
	}

	addrID, err := strconv.ParseInt(addressID, 10, 64)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("invalid address ID format")
	}

	address, err := s.addressRepo.FindByID(ctx, tx, int(addrID))
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	if address == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("address not found")
	}

	if address.UserID != userID {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("address does not belong to user")
	}

	state, err := s.stateRepo.FindByName(ctx, tx, req.State)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find state: %v", err)
	}
	if state == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("state '%s' not found", req.State)
	}

	city, err := s.cityRepo.FindByNameAndStateID(ctx, tx, req.City, state.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find city: %v", err)
	}
	if city == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("city '%s' not found in state '%s'", req.City, req.State)
	}

	pincode, err := s.pinCodeRepo.FindByPincodeAndCityID(ctx, tx, req.Pincode, city.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to find pincode: %v", err)
	}
	if pincode == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("pincode '%d' not found in city '%s'", req.Pincode, req.City)
	}

	isDeliverable, err := s.pinCodeRepo.IsDeliverable(ctx, tx, req.Pincode, city.ID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to check pincode deliverability: %v", err)
	}
	if !isDeliverable {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("pincode '%d' is not deliverable in city '%s'", req.Pincode, req.City)
	}

	if req.IsDefault && !address.IsDefault {
		err := tx.Model(&entities.Address{}).
			Where("user_id = ? AND is_default = ? AND is_deleted = ?", userID, true, false).
			Update("is_default", false).Error
		if err != nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	now := time.Now()
	updateFields := map[string]interface{}{
		"address1":         req.Address1,
		"pincode":          req.Pincode,
		"pin_code_id":      pincode.ID,
		"city_id":          city.ID,
		"state_id":         state.ID,
		"is_default":       req.IsDefault,
		"last_modified_by": userID,
		"last_modified_on": now,
	}

	if req.Address2 != nil {
		updateFields["address2"] = *req.Address2
	}
	if req.NearestLandmark != nil {
		updateFields["nearest_landmark"] = *req.NearestLandmark
	}
	if req.ShippingMobile != nil {
		updateFields["shipping_mobile"] = *req.ShippingMobile
	}

	if err := s.addressRepo.UpdateFields(ctx, tx, int(addrID), updateFields); err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	updatedAddress, err := s.addressRepo.FindByID(ctx, tx, int(addrID))
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}

	addressDTO := &dtos.AddressResponseDTO{
		ID:              updatedAddress.ID,
		Address1:        updatedAddress.Address1,
		Address2:        updatedAddress.Address2,
		Pincode:         pincode.Pincode,
		State:           state.Name,
		City:            city.Name,
		NearestLandmark: updatedAddress.NearestLandmark,
		ShippingMobile:  updatedAddress.ShippingMobile,
		IsDefault:       updatedAddress.IsDefault,
		IsActive:        updatedAddress.IsActive,
		CreatedOn:       updatedAddress.CreatedOn,
		LastModifiedOn:  updatedAddress.LastModifiedOn,
	}

	s.txnManager.CommitTxn(tx)
	return addressDTO, nil
}

func (s *userService) DeleteUserAddress(ctx context.Context, userID string, addressID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("user not found")
	}

	addrID, err := strconv.ParseInt(addressID, 10, 64)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("invalid address ID format")
	}

	address, err := s.addressRepo.FindByID(ctx, tx, int(addrID))
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return err
	}

	if address == nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("address not found")
	}

	if address.UserID != userID {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("address does not belong to user")
	}

	wasDefault := address.IsDefault

	if err := s.addressRepo.Delete(ctx, tx, int(addrID)); err != nil {
		s.txnManager.AbortTxn(tx)
		return err
	}

	if wasDefault {
		remainingAddresses, err := s.addressRepo.FindWithConditions(ctx, tx, map[string]interface{}{
			"user_id":    userID,
			"is_deleted": false,
		})
		if err != nil && err != gorm.ErrRecordNotFound {
			s.txnManager.AbortTxn(tx)
			return err
		}

		if len(remainingAddresses) > 0 {
			var mostRecentAddress *entities.Address
			for _, addr := range remainingAddresses {
				if mostRecentAddress == nil || addr.CreatedOn.After(mostRecentAddress.CreatedOn) {
					addrCopy := addr
					mostRecentAddress = &addrCopy
				}
			}

			if mostRecentAddress != nil {
				if err := s.addressRepo.UpdateFields(ctx, tx, mostRecentAddress.ID, map[string]interface{}{
					"is_default": true,
				}); err != nil {
					s.txnManager.AbortTxn(tx)
					return err
				}
			}
		}
	}

	s.txnManager.CommitTxn(tx)
	return nil
}
