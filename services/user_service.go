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
	GetUser(ctx context.Context, userID string) (*entities.User, *string, error)
	UpdateUser(ctx context.Context, userID string, req dtos.UpdateProfileRequestDTO) (*entities.User, error)
	GetUserAddresses(ctx context.Context, userID string) ([]dtos.AddressResponseDTO, error)
	AddUserAddress(ctx context.Context, userID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error)
	UpdateUserAddress(ctx context.Context, userID string, addressID string, req dtos.AddressRequestDTO) (*dtos.AddressResponseDTO, error)
	DeleteUserAddress(ctx context.Context, userID string, addressID string) error
	GetQuestions(ctx context.Context, userID string, languageID int) ([]dtos.QuestionResponseDTO, error)
	GetQuestionIDByText(ctx context.Context, questionText string, languageID int) (int, error)
	GetQuestionByID(ctx context.Context, questionID int, languageID int) (*dtos.QuestionResponseDTO, error)
	AnswerQuestions(ctx context.Context, userID string, answers []dtos.AnswerQuestionsRequestDTO) error
}

type userService struct {
	txnManager                 *utils.TransactionManager
	userRepo                   repository.UserRepository
	addressRepo                repository.GenericRepository[entities.Address]
	stateRepo                  repository.StateRepository
	cityRepo                   repository.CityRepository
	pinCodeRepo                repository.PinCodeRepository
	avatarRepo                 repository.GenericRepository[entities.Avatar]
	gcsService                 utils.GCSService
	questionAnswerRepo         repository.UserQuestionAnswerRepository
	questionMasterRepo         repository.QuestionRepository
	questionMasterLanguageRepo repository.QuestionMasterLanguageRepository
	optionMasterRepo           repository.OptionMasterRepository
	optionMasterLanguageRepo   repository.OptionMasterLanguageRepository
}

func NewUserService(
	txnManager *utils.TransactionManager,
	userRepo repository.UserRepository,
	addressRepo repository.GenericRepository[entities.Address],
	stateRepo repository.StateRepository,
	cityRepo repository.CityRepository,
	pinCodeRepo repository.PinCodeRepository,
	avatarRepo repository.GenericRepository[entities.Avatar],
	gcsService utils.GCSService,
	questionAnswerRepo repository.UserQuestionAnswerRepository,
	questionMasterRepo repository.QuestionRepository,
	questionMasterLanguageRepo repository.QuestionMasterLanguageRepository,
	optionMasterRepo repository.OptionMasterRepository,
	optionMasterLanguageRepo repository.OptionMasterLanguageRepository,
) UserService {
	return &userService{
		txnManager:                 txnManager,
		userRepo:                   userRepo,
		addressRepo:                addressRepo,
		stateRepo:                  stateRepo,
		cityRepo:                   cityRepo,
		pinCodeRepo:                pinCodeRepo,
		avatarRepo:                 avatarRepo,
		gcsService:                 gcsService,
		questionAnswerRepo:         questionAnswerRepo,
		questionMasterRepo:         questionMasterRepo,
		questionMasterLanguageRepo: questionMasterLanguageRepo,
		optionMasterRepo:           optionMasterRepo,
		optionMasterLanguageRepo:   optionMasterLanguageRepo,
	}
}

