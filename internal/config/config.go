package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Server
	Port string

	// Request Execution
	RequestTimeout  time.Duration
	MaxRequestSize  int64
	MaxResponseSize int64
	MaxHeaderCount  int
	MaxRedirects    int

	// Rate Limiting
	RateLimitRPS   int
	RateLimitBurst int

	// SSRF Protection
	AllowLocalhost  bool
	AllowPrivateIPs bool
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "dev"),
		DBPassword: getEnv("DB_PASSWORD", "localdb"),
		DBName:     getEnv("DB_NAME", "karikatokdevdb"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		Port:       getEnv("PORT", "8080"),
	}

	// Parse durations and integers
	var err error
	cfg.RequestTimeout, err = time.ParseDuration(getEnv("REQUEST_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid REQUEST_TIMEOUT: %w", err)
	}

	cfg.MaxRequestSize, err = strconv.ParseInt(getEnv("MAX_REQUEST_SIZE", "10485760"), 10, 64) // 10MB
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_REQUEST_SIZE: %w", err)
	}

	cfg.MaxResponseSize, err = strconv.ParseInt(getEnv("MAX_RESPONSE_SIZE", "52428800"), 10, 64) // 50MB
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_RESPONSE_SIZE: %w", err)
	}

	cfg.MaxHeaderCount, err = strconv.Atoi(getEnv("MAX_HEADER_COUNT", "50"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_HEADER_COUNT: %w", err)
	}

	cfg.MaxRedirects, err = strconv.Atoi(getEnv("MAX_REDIRECTS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_REDIRECTS: %w", err)
	}

	cfg.RateLimitRPS, err = strconv.Atoi(getEnv("RATE_LIMIT_RPS", "1000"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}

	cfg.RateLimitBurst, err = strconv.Atoi(getEnv("RATE_LIMIT_BURST", "2000"))
	cfg.AllowLocalhost = getEnv("ALLOW_LOCALHOST", "true") == "true"
	cfg.AllowPrivateIPs = getEnv("ALLOW_PRIVATE_IPS", "true") == "true"

	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
