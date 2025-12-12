package constants

import "time"

const (
	OTP_LENGTH             = 6
	OTP_EXPIRY_MINUTES     = 5
	MAX_OTP_ATTEMPTS       = 5
	OTP_COOLDOWN_MINUTES   = 2
	MAX_VERIFICATION_TRIES = 3
	VERIFICATION_LOCKOUT   = 15

	NOTIFICATION_CATEGORY = "thums_up_notification"

	ROLE_USER  = "user"
	ROLE_ADMIN = "admin"

	PLATFORM_ANDROID = 1
	PLATFORM_IOS     = 2
	PLATFORM_WEB     = 3

	STATUS_ACTIVE   = "active"
	STATUS_INACTIVE = "inactive"
	STATUS_PENDING  = "pending"

	NOTIFY_ME_TOPIC = "notify-me"

	DEFAULT_PAGE_SIZE = 20
	MAX_PAGE_SIZE     = 100

	// Background task timeouts
	BACKGROUND_TASK_TIMEOUT = 5 * time.Second
	PUBSUB_PUBLISH_TIMEOUT  = 30 * time.Second

	// Database settings
	DEFAULT_DB_MAX_IDLE_CONNS     = 10
	DEFAULT_DB_MAX_OPEN_CONNS     = 100
	DEFAULT_DB_CONN_MAX_IDLE_TIME = 5 * time.Minute
	DEFAULT_DB_CONN_MAX_LIFETIME  = 30 * time.Minute

	// Worker pool settings
	WORKER_POOL_SIZE    = 10
	TASK_QUEUE_SIZE     = 1000
	PUBSUB_ACK_DEADLINE = 20 * time.Second

	// Batch settings
	NOTIFY_ME_BATCH_SIZE = 1000
	FIREBASE_BATCH_SIZE  = 500

	// Timeouts
	HTTP_CLIENT_TIMEOUT       = 30 * time.Second
	GRACEFUL_SHUTDOWN_TIMEOUT = 10 * time.Second
	MESSAGE_HANDLER_TIMEOUT   = 30 * time.Second
)

var (
	AllowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg", "image/webp"}
	MaxFileSize      = int64(5 * 1024 * 1024)
)