func (s *userService) GetUser(ctx context.Context, userID string) (*entities.User, *string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return nil, nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, nil, err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return nil, nil, fmt.Errorf("user not found")
	}

	var avatarImageURL *string
	if user.AvatarID != nil {
		avatar, err := s.avatarRepo.FindByID(ctx, tx, *user.AvatarID)
		if err == nil && avatar != nil && !avatar.IsDeleted && avatar.IsActive {
			// Reconstruct full path: avatars/{userID}/{filename}
			fullPath := fmt.Sprintf("avatars/%s/%s", avatar.CreatedBy, avatar.ImageKey)
			imageURL := s.gcsService.GetPublicURL(fullPath)
			avatarImageURL = &imageURL
		}
	}

	s.txnManager.CommitTxn(tx)
	return user, avatarImageURL, nil
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

	if req.AvatarID != nil {
		avatar, err := s.avatarRepo.FindByID(ctx, tx, *req.AvatarID)
		if err != nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to fetch avatar: %v", err)
		}
		if avatar == nil || avatar.IsDeleted || !avatar.IsActive {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("avatar not found or not available")
		}
		updateFields["avatar_id"] = *req.AvatarID
	}

	if req.SharingPlatform != nil {
		updateFields["sharing_platform"] = *req.SharingPlatform
	}

	if req.PlatformUserName != nil {
		updateFields["platform_user_name"] = *req.PlatformUserName
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

func (s *userService) GetQuestions(ctx context.Context, userID string, languageID int) ([]dtos.QuestionResponseDTO, error) {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	// Step 1: Fetch all active questions from question_master
	questions, err := s.questionMasterRepo.FindActive(ctx, tx)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get questions: %v", err)
	}

	// Step 2: Get question details from question_master_language table based on question ID and language ID
	questionLanguageMap := make(map[int]string)
	for _, question := range questions {
		questionLanguage, err := s.questionMasterLanguageRepo.FindByQuestionMasterIDAndLanguageID(ctx, tx, question.ID, languageID)
		if err != nil {
			s.txnManager.AbortTxn(tx)
			return nil, fmt.Errorf("failed to get question language text for question %d: %v", question.ID, err)
		}
		if questionLanguage != nil {
			questionLanguageMap[question.ID] = questionLanguage.QuestionText
		} else {
			questionLanguageMap[question.ID] = question.QuestionText // fallback to original text
		}
	}

	// Step 3: Fetch options from option_master table using question IDs
	questionIDs := make([]int, 0)
	for _, q := range questions {
		questionIDs = append(questionIDs, q.ID)
	}
	options, err := s.optionMasterRepo.FindByQuestionIDs(ctx, tx, questionIDs)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get options: %v", err)
	}

	// Step 4: Fetch option details from option_master_language table based on option ID and language ID
	optionLanguageMap := make(map[int]string)
	optionIDs := make([]int, 0)
	for _, opt := range options {
		optionIDs = append(optionIDs, opt.ID)
	}
	optionLanguages, err := s.optionMasterLanguageRepo.FindByOptionMasterIDsAndLanguageID(ctx, tx, optionIDs, languageID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get option language texts: %v", err)
	}
	for _, optLang := range optionLanguages {
		optionLanguageMap[optLang.OptionMasterID] = optLang.OptionText
	}

	// Step 5: Fetch all user questions from user_question_answer table
	userAnswers, err := s.questionAnswerRepo.FindByUserID(ctx, tx, userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get user answers: %v", err)
	}

	// Create a map of user answers for quick lookup
	userAnswerMap := make(map[int]entities.UserQuestionAnswer)
	for _, ans := range userAnswers {
		userAnswerMap[ans.QuestionMasterID] = ans
	}

	// Step 6: Fill options accordingly and construct response
	response := make([]dtos.QuestionResponseDTO, 0)
	for _, question := range questions {
		// Get question text (language-specific or fallback)
		questionText := questionLanguageMap[question.ID]

		// Get options for this question
		var questionOptions []dtos.OptionDTO
		for _, opt := range options {
			if opt.QuestionMasterID == question.ID {
				// Get option text (language-specific or fallback)
				optionText := opt.OptionText
				if langText, exists := optionLanguageMap[opt.ID]; exists {
					optionText = langText
				}

				questionOptions = append(questionOptions, dtos.OptionDTO{
					ID:           opt.ID,
					OptionText:   optionText,
					DisplayOrder: opt.DisplayOrder,
				})
			}
		}

		// Check if user has answered this question
		var selectedOption *int
		if userAnswer, exists := userAnswerMap[question.ID]; exists {
			selectedOption = &userAnswer.OptionID
		}

		response = append(response, dtos.QuestionResponseDTO{
			ID:             question.ID,
			QuestionText:   questionText,
			LanguageID:     languageID,
			Options:        questionOptions,
			SelectedOption: selectedOption,
		})
	}

	s.txnManager.CommitTxn(tx)
	return response, nil
}

func (s *userService) GetQuestionIDByText(ctx context.Context, questionText string, languageID int) (int, error) {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return 0, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	questionLanguage, err := s.questionMasterLanguageRepo.FindByQuestionTextAndLanguageID(ctx, tx, questionText, languageID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return 0, fmt.Errorf("failed to search question in language table: %v", err)
	}
	if questionLanguage != nil {
		s.txnManager.CommitTxn(tx)
		return questionLanguage.QuestionMasterID, nil
	}

	question, err := s.questionMasterRepo.FindByQuestionTextAndLanguageID(ctx, tx, questionText, languageID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return 0, fmt.Errorf("failed to search question in master table: %v", err)
	}
	if question != nil {
		s.txnManager.CommitTxn(tx)
		return question.ID, nil
	}

	s.txnManager.AbortTxn(tx)
	return 0, fmt.Errorf("question not found")
}

