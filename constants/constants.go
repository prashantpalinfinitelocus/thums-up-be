package constants

const (
	OTP_LENGTH              = 6
	OTP_EXPIRY_MINUTES      = 5
	MAX_OTP_ATTEMPTS        = 5
	OTP_COOLDOWN_MINUTES    = 2
	MAX_VERIFICATION_TRIES  = 3
	VERIFICATION_LOCKOUT    = 15
	
	NOTIFICATION_CATEGORY   = "thums_up_notification"
	
	ROLE_USER               = "user"
	ROLE_ADMIN              = "admin"
	
	PLATFORM_ANDROID        = 1
	PLATFORM_IOS            = 2
	PLATFORM_WEB            = 3
	
	STATUS_ACTIVE           = "active"
	STATUS_INACTIVE         = "inactive"
	STATUS_PENDING          = "pending"
	
	NOTIFY_ME_TOPIC         = "notify-me"
	
	DEFAULT_PAGE_SIZE       = 20
	MAX_PAGE_SIZE           = 100
)

var (
	AllowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg", "image/webp"}
	MaxFileSize      = int64(5 * 1024 * 1024)
)

