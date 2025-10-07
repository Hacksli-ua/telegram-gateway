package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Telegram API
	TelegramAPIID         int
	TelegramAPIHash       string
	TelegramPublicKeyFile string
	TelegramDCID          int
	TelegramServerHost    string
	TelegramServerPort    string

	// Server
	ServerPort string
	ServerHost string

	// Session
	SessionTimeout  time.Duration
	CleanupInterval time.Duration

	// Long Polling
	PollTimeout time.Duration

	// Database (для production)
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Redis (для production)
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// Security
	CORSAllowedOrigins string
	EnableHTTPS        bool
	SSLCertPath        string
	SSLKeyPath         string
}

var AppConfig *Config

// LoadConfig завантажує конфігурацію з .env файлу
func LoadConfig() (*Config, error) {
	// Завантаження .env файлу (ігнорується помилка, якщо файлу немає)
	_ = godotenv.Load()

	apiID, err := strconv.Atoi(getEnv("TELEGRAM_API_ID", ""))
	if err != nil || apiID == 0 {
		log.Fatal("TELEGRAM_API_ID must be set and valid")
	}

	apiHash := getEnv("TELEGRAM_API_HASH", "")
	if apiHash == "" {
		log.Fatal("TELEGRAM_API_HASH must be set")
	}

	dcID, _ := strconv.Atoi(getEnv("TELEGRAM_DC_ID", "2"))

	config := &Config{
		TelegramAPIID:         apiID,
		TelegramAPIHash:       apiHash,
		TelegramPublicKeyFile: getEnv("TELEGRAM_PUBLIC_KEY_FILE", "telegram_public_key.pem"),
		TelegramDCID:          dcID,
		TelegramServerHost:    getEnv("TELEGRAM_SERVER_HOST", "149.154.167.50"),
		TelegramServerPort:    getEnv("TELEGRAM_SERVER_PORT", "443"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "localhost"),

		SessionTimeout:  parseDuration(getEnv("SESSION_TIMEOUT", "30m")),
		CleanupInterval: parseDuration(getEnv("CLEANUP_INTERVAL", "5m")),
		PollTimeout:     parseDuration(getEnv("POLL_TIMEOUT", "50s")),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "telegram_gateway"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		EnableHTTPS:        getEnv("ENABLE_HTTPS", "false") == "true",
		SSLCertPath:        getEnv("SSL_CERT_PATH", ""),
		SSLKeyPath:         getEnv("SSL_KEY_PATH", ""),
	}

	AppConfig = config
	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Invalid duration %s, using default", s)
		return 30 * time.Minute
	}
	return d
}