func (s *userService) GetQuestionByID(ctx context.Context, questionID int, languageID int) (*dtos.QuestionResponseDTO, error) {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	question, err := s.questionMasterRepo.FindByIDTx(ctx, tx, questionID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get question: %v", err)
	}
	if question == nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("question not found")
	}

	questionLanguage, err := s.questionMasterLanguageRepo.FindByQuestionMasterIDAndLanguageID(ctx, tx, question.ID, languageID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get question language text: %v", err)
	}

	questionText := question.QuestionText
	if questionLanguage != nil {
		questionText = questionLanguage.QuestionText
	}

	options, err := s.optionMasterRepo.FindByQuestionID(ctx, tx, questionID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get options: %v", err)
	}

	optionIDs := make([]int, 0)
	for _, opt := range options {
		optionIDs = append(optionIDs, opt.ID)
	}

	optionLanguages, err := s.optionMasterLanguageRepo.FindByOptionMasterIDsAndLanguageID(ctx, tx, optionIDs, languageID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return nil, fmt.Errorf("failed to get option language texts: %v", err)
	}

	optionLanguageMap := make(map[int]string)
	for _, optLang := range optionLanguages {
		optionLanguageMap[optLang.OptionMasterID] = optLang.OptionText
	}

	var questionOptions []dtos.OptionDTO
	for _, opt := range options {
		optionText := opt.OptionText
		if langText, exists := optionLanguageMap[opt.ID]; exists {
			optionText = langText
		}

		questionOptions = append(questionOptions, dtos.OptionDTO{
			ID:           opt.ID,
			OptionText:   optionText,
			DisplayOrder: opt.DisplayOrder,
		})
	}

	response := &dtos.QuestionResponseDTO{
		ID:           question.ID,
		QuestionText: questionText,
		LanguageID:   languageID,
		Options:      questionOptions,
	}

	s.txnManager.CommitTxn(tx)
	return response, nil
}

func (s *userService) AnswerQuestions(ctx context.Context, userID string, answers []dtos.AnswerQuestionsRequestDTO) error {
	tx, err := s.txnManager.StartTxn()
	if err != nil {
		return err
	}
	defer s.txnManager.RollbackOnPanic(tx)

	// First verify if user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindById(ctx, tx, userUUID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return err
	}

	if user == nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("user not found")
	}

	// Check existing answers before creating new ones
	existingAnswers, err := s.questionAnswerRepo.FindByUserID(ctx, tx, userID)
	if err != nil {
		s.txnManager.AbortTxn(tx)
		return fmt.Errorf("failed to get existing answers: %v", err)
	}

	// Create a map of existing answers for quick lookup
	existingAnswerMap := make(map[int]entities.UserQuestionAnswer)
	for _, ans := range existingAnswers {
		existingAnswerMap[ans.QuestionMasterID] = ans
	}

	// Process each answer
	now := time.Now()
	newAnswers := make([]entities.UserQuestionAnswer, 0)
	updatedAnswers := make([]entities.UserQuestionAnswer, 0)

	for _, answer := range answers {
		if existingAnswer, exists := existingAnswerMap[answer.QuestionID]; exists {
			// Question already answered, just update the option ID
			existingAnswer.OptionID = answer.AnswerID
			existingAnswer.LastModifiedBy = &userID
			existingAnswer.LastModifiedOn = &now
			updatedAnswers = append(updatedAnswers, existingAnswer)
		} else {
			// New answer, create it
			newAnswer := entities.UserQuestionAnswer{
				UserID:           userID,
				QuestionMasterID: answer.QuestionID,
				OptionID:         answer.AnswerID,
				SelectedAnswer:   true,
				IsActive:         true,
				IsDeleted:        false,
				CreatedBy:        userID,
				CreatedOn:        now,
			}
			newAnswers = append(newAnswers, newAnswer)
		}
	}

	// Update existing answers
	for _, answer := range updatedAnswers {
		if err := s.questionAnswerRepo.Update(ctx, tx, &answer); err != nil {
			s.txnManager.AbortTxn(tx)
			return fmt.Errorf("failed to update existing answer: %v", err)
		}
	}

	// Save new answers
	if len(newAnswers) > 0 {
		if err := s.questionAnswerRepo.CreateMany(ctx, tx, newAnswers); err != nil {
			s.txnManager.AbortTxn(tx)
			return fmt.Errorf("failed to save new answers: %v", err)
		}
	}

	s.txnManager.CommitTxn(tx)
	return nil
}
