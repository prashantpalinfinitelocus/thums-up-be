package errors

const (
	ErrValidationFailed           = "Validation failed"
	ErrInvalidRequestBody         = "Invalid request body"
	ErrInvalidUserIDFormat        = "Invalid user ID format"
	ErrInvalidAddressIDFormat     = "Invalid address ID format"
	ErrInvalidWeekNumber          = "Invalid week number"
	ErrInvalidUserContext         = "Invalid user context"
	
	ErrUserNotAuthenticated       = "User not authenticated"
	ErrUserAlreadyExists          = "User with this phone number already exists"
	ErrEmailAlreadyInUse          = "Email already in use"
	ErrInvalidUserID              = "Invalid user ID in token"
	
	ErrAuthHeaderRequired         = "Authorization header required"
	ErrInvalidAuthHeaderFormat    = "Invalid authorization header format"
	ErrInvalidOrExpiredToken      = "Invalid or expired token"
	ErrInvalidTokenClaims         = "Invalid token claims"
	
	ErrOTPSendFailed              = "Failed to send OTP"
	ErrOTPVerifyFailed            = "Failed to verify OTP"
	ErrOTPTooManyRequests         = "Too many OTP requests. Please try again later"
	ErrOTPInvalidOrExpired        = "Invalid or expired OTP"
	ErrOTPSMSFailed               = "Failed to send OTP via SMS"
	
	ErrTokenGenerationFailed      = "Failed to generate access token"
	ErrTokenRefreshFailed         = "Failed to refresh token"
	ErrRefreshTokenInvalid        = "Invalid refresh token"
	ErrRefreshTokenRevoked        = "Refresh token has been revoked"
	ErrRefreshTokenExpired        = "Refresh token has expired"
	
	ErrSubscriptionFailed         = "Failed to subscribe"
	ErrSubscriptionNotFound       = "Subscription not found"
	ErrSubscriptionCheck          = "Failed to check subscription"
	ErrSubscriptionAlreadyExists  = "Phone number is already subscribed to notifications"
	
	ErrQuestionNotFound           = "Question not found"
	ErrQuestionSubmitFailed       = "Failed to submit question"
	ErrQuestionFetchFailed        = "Failed to get active questions"
	ErrQuestionVerifyFailed       = "Failed to verify question"
	ErrQuestionByLanguageFailed   = "Failed to get questions by language"
	
	ErrAnswerSubmitFailed         = "Failed to submit answer"
	ErrAnswerAlreadySubmitted     = "You have already submitted an answer for this question"
	ErrSubmissionCheckFailed      = "Failed to check submission"
	ErrSubmissionsFetchFailed     = "Failed to get user submissions"
	ErrCurrentWeekFailed          = "Failed to get current week"
	
	ErrWinnerSelectFailed         = "Failed to select winners"
	ErrWinnerSaveFailed           = "Failed to save winners"
	ErrWinnerFetchFailed          = "Failed to get winners"
	ErrWinnerNoEligibleEntries    = "No eligible entries found for winner selection"
	ErrWinnerGetExistingFailed    = "Failed to get existing winners"
	ErrWinnerGetRandomFailed      = "Failed to select random entries"
	
	ErrProfileFetchFailed         = "Failed to fetch user profile"
	ErrProfileUpdateFailed        = "Failed to update profile"
	ErrProfileCreateFailed        = "Failed to create user"
	
	ErrAddressFetchFailed         = "Failed to fetch addresses"
	ErrAddressAddFailed           = "Failed to add address"
	ErrAddressUpdateFailed        = "Failed to update address"
	ErrAddressDeleteFailed        = "Failed to delete address"
	ErrAddressNotBelongToUser     = "Address does not belong to user"
	ErrAddressIDRequired          = "Address ID is required"
	ErrAddressCreationFailed      = "Failed to create address"
	
	ErrStateFindFailed            = "Failed to find state"
	ErrCityFindFailed             = "Failed to find city"
	ErrPinCodeFindFailed          = "Failed to find pincode"
	ErrPinCodeNotFound            = "Pincode not found in city"
	ErrPinCodeNotDeliverable      = "Pincode is not deliverable in city"
	ErrPinCodeCheckFailed         = "Failed to check pincode deliverability"
	
	ErrLocationDetailsFetchFailed = "Failed to get location details"
	ErrLocationDetailsMissing     = "Missing location details for address"
	
	ErrPhoneNumberRequired        = "Phone number is required"
	ErrPhoneNumberCheck           = "Failed to check phone number"
	ErrEmailCheck                 = "Failed to check email"
	ErrEmailUniquenessCheck       = "Failed to check email uniqueness"
	
	ErrDatabaseOperation          = "Database operation failed"
	ErrTransactionFailed          = "Transaction failed"
	ErrRecordNotFound             = "Record not found"
	
	ErrNotificationPublishFailed  = "Failed to publish notify me message"
	ErrUnnotifiedFetchFailed      = "Failed to get unnotified subscriptions"
	ErrMarkNotifiedFailed         = "Failed to mark as notified"
	
	ErrInternalServer             = "Internal server error"
	ErrServiceUnavailable         = "Service unavailable"
)

