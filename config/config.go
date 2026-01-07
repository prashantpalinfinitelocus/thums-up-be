package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	AppPort        string
	AllowedOrigins string
	SwaggerHost    string
	DbConfig       DatabaseConfig
	InfobipConfig  InfobipConfig
	JwtConfig      JwtConfig
	FirebaseConfig FirebaseConfig
	GcsConfig      GcsConfig
	PubSubConfig   PubSubConfig
	XAPIKey        string
}

var (
	config     *Config
	configOnce sync.Once
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type InfobipConfig struct {
	BaseURL  string
	APIKey   string
	WANumber string
}

type JwtConfig struct {
	SecretKey          string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
}

type FirebaseConfig struct {
	ServiceKeyPath string
}

type GcsConfig struct {
	BucketName string
	ProjectID  string
	GcpUrl     string
}

type PubSubConfig struct {
	ProjectID      string
	SubscriptionID string
	TopicID        string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		AppPort:        getEnv("APP_PORT", "8080"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		SwaggerHost:    getEnv("SWAGGER_HOST", "localhost:8080"),

		DbConfig: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     parseEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", ""),
			SSLMode:  getSSLMode(),
		},

		InfobipConfig: InfobipConfig{
			BaseURL:  getEnv("INFOBIP_BASE_URL", ""),
			APIKey:   getEnv("INFOBIP_API_KEY", ""),
			WANumber: getEnv("INFOBIP_WA_NUMBER", ""),
		},

		JwtConfig: JwtConfig{
			SecretKey:          getEnv("JWT_SECRET_KEY", ""),
			AccessTokenExpiry:  parseEnvInt("JWT_ACCESS_TOKEN_EXPIRY", 3600),
			RefreshTokenExpiry: parseEnvInt("JWT_REFRESH_TOKEN_EXPIRY", 2592000),
		},

		FirebaseConfig: FirebaseConfig{
			ServiceKeyPath: getEnv("FIREBASE_SERVICE_KEY_PATH", ""),
		},

		GcsConfig: GcsConfig{
			BucketName: getEnv("GCP_BUCKET_NAME", ""),
			ProjectID:  getEnv("GCP_PROJECT_ID", ""),
			GcpUrl:     getEnv("GCP_URL", ""),
		},

		PubSubConfig: PubSubConfig{
			ProjectID:      getEnv("GOOGLE_PUBSUB_PROJECT_ID", ""),
			SubscriptionID: getEnv("GOOGLE_PUBSUB_SUBSCRIPTION_ID", ""),
			TopicID:        getEnv("GOOGLE_PUBSUB_TOPIC_ID", ""),
		},

		XAPIKey: getEnv("X_API_KEY", ""),
	}, nil
}

func parseEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getSSLMode converts DATABASE_SSL boolean to PostgreSQL SSL mode string
// Valid PostgreSQL SSL modes: disable, allow, prefer, require, verify-ca, verify-full
func getSSLMode() string {
	// First check if DB_SSL_MODE is set directly (preferred)
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		return sslMode
	}

	// Otherwise, convert DATABASE_SSL boolean to SSL mode
	databaseSSL := os.Getenv("DATABASE_SSL")
	databaseSSLRejectUnauthorized := os.Getenv("DATABASE_SSL_REJECT_UNAUTHORIZED")

	// Parse boolean values (accept "true", "1", "yes", "on" as true)
	isSSLEnabled := databaseSSL == "true" || databaseSSL == "1" || databaseSSL == "yes" || databaseSSL == "on"
	isRejectUnauthorized := databaseSSLRejectUnauthorized == "true" || databaseSSLRejectUnauthorized == "1" || databaseSSLRejectUnauthorized == "yes" || databaseSSLRejectUnauthorized == "on"

	if isSSLEnabled {
		if isRejectUnauthorized {
			return "verify-full"
		}
		return "require"
	}

	return "disable"
}

func GetConfig() *Config {
	configOnce.Do(func() {
		var err error
		config, err = loadConfig()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	})
	return config
}
